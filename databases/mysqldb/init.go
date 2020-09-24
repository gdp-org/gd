/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package mysqldb

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

const default_char_set = "utf8mb4"

func (c *MysqlClient) initObjForMysqldb(dbConfPath string) error {
	//init mysql for mipush
	dbConfRealPath := dbConfPath
	if dbConfRealPath == "" {
		return errors.New("dbConf not set in g_cfg")
	}
	if !strings.HasSuffix(dbConfRealPath, ".ini") {
		return errors.New("dbConf not an ini file")
	}
	dbConf, err := ini.Load(dbConfRealPath)
	if err != nil {
		return err
	}
	err = c.initDbs(dbConf)
	if err != nil {
		return err
	}
	return nil
}

func (c *MysqlClient) initDbs(f *ini.File) error {
	m := f.Section("Mysql")
	s := f.Section("MysqlSlave")
	var err error
	master_ip := m.Key("master_ip").String()
	master_port := m.Key("master_port").String()
	user_write := m.Key("user_write").String()
	pass_write := m.Key("pass_write").String()

	user_read := m.Key("user_read").String()
	pass_read := m.Key("pass_read").String()

	db := m.Key("db").String()
	masterProxy, _ := m.Key("master_is_proxy").Bool()
	slave_ip := s.Key("slave_ip").String()
	slave_port := s.Key("slave_port").String()
	slaveProxy, _ := s.Key("slave_is_proxy").Bool()

	timeout := f.Section("").Key("timeout").String()
	if timeout == "" {
		timeout = "5s"
	} else if !strings.HasSuffix(timeout, "s") {
		timeout += "s"
	}
	connTimeout := f.Section("").Key("connTimeout").String()
	if connTimeout == "" {
		connTimeout = "1s"
	} else if !strings.HasSuffix(timeout, "s") {
		connTimeout += "s"
	}

	maxOpen, err := f.Section("").Key("maxopen").Int()
	if err != nil {
		maxOpen = 100
	}
	maxIdle, err := f.Section("").Key("maxidle").Int()
	if err != nil {
		maxIdle = 1
	}

	enableSqlSafeUpdates, err := f.Section("").Key("enable_sql_safe_updates").Bool()
	if err != nil {
		enableSqlSafeUpdates = false
	}

	masterIps := strings.Split(master_ip, ",")
	connMasters := []string{}
	for _, masterIp := range masterIps {
		if masterIp == "" {
			continue
		}

		connMaster := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s", user_write, pass_write, masterIp, master_port, db, connTimeout, timeout, timeout)
		if enableSqlSafeUpdates {
			connMaster = connMaster + "&sql_safe_updates=1"
		}

		connMasters = append(connMasters, connMaster)
	}

	slaveIps := strings.Split(slave_ip, ",")
	connSlaves := []string{}
	for _, slaveIp := range slaveIps {
		if slaveIp == "" {
			continue
		}
		connSlaves = append(connSlaves, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s", user_read, pass_read, slaveIp, slave_port, db, connTimeout, timeout, timeout))
	}

	ctxSuffix := f.Section("").Key("ctxSuffix").String()
	to, _ := time.ParseDuration(timeout)
	return c.initMainDbsMaxOpen(connMasters, connSlaves, maxOpen, maxIdle, ctxSuffix, to, masterProxy, slaveProxy)
}

type CommonDbConf struct {
	DbName      string
	ConnTime    string // connect timeout
	ReadTime    string // read timeout
	WriteTime   string // write timeout
	MaxOpen     int    // connect pool
	MaxIdle     int    // max idle connect
	MaxLifeTime int64  // connect life time
	CtxSuffix   string
	Master      *DbConnectConf
	Slave       *DbConnectConf
}

type DbConnectConf struct {
	Addrs                []string
	User                 string
	Pass                 string
	CharSet              string // default utf8mb4
	ClientFoundRows      bool   //对于update操作,若更改的字段值跟原来值相同,当clientFoundRows为false时,sql执行结果会返回0;当clientFoundRows为true,sql执行结果返回1
	IsProxy              bool
	EnableSqlSafeUpdates bool   // (safe update mode)，该模式不允许没有带WHERE条件的更新语句
}

func (c *MysqlClient) initDbsWithCommonConf(dbConf *CommonDbConf) error {
	var err error
	if dbConf == nil {
		return errors.New("dbConf is nil")
	}
	if dbConf.Master == nil || len(dbConf.Master.Addrs) == 0 {
		return errors.New("masterAddr is nil")
	}
	if dbConf.DbName == "" {
		return errors.New("dbName is nil")
	}
	connTimeout := dbConf.ConnTime
	if connTimeout == "" {
		connTimeout = "200ms"
	}
	readTimeout := dbConf.ReadTime
	if readTimeout == "" {
		readTimeout = "500ms"
	}
	writeTimeout := dbConf.WriteTime
	if writeTimeout == "" {
		writeTimeout = "500ms"
	}
	maxOpen := dbConf.MaxOpen
	if maxOpen <= 0 {
		maxOpen = 100
	}
	maxIdle := dbConf.MaxIdle
	if maxIdle <= 0 {
		maxIdle = 1
	}
	connMasters, err := c.getReadWriteConnectString(dbConf.Master, connTimeout, readTimeout, writeTimeout, dbConf.DbName)
	if err != nil {
		return err
	}
	if len(connMasters) == 0 {
		return errors.New("no valid master ip found")
	}
	connSlave, err := c.getReadWriteConnectString(dbConf.Slave, connTimeout, readTimeout, writeTimeout, dbConf.DbName)
	if err != nil {
		return err
	}
	slaveisproxy := false
	if dbConf.Slave != nil {
		slaveisproxy = dbConf.Slave.IsProxy
	}

	to, err := time.ParseDuration(readTimeout)
	if err != nil {
		return fmt.Errorf("init mysqldb invalid duration %v", readTimeout)
	}
	return c.initMainDbsMaxOpen(connMasters, connSlave, maxOpen, maxIdle, dbConf.CtxSuffix, to, dbConf.Master.IsProxy, slaveisproxy)
}

func (c *MysqlClient) getConnectString(conf *DbConnectConf, connTimeout, optTimeout int64, dbname string) ([]string, error) {
	if conf == nil || len(conf.Addrs) == 0 {
		return nil, nil
	}

	if conf.CharSet == "" {
		conf.CharSet = default_char_set
	}
	constrs := make([]string, 0, len(conf.Addrs))
	for _, host := range conf.Addrs {
		if host != "" {
			var constr string
			if conf.ClientFoundRows {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=%s&clientFoundRows=true",
					conf.User, conf.Pass, host, dbname, connTimeout, optTimeout, optTimeout, conf.CharSet)
			} else {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=%s",
					conf.User, conf.Pass, host, dbname, connTimeout, optTimeout, optTimeout, conf.CharSet)
			}

			if conf.EnableSqlSafeUpdates {
				constr = constr + "&sql_safe_updates=1"
			}

			constrs = append(constrs, constr)
		}
	}
	return constrs, nil
}

func (c *MysqlClient) getReadWriteConnectString(conf *DbConnectConf, connTimeout, readTimeout, writeTimeout string, dbname string) ([]string, error) {
	if conf == nil || len(conf.Addrs) == 0 {
		return nil, nil
	}

	if conf.CharSet == "" {
		conf.CharSet = default_char_set
	}
	constrs := make([]string, 0, len(conf.Addrs))
	for _, host := range conf.Addrs {
		if host != "" {
			var constr string
			if conf.ClientFoundRows {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s&clientFoundRows=true",
					conf.User, conf.Pass, host, dbname, connTimeout, readTimeout, writeTimeout, conf.CharSet)
			} else {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s",
					conf.User, conf.Pass, host, dbname, connTimeout, readTimeout, writeTimeout, conf.CharSet)
			}

			if conf.EnableSqlSafeUpdates {
				constr = constr + "&sql_safe_updates=1"
			}

			constrs = append(constrs, constr)
		}
	}

	//log.Debug("connstr %v", constrs)
	return constrs, nil
}

