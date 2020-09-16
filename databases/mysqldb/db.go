/**
 * Copyright 2019 mysqldb Author. All rights reserved.
 * Author: Chuck1024
 */

package mysqldb

import (
	"database/sql"
	"errors"
	"fmt"
	log "github.com/chuck1024/gd/dlog"
	"gopkg.in/ini.v1"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
)

const (
	defaultDbConf = "db.ini"
)

type MysqlClient struct {
	DbConfig   *CommonDbConf
	DbConf     *ini.File
	DbConfPath string

	dbWrite []*DbWrap
	dbRead  []*DbWrap

	startOnce sync.Once
	closeOnce sync.Once
}

/**
try init db with configs:
	1、try struct config
	2、try loaded ini.File
	3、try ini file path
*/
func (c *MysqlClient) Start() error {
	var err error
	c.startOnce.Do(func() {
		if c.DbConfig != nil {
			err = c.initDbsWithCommonConf(c.DbConfig)
		} else if c.DbConf != nil {
			err = c.initDbs(c.DbConf)
		} else {
			if c.DbConfPath == "" {
				c.DbConfPath = defaultDbConf
			}

			err = c.initObjForMysqldb(c.DbConfPath)
		}
	})
	return err
}

func (c *MysqlClient) Close() {
	c.closeOnce.Do(func() {
		c.closeMainDbs()
	})
}

func (c *MysqlClient) getReadDbs() []*DbWrap {
	return c.dbRead
}

func (c *MysqlClient) GetReadDbRandom() (*DbWrap, error) {
	return c.getReadDbRandomly()
}

func (c *MysqlClient) getReadDbRandomly() (*DbWrap, error) {
	_db_read := c.getReadDbs()
	max := len(_db_read)
	if max <= 0 {
		return nil, fmt.Errorf("no read db found")
	}
	readDbIdx := rand.Intn(max)
	readDb := _db_read[readDbIdx]
	return readDb, nil
}

func (c *MysqlClient) getWriteDbs() *DbWrap {
	max := len(c.dbWrite)
	if max <= 1 {
		return c.dbWrite[0]
	}

	idx := rand.Intn(max)
	return c.dbWrite[idx]
}

func (c *MysqlClient) GetWriteDbs() *DbWrap {
	return c.getWriteDbs()
}

func (c *MysqlClient) getWriteDbsArray() []*DbWrap {
	return c.dbWrite
}

func getHostFromConnStr(connStr string) (string, error) {
	//mysql:"%s:%s@tcp(%s:%s)/%s?timeout=%s"
	fi := strings.Index(connStr, "@") + 5
	ei := strings.Index(connStr, "/") - 1
	if fi <= 0 || ei <= 0 || ei <= fi {
		err := fmt.Errorf("not find host in db conn str:%s,fi=%d,ei=%d", connStr, fi, ei)
		return "", err
	}
	host := connStr[fi:ei]
	if host == "" || host == ":" {
		return "", fmt.Errorf("host not found %s", connStr)
	}
	log.Debug("find host in db conn str:%s for %s", host, connStr)
	return host, nil
}

func (c *MysqlClient) initMainDbsMaxOpen(connMasters []string, connSlaves []string, maxOpen int, maxIdle int, ctxSuffix string, timeout time.Duration, masterProxy, slaveProxy bool) error {
	//log.Debug("open master=%v,slave=%v", connMasters, connSlaves)
	if len(connMasters) <= 0 {
		return fmt.Errorf("masters empty,master=%v,slave=%v", connMasters, connSlaves)
	}
	var dbWrites []*DbWrap
	for _, connMaster := range connMasters {
		db, err := sql.Open("mysql", connMaster)
		if err != nil {
			return err
		}
		db.SetMaxOpenConns(maxOpen)
		db.SetMaxIdleConns(maxIdle)
		hst, err := getHostFromConnStr(connMaster)
		if err != nil {
			return err
		}
		dbw := NewDbWrappedRetryProxy(hst, db, c, timeout, default_db_retry, masterProxy)
		dbw.ctxSuffix = ctxSuffix
		dbWrites = append(dbWrites, dbw)
	}
	c.dbWrite = dbWrites

	if connSlaves == nil || len(connSlaves) <= 0 {
		log.Info("read slaves empty, use master for read")
		c.dbRead = c.dbWrite
	} else {
		_db_read := make([]*DbWrap, len(connSlaves))
		for idx, rs := range connSlaves {
			//log.Info("open slave db %s", rs)
			d, err := sql.Open("mysql", rs)
			if err != nil {
				return err
			}
			hst, err := getHostFromConnStr(rs)
			if err != nil {
				return err
			}
			dbr := NewDbWrappedRetryProxy(hst, d, c, timeout, default_db_retry, slaveProxy)
			dbr.SetMaxOpenConns(maxOpen)
			dbr.SetMaxIdleConns(maxIdle)
			dbr.ctxSuffix = ctxSuffix
			_db_read[idx] = dbr
		}
		c.dbRead = _db_read
	}

	return nil
}

func (c *MysqlClient) closeMainDbs() {
	_db_read := c.dbRead
	_db_write := c.dbWrite
	for _, dbw := range _db_write {
		err := dbw.Close()
		if err != nil {
			log.Warn("write close err, %v", err)
		}
	}
	if _db_read != nil {
		for _, r := range _db_read {
			err := r.Close()
			if err != nil {
				log.Warn("read close err, %v", err)
			}
		}
	}
	log.Info("db close finish")
}

func (c *MysqlClient) GetCount(query string, args ...interface{}) (int64, error) {
	total := int64(0)

	row, err := c.queryRow(query, args...)
	if err != nil {
		return 0, err
	}

	err = row.Scan(&total)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		} else {
			return 0, err
		}
	} else {
		return total, nil
	}
}

func (c *MysqlClient) queryList(query string, args ...interface{}) (*sql.Rows, error) {
	readDbs := c.getReadDbs()
	if readDbs == nil || len(readDbs) == 0 {
		return nil, errors.New("no available db")
	}

	var randNums []int
	randNums = rand.Perm(len(readDbs))

	var ret *sql.Rows
	var err error
	var retry int
	for _, i := range randNums {
		readDb := readDbs[i]
		log.Debug("MysqlClient queryList, use db:%s", readDb.host)

		ret, err = readDb.Query(query, args...)
		if err == nil {
			return ret, err
		}

		if retry < 1 {
			retry++

			if IsDbConnError(err) {
				continue
			} else {
				errMsg := err.Error()
				if strings.Contains(errMsg, "getsockopt") {
					errt := reflect.TypeOf(err)
					log.Error("SOCKOPT_FAIL", "query err,type=%v,err=%v", errt, err)
					continue
				}
			}
		}

		return ret, err
	}

	log.Warn("no available db...")
	return nil, errors.New("no available db")
}

func (c *MysqlClient) queryRow(query string, args ...interface{}) (*Row, error) {
	readDbs := c.getReadDbs()
	if readDbs == nil || len(readDbs) == 0 {
		return nil, errors.New("no available db")
	}

	var randNums []int
	randNums = rand.Perm(len(readDbs))

	var row *Row
	var err error
	var retry int
	for _, i := range randNums {
		readDb := readDbs[i]
		log.Debug("MysqlClient queryRow, use db:%s", readDb.host)

		row = readDb.QueryRow(query, args...)
		err = row.err
		if err == nil {
			return row, err
		}

		if retry < 1 {
			retry++

			if IsDbConnError(err) {
				continue
			} else {
				errMsg := err.Error()
				if strings.Contains(errMsg, "getsockopt") {
					errt := reflect.TypeOf(err)
					log.Error("SOCKOPT_FAIL", "query err,type=%v,err=%v", errt, err)
					continue
				}
			}
		}

		return row, err
	}

	return nil, fmt.Errorf("no available db,lastErr=%v", err)
}

// no data return nil,nil
func (c *MysqlClient) Query(dataType interface{}, query string, args ...interface{}) (interface{}, error) {
	fieldNames, err := GetDataStructFields(dataType)
	if err != nil {
		return nil, err
	}

	typeOf := reflect.TypeOf(dataType).Elem()
	dataObj := reflect.New(typeOf).Interface()
	dests, err := GetDataStructDests(dataObj)
	if err != nil {
		return nil, err
	}

	query = strings.Replace(query, "?", strings.Join(fieldNames, ","), 1)

	row, err := c.queryRow(query, args...)
	if err != nil {
		return nil, err
	}

	err = row.Scan(dests...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	} else {
		return dataObj, nil
	}
}

func (c *MysqlClient) QueryList(dataType interface{}, query string, args ...interface{}) ([]interface{}, error) {
	fieldNames, err := GetDataStructFields(dataType)
	if err != nil {
		return nil, err
	}

	query = strings.Replace(query, "?", strings.Join(fieldNames, ","), 1)

	rows, err := c.queryList(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rets []interface{}
	for rows.Next() {
		typeOf := reflect.TypeOf(dataType).Elem()
		dataObj := reflect.New(typeOf).Interface()
		dests, err := GetDataStructDests(dataObj)
		if err != nil {
			return nil, err
		}
		if err = rows.Scan(dests...); err != nil {
			return nil, err
		}
		rets = append(rets, dataObj)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return rets, nil
}

// d must be a struct pointer
func (c *MysqlClient) Update(tableName string, d interface{}, primaryKeys map[string]interface{}, fieldsToUpdate []string) error {
	if len(primaryKeys) <= 0 {
		return errors.New("primary keys empty on update")
	}
	escapedName := MysqlEscapeString(tableName)
	tableName = escapedName

	// d must be a struct pointer
	typ := reflect.TypeOf(d)
	if typ == nil {
		return fmt.Errorf("input cannot be nil %v", typ)
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct type %v", typ)
	}

	writeDb := c.getWriteDbs()

	rv := reflect.ValueOf(d)
	te := rv.Elem()
	tt := te.Type()
	nf := te.NumField()
	var fieldNamesArray []string
	var dests []interface{}
	whereFields := ""
	for i := 0; i < nf; i++ {
		tf := tt.Field(i)
		mysqlFieldName := tf.Tag.Get("mysqlField")
		if mysqlFieldName != "" {
			if len(fieldsToUpdate) > 0 {
				found := false
				for _, f2u := range fieldsToUpdate {
					if f2u == mysqlFieldName {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			mysqlFieldName = "`" + mysqlFieldName + "`"
			fieldNamesArray = append(fieldNamesArray, mysqlFieldName+"=?")
		} else {
			continue
		}

		f := te.Field(i)
		if !f.CanInterface() {
			return fmt.Errorf("invalid struct field %s.%s", tt.Name(), tf.Name)
		}
		dests = append(dests, f.Interface())
	}
	fieldsNames := strings.Join(fieldNamesArray, ",")
	lpk := len(primaryKeys)
	pki := 0
	for pkn, pkv := range primaryKeys {
		if pki < lpk-1 {
			whereFields = whereFields + fmt.Sprintf(" `%s` = ? AND ", pkn)
		} else {
			whereFields = whereFields + fmt.Sprintf(" `%s` = ? ", pkn)
		}
		dests = append(dests, pkv)
		pki++
	}

	result, err := writeDb.Exec("update `"+tableName+"` set "+fieldsNames+" where "+whereFields, dests...)
	if err != nil {
		return err
	} else {
		log.Debug("update table=%s,data=%v,ret=%v,err=%v", tableName, d, result, err)
		return nil
	}
}

// d must be a struct pointer
func (c *MysqlClient) Add(tableName string, d interface{}, ondupUpdate bool) error {
	_, err := c.AddEscapeAutoIncr(tableName, d, ondupUpdate, "")
	return err
}

func (c *MysqlClient) AddEscapeAutoIncr(tableName string, d interface{}, ondupUpdate bool, atuoincrkey string) (int64, error) {
	result, err := c.addEscapeAutoIncr(tableName, d, ondupUpdate, atuoincrkey)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
}

//AddEscapeAutoIncrAndRetLastId执行纯插入操作(若数据已存在，则返回失败)，其会跳过由atuoincrkey指定的自增列，若执行成功，返回所插入的行id
func (c *MysqlClient) AddEscapeAutoIncrAndRetLastId(tableName string, d interface{}, atuoincrkey string) (int64, error) {
	sqlRet, err := c.addEscapeAutoIncr(tableName, d, false, atuoincrkey)
	if err != nil {
		return -1, err
	} else {
		return sqlRet.LastInsertId()
	}
}

func (c *MysqlClient) addEscapeAutoIncr(tableName string, d interface{}, ondupUpdate bool, atuoincrkey string) (sql.Result, error) {
	escapedName := MysqlEscapeString(tableName)
	tableName = escapedName

	// d must be a struct pointer
	typ := reflect.TypeOf(d)
	if typ == nil {
		return nil, fmt.Errorf("input cannot be nil %v", typ)
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct type %v", typ)
	}

	writeDb := c.getWriteDbs()

	rv := reflect.ValueOf(d)
	te := rv.Elem()
	tt := te.Type()
	nf := te.NumField()
	var fieldNamesArray []string
	var placeHoldersArray []string
	var dests []interface{}
	for i := 0; i < nf; i++ {
		tf := tt.Field(i)
		mysqlFieldName := tf.Tag.Get("mysqlField")
		if mysqlFieldName != "" {
			if atuoincrkey != "" && mysqlFieldName == atuoincrkey {
				continue
			}
			fieldNamesArray = append(fieldNamesArray, "`"+mysqlFieldName+"`")
			placeHoldersArray = append(placeHoldersArray, "?")
		} else {
			continue
		}

		f := te.Field(i)
		if !f.CanInterface() {
			return nil, fmt.Errorf("invalid struct field %s.%s", tt.Name(), tf.Name)
		}
		dests = append(dests, f.Interface())
	}
	fieldsNames := strings.Join(fieldNamesArray, ",")
	placeHolders := strings.Join(placeHoldersArray, ",")

	sqlStr := "insert into `" + tableName + "`(" + fieldsNames + ") values (" + placeHolders + ")"
	if ondupUpdate {
		setStr := ""
		fnl := len(fieldNamesArray)
		for i, fn := range fieldNamesArray {
			if i < fnl-1 {
				setStr += fmt.Sprintf("%s = ?,", fn)
			} else {
				setStr += fmt.Sprintf("%s = ?", fn)
			}
		}

		for _, dt := range dests {
			dests = append(dests, dt)
		}
		sqlStr = sqlStr + " ON DUPLICATE KEY UPDATE " + setStr
	}
	result, err := writeDb.Exec(sqlStr, dests...)
	if err != nil {
		return nil, err
	} else {
		log.Debug("insert table=%s, data=%v,ret=%v,err=%v", tableName, d, result, err)
		return result, nil
	}
}

func (c *MysqlClient) InsertOrUpdateOnDup(tableName string, d interface{}, primaryKeys []string, updateFields []string, useSqlOnDup bool) (int64, error) {
	if len(primaryKeys) <= 0 || len(updateFields) <= 0 {
		return 0, errors.New("primaryKeys or updateFields are nil")
	}

	escapedName := MysqlEscapeString(tableName)
	tableName = escapedName

	typ := reflect.TypeOf(d)
	if typ == nil {
		return 0, fmt.Errorf("input cannot be nil %v", typ)
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return 0, fmt.Errorf("not a struct type %v", typ)
	}

	rv := reflect.ValueOf(d)
	te := rv.Elem()
	tt := te.Type()
	nf := te.NumField()

	var insertSqlFieldNames []string
	var insertSqlPlaceHolders []string
	var insertSqlFieldValues []interface{}
	var updateSqlFieldNames []string
	var updateSqlFieldValues []interface{}
	var whereSqlFieldNames []string
	var whereSqlFieldValues []interface{}

	for i := 0; i < nf; i++ {
		tf := tt.Field(i)
		mysqlFieldName := tf.Tag.Get("mysqlField")
		if mysqlFieldName == "" {
			continue
		}

		f := te.Field(i)
		if !f.CanInterface() {
			return 0, fmt.Errorf("invalid struct field %s.%s", tt.Name(), tf.Name)
		}

		insertSqlFieldNames = append(insertSqlFieldNames, "`"+mysqlFieldName+"`")
		insertSqlPlaceHolders = append(insertSqlPlaceHolders, "?")
		insertSqlFieldValues = append(insertSqlFieldValues, f.Interface())

		for _, updateField := range updateFields {
			if updateField == mysqlFieldName {
				updateSqlFieldNames = append(updateSqlFieldNames, "`"+mysqlFieldName+"`=?")
				updateSqlFieldValues = append(updateSqlFieldValues, f.Interface())
				break
			}
		}

		for _, primaryKey := range primaryKeys {
			if primaryKey == mysqlFieldName {
				whereSqlFieldNames = append(whereSqlFieldNames, "`"+mysqlFieldName+"`=?")
				whereSqlFieldValues = append(whereSqlFieldValues, f.Interface())
				break
			}
		}
	}

	insertSql := fmt.Sprintf("insert into `%s` (%s) values (%s)", tableName, strings.Join(insertSqlFieldNames, ","), strings.Join(insertSqlPlaceHolders, ","))
	if useSqlOnDup {
		insertSql = fmt.Sprintf("%s ON DUPLICATE KEY UPDATE %s", insertSql, strings.Join(updateSqlFieldNames, ","))
		for _, value := range updateSqlFieldValues {
			insertSqlFieldValues = append(insertSqlFieldValues, value)
		}

		iRows, err := c.Execute(insertSql, insertSqlFieldValues...)
		log.Debug("InsertOrUpdateOnDup use SqlOnDup, table=%s, sql=%s, values=%v, ret=%d, err=%v", tableName, insertSql, insertSqlFieldValues, iRows, err)
		return iRows, err
	} else {
		updateSql := fmt.Sprintf("update `%s` set %s where %s", tableName, strings.Join(updateSqlFieldNames, ","), strings.Join(whereSqlFieldNames, " and "))
		for _, value := range whereSqlFieldValues {
			updateSqlFieldValues = append(updateSqlFieldValues, value)
		}

		uRows, err := c.Execute(updateSql, updateSqlFieldValues...)
		log.Debug("InsertOrUpdateOnDup no use SqlOnDup, table=%s, sql=%s, values=%v, ret=%d, err=%v", tableName, updateSql, updateSqlFieldValues, uRows, err)
		if err != nil {
			return uRows, err
		}

		if uRows == 0 {
			uRows, err = c.Execute(insertSql, insertSqlFieldValues...)
			log.Debug("InsertOrUpdateOnDup no use SqlOnDup, table=%s, sql=%s, values=%v, ret=%d, err=%v", tableName, insertSql, insertSqlFieldValues, uRows, err)
			if err != nil {
				if mysqlError, ok := err.(*mysql.MySQLError); ok {
					//1062 is duplicate entry error
					if mysqlError.Number == 1062 {
						log.Debug("InsertOrUpdateOnDup no use SqlOnDup, occur duplicate entry error, error:%v", mysqlError)

						uRows, err = c.Execute(updateSql, updateSqlFieldValues...)
					}
				}
			}
		}

		return uRows, err
	}
}

/*
	codition value supports limit kinds of slice:[]int64,[]string,[]interface
*/
func (c *MysqlClient) Delete(tableName string, condition map[string]interface{}) (int64, error) {
	escapedName := MysqlEscapeString(tableName)
	tableName = escapedName
	if len(condition) <= 0 {
		return 0, errors.New("del with empty condition?")
	}
	condStr, dest := buildWhereSql(condition)
	writeDb := c.getWriteDbs()
	rows, err := writeDb.Exec("delete from `"+tableName+"` where "+condStr, dest...)
	if err != nil {
		return 0, err
	}
	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

func (c *MysqlClient) Execute(sql string, args ...interface{}) (int64, error) {
	writeDb := c.getWriteDbs()
	result, err := writeDb.Exec(sql, args...)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (c *MysqlClient) ExecTransaction(transactionExec TransactionExec) (int64, error) {
	writeDb := c.getWriteDbs()
	result, err := writeDb.ExecTransaction(transactionExec)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
