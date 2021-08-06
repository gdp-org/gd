/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package redisdb

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/gdp-org/gd/dlog"
	"github.com/gdp-org/gd/runtime/gl"
	"github.com/gdp-org/gd/runtime/gr"
	"github.com/gdp-org/gd/runtime/pc"
	"github.com/gdp-org/gd/utls"
	"github.com/go-redis/redis/v8"
	"gopkg.in/ini.v1"
	"io"
	"strings"
	"sync"
	"time"
)

const (
	MaxGoroutinePoolSize = 5

	MGetCostMax                 = 60
	MSetCostMax                 = 60
	MDelCostMax                 = 60
	RedisClusterCommonCostMax   = 20
	RedisClusterCmdNormal       = "redis_cluster_cmd_normal"
	RedisClusterCmd             = "redis_cluster_cmd_%v"
	RedisClusterCmdSlowCount    = "redis_cluster_%v_slow_count"
	RedisClusterNormalSlowCount = "redis_cluster_common_slow_count"

	glRedisClusterCall     = "redisCluster_call"
	glRedisClusterCost     = "redisCluster_cost"
	glRedisClusterCallFail = "redisCluster_call_fail"
)

type RedisClusterConf struct {
	//cluster name
	ClusterName string

	//server address
	Addrs []string

	//conn/read/write timeout, 单位: 毫秒, 0表示默认值(1秒), -1表示不超时
	DialTimeout  int64
	ReadTimeout  int64
	WriteTimeout int64

	//pool config
	//每个redis节点会起一个连接池, 该poolSize表示每个redis节点的最大连接数
	PoolSize int
	//从连接池获取可用连接的超时, 单位: 毫秒, 0表示默认值(1秒), -1表示不超时
	PoolTimeout int64
	//最小空闲连接数
	MinIdleConns int
	//空闲连接可被回收的判断阈值, 单位: 秒
	IdleTimeout int64
	//空闲连接检查频率, 单位: 秒, 0表示默认值(30分钟), -1表示不做检查
	IdleCheckFrequency int64

	//当访问的key不在某节点或者某节点有异常, 会做move(redirect)操作, 该参数表示最大move操作次数, 0表示默认值(2次)
	MaxRedirects int

	Password string
}

var ErrNil = errors.New("redis: nil returned")

type RedisClusterCustomError string

func (e RedisClusterCustomError) Error() string { return string(e) }

type RedisCluster struct {
	clusterName   string
	clusterClient *redis.ClusterClient
	stop          chan bool
}

type RedisClusterClient struct {
	RedisConfig   *RedisClusterConf `inject:"redisClusterConfig" canNil:"true"`
	RedisConf     *ini.File         `inject:"redisClusterConf" canNil:"true"`
	RedisConfPath string            `inject:"redisClusterConfPath" canNil:"true"`
	ClusterName   string            `inject:"clusterName" canNil:"true"`

	redisCluster *RedisCluster
	startOnce    sync.Once
	closeOnce    sync.Once
}

func (r *RedisClusterClient) Start() error {
	var err error
	r.startOnce.Do(func() {
		if r.RedisConfig != nil {
			err = r.newRedisCluster(r.RedisConfig)
		} else if r.RedisConf != nil {
			err = r.initRedisCluster(r.RedisConf, r.ClusterName)
		} else {
			if r.RedisConfPath == "" {
				r.RedisConfPath = defaultConf
			}

			err = r.initObjForRedisCluster(r.RedisConfPath)
		}
	})
	return err
}

func (r *RedisClusterClient) Close() {
	r.closeOnce.Do(func() {
		if r.redisCluster.stop != nil {
			close(r.redisCluster.stop)
		}

		if r.redisCluster.clusterClient != nil {
			err := r.redisCluster.clusterClient.Close()
			if err != nil {
				return
			}
		}
		return
	})
}

func (r *RedisClusterClient) newRedisCluster(clusterConf *RedisClusterConf) error {
	if clusterConf == nil {
		return errors.New("redisClusterConf is nil")
	}

	addrs := clusterConf.Addrs
	if addrs == nil || len(addrs) == 0 {
		return errors.New("addrs not set")
	}

	//dialTimeout not allowed -1
	dialTimeout := time.Duration(clusterConf.DialTimeout) * time.Millisecond
	if dialTimeout <= 0 {
		dialTimeout = 1 * time.Second
	}

	readTimeout := time.Duration(clusterConf.ReadTimeout) * time.Millisecond
	if readTimeout < 0 {
		readTimeout = -1
	} else if readTimeout == 0 {
		readTimeout = 1 * time.Second
	}

	writeTimeout := time.Duration(clusterConf.WriteTimeout) * time.Millisecond
	if writeTimeout < 0 {
		writeTimeout = -1
	} else if writeTimeout == 0 {
		writeTimeout = 1 * time.Second
	}

	poolSize := clusterConf.PoolSize
	if poolSize <= 0 {
		poolSize = 100
	}

	poolTimeout := time.Duration(clusterConf.PoolTimeout) * time.Millisecond
	if poolTimeout < 0 {
		poolTimeout = -1
	} else if poolTimeout == 0 {
		poolTimeout = 1 * time.Second
	}

	minIdleConns := clusterConf.MinIdleConns
	if minIdleConns < 0 {
		minIdleConns = 0
	}

	idleTimeout := time.Duration(clusterConf.IdleTimeout) * time.Second
	if idleTimeout <= 0 {
		idleTimeout = 30 * time.Minute
	}

	idleCheckFrequency := time.Duration(clusterConf.IdleCheckFrequency) * time.Second
	if idleCheckFrequency < 0 {
		idleCheckFrequency = -1
	} else if idleCheckFrequency == 0 {
		idleCheckFrequency = 30 * time.Minute
	}

	maxRedirects := clusterConf.MaxRedirects
	if maxRedirects == 0 {
		maxRedirects = 3
	}

	clusterOptions := &redis.ClusterOptions{
		Addrs:              addrs,
		DialTimeout:        dialTimeout,
		ReadTimeout:        readTimeout,
		WriteTimeout:       writeTimeout,
		PoolSize:           poolSize,
		PoolTimeout:        poolTimeout,
		MinIdleConns:       minIdleConns,
		IdleTimeout:        idleTimeout,
		IdleCheckFrequency: idleCheckFrequency,
		MaxRedirects:       maxRedirects,
		MinRetryBackoff:    1 * time.Millisecond,
		MaxRetryBackoff:    10 * time.Millisecond,
		Password:           clusterConf.Password,
	}

	clusterClient := redis.NewClusterClient(clusterOptions)

	_, err := clusterClient.Ping(context.TODO()).Result()
	if err != nil {
		log.Error("init cluster client fail, %v", err)
	}

	ret := &RedisCluster{
		clusterName:   clusterConf.ClusterName,
		clusterClient: clusterClient,
	}

	ret.stop = make(chan bool)

	r.redisCluster = ret
	return nil
}

func (r *RedisClusterClient) initRedisCluster(f *ini.File, cn string) error {
	c := f.Section(fmt.Sprintf("%s.%s", "Redis", cn))
	addr := c.Key("addr").String()
	poolSize, _ := c.Key("poolSize").Int()
	maxRedirects, _ := c.Key("maxRedirects").Int()
	poolTimeout, _ := c.Key("poolTimeout").Int64()
	minIdleConns, _ := c.Key("minIdleConns").Int()
	idleTimeout, _ := c.Key("idleTimeout").Int64()
	idleCheckFrequency, _ := c.Key("idleCheckFrequency").Int64()
	connTimeout, _ := c.Key("dialTimeout").Int64()
	readTimeout, _ := c.Key("readTimeout").Int64()
	writeTimeout, _ := c.Key("writeTimeout").Int64()
	password := c.Key("password").String()

	addrs := strings.Split(addr, ",")
	err := r.newRedisCluster(&RedisClusterConf{
		ClusterName:        cn,
		Addrs:              addrs,
		PoolSize:           poolSize,
		PoolTimeout:        poolTimeout,
		MinIdleConns:       minIdleConns,
		IdleTimeout:        idleTimeout,
		DialTimeout:        connTimeout,
		ReadTimeout:        readTimeout,
		WriteTimeout:       writeTimeout,
		IdleCheckFrequency: idleCheckFrequency,
		MaxRedirects:       maxRedirects,
		Password:           password,
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClusterClient) initObjForRedisCluster(redisConfPath string) error {
	redisConfRealPath := redisConfPath
	if redisConfRealPath == "" {
		return errors.New("redisConf not set in g_cfg")
	}

	if !strings.HasSuffix(redisConfRealPath, ".ini") {
		return errors.New("redisConf not an ini file")
	}

	redisConf, err := ini.Load(redisConfRealPath)
	if err != nil {
		return err
	}

	if err = r.initRedisCluster(redisConf, r.ClusterName); err != nil {
		return err
	}
	return nil
}

func (r *RedisClusterClient) getClusterClient() *redis.ClusterClient {
	return r.redisCluster.clusterClient
}

func reportPerf(clusterName string, cmdName string, sTime time.Time, err error, key interface{}) {
	cost := time.Now().Sub(sTime)

	if cmdName == "MGet" || cmdName == "MSet" || cmdName == "MDel" {
		var slow bool
		if cmdName == "MGet" && cost/time.Millisecond > MGetCostMax {
			slow = true
		} else if cmdName == "MSet" && cost/time.Millisecond > MSetCostMax {
			slow = true
		} else if cmdName == "MDel" && cost/time.Millisecond > MDelCostMax {
			slow = true
		}
		if slow {
			pc.Incr(fmt.Sprintf(RedisClusterCmdSlowCount, strings.ToLower(cmdName)), 1)
			if cost/time.Millisecond > 200 {
				log.Warn("redisCluster slow, cluster:%s, cmd:%v, key:%v, cost:%v", clusterName, cmdName, key, cost)
			}
		}
		pcKey := fmt.Sprintf(RedisClusterCmd, strings.ToLower(cmdName))
		pc.Cost(fmt.Sprintf("rediscluster,name=%s,cmd=%s", clusterName, pcKey), cost)
	} else {
		if cost/time.Millisecond > RedisClusterCommonCostMax {
			pc.Incr(RedisClusterNormalSlowCount, 1)
			if cost/time.Millisecond > 100 {
				log.Warn("redisCluster slow, cluster:%s, cmd:%v, key:%v, cost:%v", clusterName, cmdName, key, cost)
			}
		}
		pc.Cost(fmt.Sprintf("redisCluster,name=%s,cmd=%s", clusterName, RedisClusterCmdNormal), cost)
	}

	gl.Incr(glRedisClusterCall, 1)
	gl.IncrCost(glRedisClusterCost, cost)

	if log.IsEnabledFor(log.DEBUG) {
		log.Debug("RedisCluster call, cluster=%s, cmd=%s, key=%v, cost=%d ms, err=%v", clusterName, cmdName, key, cost/time.Millisecond, err)
	}

	if err != nil && err != redis.Nil && err != ErrNil {
		if strings.Index(err.Error(), "WRONGTYPE Operation") != -1 {
			log.Info("RedisCluster call fail, cluster=%s, cmd=%s, key=%v, cost=%d ms, err=%v", clusterName, cmdName, key, cost/time.Millisecond, err)
		} else {
			pc.CostFail(fmt.Sprintf("rediscluster,name=%s", clusterName), 1)
			gl.Incr(glRedisClusterCallFail, 1)
			log.Warn("RedisCluster call fail, cluster=%s, cmd=%s, key=%v, cost=%d ms, err=%v", clusterName, cmdName, key, cost/time.Millisecond, err)
		}
	}

	return
}

/**
1. 若key存在且成功, 返回(string,nil)
2. 若key不存在且成功, 返回("",redisCluster.ErrNil)
3. 若异常, 返回("",error)
*/
func (r *RedisClusterClient) Get(key string) (string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "Get", st, err, key)
	}()

	ret, err := clusterClient.Get(context.TODO(), key).Result()
	if err == redis.Nil {
		return "", ErrNil
	}

	return ret, err
}

/**
1. 若expire为0, 表示不设置过期
2. 若设置成功, 返回("ok",nil)
3. 若异常, 返回("",error)
*/
func (r *RedisClusterClient) Set(key string, value string, expire time.Duration) (string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "Set", st, err, key)
	}()

	ret, err := clusterClient.Set(context.TODO(), key, value, expire).Result()
	return ret, err
}

/**
1. 若expire为0, 表示不设置过期
2. 如果 err 不为空, 则发生异常;
3. 在 err 为空的情况下, bool=false 表示key已存在set无效, bool=true表示key不存在set成功
https://redis.io/commands/setnx
*/
func (r *RedisClusterClient) SetNX(key string, value string, expire time.Duration) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "SetNX", st, err, key)
	}()

	isNew, err := clusterClient.SetNX(context.TODO(), key, value, expire).Result()
	return isNew, err
}

/**
1. 若正常, 返回 (num,nil), num为删除的key个数
2. 若异常, 返回 (0,error)
*/
func (r *RedisClusterClient) Del(key string) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "Del", st, err, key)
	}()

	ret, err := clusterClient.Del(context.TODO(), key).Result()
	return ret, err
}

/**
1. 若有异常, 返回 (nil,error)
2. 若正常, 返回 ([]interface{}{...},nil), 其中结果里的value顺序与keys的顺序一一对应, 若某个key不存在, 所对应value为nil
	例如 MGet([]string{"key1","key2","key3"}), 假设key2不存在, 返回数据如下:
	 []interface{}{
	 	"value1",		//key1的值
	 	nil,			//不存在
	 	"value3",		//key3的值
	 }, nil
*/
func (r *RedisClusterClient) MGet(keys []string) ([]interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "MGet", st, err, keys)
	}()

	retMap := make(map[string]interface{}, len(keys))
	errMsg := make([]error, 0, len(keys))

	//or use : clusterClient.Pipelined()
	pipeline := clusterClient.Pipeline()
	defer pipeline.Close()
	for _, k := range keys {
		pipeline.Get(context.TODO(), k)
	}
	cmds, err := pipeline.Exec(context.TODO())
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for i, cmd := range cmds {
		k := keys[i]
		getCmd := cmd.(*redis.StringCmd)
		getVal, err := getCmd.Result()
		if err != nil {
			if err == redis.Nil {
				retMap[k] = nil
			} else {
				errMsg = append(errMsg, err)
			}
		} else {
			retMap[k] = getVal
		}

	}

	if len(errMsg) != 0 {
		//part fail
		log.Warn("redisCluster MGet error, errMsg:%v", errMsg)
	}
	ret := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		ret = append(ret, retMap[key])
	}
	log.Debug("redisCluster MGet in:%v, out:%v", keys, ret)
	return ret, nil

}

/**
1. 若有异常, 返回 (nil,error)
2. 若正常, 返回 ([]interface{}{...},nil), 其中结果里的value顺序与keys的顺序一一对应, 若某个key不存在, 所对应value为nil
	例如 MGet([]string{"key1","key2","key3"}), 假设key2不存在, 返回数据如下:
	 []interface{}{
	 	"value1",		//key1的值
	 	nil,			//不存在
	 	"value3",		//key3的值
	 }, nil
*/
func (r *RedisClusterClient) MGet2(keys []string) ([]interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "MGet2", st, err, keys)
	}()

	nodeAndKeyMap := r.getNodeAndKeyMap(keys)
	poolSize := len(nodeAndKeyMap)
	if poolSize > MaxGoroutinePoolSize {
		poolSize = MaxGoroutinePoolSize
	}

	fixPool := &gr.FixedGoroutinePool{Size: int64(poolSize)}
	err = fixPool.Start()
	if err != nil {
		return nil, err
	}

	retMap := make(map[string]interface{}, len(keys))
	errMsg := make([]error, 0, len(keys))
	lock := &sync.RWMutex{}
	wg := &sync.WaitGroup{}
	for _, keysItem := range nodeAndKeyMap {
		wg.Add(1)
		var keysItemTmp = keysItem
		fixPool.Execute(func() {
			defer wg.Done()
			pipeline := clusterClient.Pipeline()
			for _, keyItem := range keysItemTmp {
				pipeline.Get(context.TODO(), keyItem)
			}
			cmds, err := pipeline.Exec(context.TODO())
			if err != nil && err != redis.Nil {
				lock.Lock()
				errMsg = append(errMsg, err)
				lock.Unlock()
				return
			}

			for i, keyItem := range keysItemTmp {
				getCmd := cmds[i].(*redis.StringCmd)
				getVal, err := getCmd.Result()
				if err != nil {
					if err == redis.Nil {
						lock.Lock()
						retMap[keyItem] = nil
						lock.Unlock()
					} else {
						lock.Lock()
						errMsg = append(errMsg, err)
						lock.Unlock()
						return
					}
				} else {
					lock.Lock()
					retMap[keyItem] = getVal
					lock.Unlock()
				}

			}
		})
	}
	wg.Wait()
	fixPool.Close()
	if len(errMsg) != 0 {
		log.Warn("redisCluster MGet error, errMsg:%v", errMsg)
		err = errMsg[0]
		return nil, errMsg[0]
	}
	ret := make([]interface{}, 0, len(keys))
	for _, key := range keys {
		ret = append(ret, retMap[key])
	}
	log.Debug("redisCluster MGet in:%v, out:%v", keys, ret)
	return ret, nil

}

func (r *RedisClusterClient) MSet(kvs map[string]string, expire time.Duration) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "MSet", st, err, kvs)
	}()

	keys := make([]string, 0, len(kvs))
	retMap := make(map[string]bool, len(kvs))
	errMsg := make([]error, 0, len(kvs))

	pipeline := clusterClient.Pipeline()
	defer pipeline.Close()
	for k, v := range kvs {
		keys = append(keys, k)
		pipeline.Set(context.TODO(), k, v, expire)
	}
	cmds, err := pipeline.Exec(context.TODO())
	if err != nil {
		return nil, err
	}

	for i, cmd := range cmds {
		setCmd := cmd.(*redis.StatusCmd)
		_, err := setCmd.Result()
		if err != nil {
			errMsg = append(errMsg, err)
		} else {
			retMap[keys[i]] = true
		}

	}

	if len(errMsg) != 0 {
		log.Warn("redisCluster MSet error, errMsg:%v", errMsg)
	}

	ret := make(map[string]bool, len(keys))
	for _, key := range keys {
		ret[key] = retMap[key]
	}

	log.Debug("redisCluster MSet in:%v, out:%v", kvs, ret)
	return ret, err
}

/**
1. err!=nil, 在err!=nil的情况下，若ret为空，则表示都失败了
2. ret标识每个key对应的成功与否，true成功，false失败
有如下情况：
ret==nil, err!=nil: 全部失败
ret!=nil, err!=nil: 部分失败, ret里会包含所有key的操作结果
ret!=nil, err==nil: 全部成功, ret里会包含所有key的操作结果

备注: expire为0表示key不过期
*/
func (r *RedisClusterClient) MSet2(kvs map[string]string, expire time.Duration) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "MSet2", st, err, kvs)
	}()

	keys := make([]string, 0, len(kvs))
	for k, _ := range kvs {
		keys = append(keys, k)
	}
	nodeAndKeyMap := r.getNodeAndKeyMap(keys)
	poolSize := len(nodeAndKeyMap)
	if poolSize > MaxGoroutinePoolSize {
		poolSize = MaxGoroutinePoolSize
	}

	fixPool := &gr.FixedGoroutinePool{Size: int64(poolSize)}
	err = fixPool.Start()
	if err != nil {
		return nil, err
	}

	retMap := make(map[string]bool, len(keys))
	errMsg := make([]error, 0, len(keys))
	lock := &sync.RWMutex{}
	wg := &sync.WaitGroup{}
	for _, keysItem := range nodeAndKeyMap {
		wg.Add(1)
		var keysItemTmp = keysItem
		fixPool.Execute(func() {
			defer wg.Done()
			pipeline := clusterClient.Pipeline()
			for _, keyItem := range keysItemTmp {
				pipeline.Set(context.TODO(), keyItem, kvs[keyItem], expire)
			}
			cmds, err := pipeline.Exec(context.TODO())
			if err != nil {
				lock.Lock()
				errMsg = append(errMsg, err)
				lock.Unlock()
				return
			}

			for i, keyItem := range keysItemTmp {
				getCmd := cmds[i].(*redis.StatusCmd)
				_, err := getCmd.Result()
				if err != nil {
					lock.Lock()
					errMsg = append(errMsg, err)
					lock.Unlock()
					return
				} else {
					lock.Lock()
					retMap[keyItem] = true
					lock.Unlock()
				}

			}
		})
	}
	wg.Wait()
	fixPool.Close()

	//var errMsgItem error
	if len(errMsg) != 0 {
		log.Warn("redisCluster MSet error, errMsg:%v", errMsg)
		err = errMsg[0]
		if len(errMsg) == len(keys) {
			return nil, err
		}
	}

	ret := make(map[string]bool, len(keys))
	for _, key := range keys {
		ret[key] = retMap[key]
	}

	log.Debug("redisCluster MSet in:%v, out:%v", kvs, ret)
	return ret, err
}

func (r *RedisClusterClient) MDel(keys []string) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "MDel", st, err, keys)
	}()

	retMap := make(map[string]bool, len(keys))
	errMsg := make([]error, 0, len(keys))

	//or use : clusterClient.Pipelined()
	pipeline := clusterClient.Pipeline()
	defer pipeline.Close()
	for _, k := range keys {
		pipeline.Del(context.TODO(), k)
	}
	cmds, err := pipeline.Exec(context.TODO())
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for i, cmd := range cmds {
		k := keys[i]
		delCmd := cmd.(*redis.IntCmd)
		delRet, err := delCmd.Result()
		log.Debug("delRet:%v, err:%v", delRet, err)
		if err != nil {
			errMsg = append(errMsg, err)
		} else {
			//因为delRet==0表示key不存在, 返回true
			retMap[k] = true
		}

	}

	if len(errMsg) != 0 {
		//part fail
		log.Warn("redisCluster MDel error, errMsg:%v", errMsg)
	}
	log.Debug("redisCluster MDel in:%v, out:%v", keys, retMap)
	return retMap, nil
}

/**
1. err!=nil, 在err!=nil的情况下，若ret为空，则表示都失败了
2. ret标识每个key对应的成功与否，true成功，false失败
有如下情况：
ret==nil, err!=nil: 全部失败
ret!=nil, err!=nil: 部分失败, ret会包含所有key的操作结果
ret!=nil, err==nil: 全部成功, ret会包含所有key的操作结果
*/
func (r *RedisClusterClient) MDel2(keys []string) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "MDel2", st, err, keys)
	}()

	nodeAndKeyMap := r.getNodeAndKeyMap(keys)
	poolSize := len(nodeAndKeyMap)
	if poolSize > MaxGoroutinePoolSize {
		poolSize = MaxGoroutinePoolSize
	}

	fixPool := &gr.FixedGoroutinePool{Size: int64(poolSize)}
	err = fixPool.Start()
	if err != nil {
		return nil, err
	}

	retMap := make(map[string]bool, len(keys))
	errMsg := make([]error, 0, len(keys))
	lock := &sync.RWMutex{}
	wg := &sync.WaitGroup{}
	for _, keysItem := range nodeAndKeyMap {
		wg.Add(1)
		var keysItemTmp = keysItem
		fixPool.Execute(func() {
			defer wg.Done()
			pipeline := clusterClient.Pipeline()
			for _, keyItem := range keysItemTmp {
				pipeline.Del(context.TODO(), keyItem)
			}
			cmds, err := pipeline.Exec(context.TODO())
			if err != nil {
				lock.Lock()
				errMsg = append(errMsg, err)
				lock.Unlock()
				return
			}

			for i, keyItem := range keysItemTmp {
				delCmd := cmds[i].(*redis.IntCmd)
				_, err := delCmd.Result()
				if err != nil {
					lock.Lock()
					errMsg = append(errMsg, err)
					lock.Unlock()
					return
				} else {
					lock.Lock()
					retMap[keyItem] = true
					lock.Unlock()
				}

			}
		})
	}
	wg.Wait()
	fixPool.Close()

	//var errMsgItem error
	if len(errMsg) != 0 {
		log.Warn("redisCluster MDel error, errMsg:%v", errMsg)
		err = errMsg[0]
		if len(errMsg) == len(keys) {
			return nil, err
		}
	}

	ret := make(map[string]bool, len(keys))
	for _, key := range keys {
		ret[key] = retMap[key]
	}

	log.Debug("redisCluster MDel in:%v, out:%v", keys, ret)
	return ret, err
}

func (r *RedisClusterClient) getNodeAndKeyMap(keys []string) map[string][]string {
	nodeAndKeyMap := make(map[string][]string)
	return nodeAndKeyMap
}

/**
1. 异常, 返回 ("",error)
2. 正常, 但key不存在或者field不存在, 返回 ("",redisCluster.ErrNil)
3. 正常, 且key存在, filed存在, 返回 (string, nil)
*/
func (r *RedisClusterClient) HGet(key string, field string) (string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HGet", st, err, key)
	}()

	ret, err := clusterClient.HGet(context.TODO(), key, field).Result()
	if err == redis.Nil {
		return "", ErrNil
	}

	return ret, err
}

/**
1. 异常, 返回 (nil, error)
2. 正常, 返回 ([]interface{}, nil), 其中slice里的值顺序与fields一一对应, 对于不存在的field, 对应值为nil
	例如: HMGet("key", []string{"field1","field2","field3"})
	返回值:
		[]interface{}{
			"value1",	//field1的值
			nil,		//field2不存在
			"value3"	//field3的值
		},nil
	备注: 若key不存在, slice里的所有值都为nil
*/
func (r *RedisClusterClient) HMGet(key string, fields []string) ([]interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HMGet", st, err, key)
	}()

	if fields == nil || len(fields) == 0 {
		return make([]interface{}, 0), nil
	}

	ret, err := clusterClient.HMGet(context.TODO(), key, fields...).Result()
	return ret, err
}

/**
1. 若异常, 返回 (nil, error)
2. 若正常, 返回 (map[string]string, nil)

Notice: 这个函数并不能严格的限制返回count个，简单测试看起来主要和redis用的存储结构有关。
        Hmap redis在存储的时候如果数据比较少（看文章是512）会使用ziplist，测试了下，在ziplist存储的状态下，会都返回，count不生效
        如果数据超过512之后会使用hmap来存储，这时基本就是准确的了。所以这个函数只能保证返回 >= count
*/
func (r *RedisClusterClient) HScan(key string, count int64) (map[string]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HScan", st, err, key)
	}()

	ret, _, err := clusterClient.HScan(context.TODO(), key, 0, "", count).Result()
	if len(ret)%2 != 0 {
		return nil, fmt.Errorf("hscan return invalid")
	}
	result := make(map[string]string, len(ret)/2)
	for i := 0; i < len(ret); i += 2 {
		result[ret[i]] = ret[i+1]
	}
	return result, err
}

/**
1. 若异常, 返回 (nil, error)
2. 若正常, 返回 (map[string]string, nil)
*/
func (r *RedisClusterClient) HGetAll(key string) (map[string]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HGetAll", st, err, key)
	}()

	ret, err := clusterClient.HGetAll(context.TODO(), key).Result()
	return ret, err
}

/**
1. 若异常, 返回 (false, error)
2. 若正常, 返回 (int64, nil)
*/
func (r *RedisClusterClient) HSet(key string, field string, value string) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HSet", st, err, key)
	}()

	ret, err := clusterClient.HSet(context.TODO(), key, field, value).Result()
	return ret, err
}

/**
1. 若异常, 返回 (false, error)
2. 若正常, 返回 (bool, nil), 其中true: field在hash中不存在且新增成功; false: field已在hash中存在, 不做更新.
*/
func (r *RedisClusterClient) HSetNx(key string, field string, value string) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HSetNx", st, err, key)
	}()

	ret, err := clusterClient.HSetNX(context.TODO(), key, field, value).Result()
	return ret, err
}

/**
1. 若异常, 返回 (false, error)
2. 若正常, 返回 (true, nil)
*/
func (r *RedisClusterClient) HMSet(key string, fields map[string]string) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HMSet", st, err, key)
	}()

	if fields == nil || len(fields) == 0 {
		return true, nil
	}

	tmpFields := make(map[string]interface{}, len(fields))
	for fieldName, fieldValue := range fields {
		tmpFields[fieldName] = fieldValue
	}

	_, err = clusterClient.HMSet(context.TODO(), key, tmpFields).Result()
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

/**
1. 若正常, 返回 (num,nil), num为删除的key个数
2. 若异常, 返回 (0,error)
*/

func (r *RedisClusterClient) HMDel(key string, fields []string) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HMDel", st, err, key)
	}()

	return clusterClient.HDel(context.TODO(), key, fields...).Result()
}

/**
1. 若正常, 返回 (true,nil)
2. 若异常, 返回 (false,error)
  key不存在返回(false,nil)
*/
func (r *RedisClusterClient) Expire(key string, expiration time.Duration) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "Expire", st, err, key)
	}()

	return clusterClient.Expire(context.TODO(), key, expiration).Result()
}

/**
1. 若正常, 返回 (true,nil)
2. 若异常, 返回 (false,error)
  key不存在返回(false,nil)
*/
func (r *RedisClusterClient) PExpire(key string, expiration time.Duration) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "PExpire", st, err, key)
	}()

	return clusterClient.PExpire(context.TODO(), key, expiration).Result()
}

/**
1. 若正常, 返回 (true,nil)
2. 若异常, 返回 (false,error)
  key不存在返回(false,nil)
*/
func (r *RedisClusterClient) PTtl(key string) (time.Duration, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "PTtl", st, err, key)
	}()

	return clusterClient.PTTL(context.TODO(), key).Result()
}

/**
1. 若设置成功, 返回(int64,nil)
2. 若异常, 返回(-1,error)
*/
func (r *RedisClusterClient) HIncrBy(key string, field string, value int64) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HIncrBy", st, err, key)
	}()

	ret, err := clusterClient.HIncrBy(context.TODO(), key, field, value).Result()
	return ret, err
}

func (r *RedisClusterClient) IncrBy(key string, value int64) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "IncrBy", st, err, key)
	}()

	ret, err := clusterClient.IncrBy(context.TODO(), key, value).Result()
	return ret, err
}

func (r *RedisClusterClient) Eval(script string, keys []string, args []interface{}) (interface{}, error) {
	sha := getScriptSha(script)
	ret, err := r.EvalSha(sha, keys, args)
	if err == nil {
		return ret, err
	}
	if str := err.Error(); !strings.Contains(str, "NOSCRIPT") {
		return ret, err
	}
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "Eval", st, err, keys)
	}()

	ret, err = clusterClient.Eval(context.TODO(), script, keys, args...).Result()
	return ret, err
}

func (r *RedisClusterClient) EvalSha(scriptSha string, keys []string, args []interface{}) (interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "IncrBy", st, err, keys)
	}()

	ret, err := clusterClient.EvalSha(context.TODO(), scriptSha, keys, args...).Result()
	return ret, err
}

func getScriptSha(script string) string {
	hash := sha1.New()
	io.WriteString(hash, script)
	return hex.EncodeToString(hash.Sum(nil))
}

func (r *RedisClusterClient) ZAdd(key string, members []*redis.Z) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZAdd", st, err, key)
	}()

	return clusterClient.ZAdd(context.TODO(), key, members...).Result()
}

type ZSetResult struct {
	Member string
	Score  float64
}

func (r *RedisClusterClient) ZRangeByScoreWithScores(key string, min, max string, offset, limit int64) ([]*ZSetResult, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZRangeByScoreWithScores", st, err, key)
	}()

	opt := &redis.ZRangeBy{
		Count:  limit,
		Max:    max,
		Min:    min,
		Offset: offset,
	}
	result, err := clusterClient.ZRangeByScoreWithScores(context.TODO(), key, opt).Result()
	if err != nil {
		return nil, err
	}
	ret := make([]*ZSetResult, 0, len(result))
	for _, item := range result {
		ret = append(ret, &ZSetResult{
			Score:  item.Score,
			Member: utls.MustString(item.Member, ""),
		})
	}
	return ret, nil
}

func (r *RedisClusterClient) ZRevRangeByScoreWithScores(key string, min, max string, offset, limit int64) ([]*ZSetResult, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZRevRangeByScoreWithScores", st, err, key)
	}()

	opt := &redis.ZRangeBy{
		Count:  limit,
		Max:    max,
		Min:    min,
		Offset: offset,
	}
	log.Debug("ZRevRangeByScoreWithScores key=%v,opt=%v", key, opt)
	result, err := clusterClient.ZRevRangeByScoreWithScores(context.TODO(), key, opt).Result()
	if err != nil {
		return nil, err
	}
	ret := make([]*ZSetResult, 0, len(result))
	for _, item := range result {
		ret = append(ret, &ZSetResult{
			Score:  item.Score,
			Member: utls.MustString(item.Member, ""),
		})
	}
	return ret, nil
}

func (r *RedisClusterClient) ZRange(key string, start, stop int64) ([]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZRange", st, err, key)
	}()

	return clusterClient.ZRange(context.TODO(), key, start, stop).Result()
}

func (r *RedisClusterClient) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZRemRangeByRank", st, err, key)
	}()

	return clusterClient.ZRemRangeByRank(context.TODO(), key, start, stop).Result()
}

func (r *RedisClusterClient) ZRemRangeByScore(key string, min, max string) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZRemRangeByScore", st, err, key)
	}()

	return clusterClient.ZRemRangeByScore(context.TODO(), key, min, max).Result()
}

func (r *RedisClusterClient) ZRem(key string, members ...interface{}) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZRem", st, err, key)
	}()

	return clusterClient.ZRem(context.TODO(), key, members...).Result()
}

func (r *RedisClusterClient) ZRevRange(key string, start, stop int64) ([]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "ZRevRange", st, err, key)
	}()

	return clusterClient.ZRevRange(context.TODO(), key, start, stop).Result()
}

func (r *RedisClusterClient) SetBit(key string, offset int64, value int) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "SetBit", st, err, key)
	}()

	return clusterClient.SetBit(context.TODO(), key, offset, value).Result()
}

func (r *RedisClusterClient) Exist(key string) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "Exist", st, err, key)
	}()

	var ret bool
	var existRet int64
	existRet, err = clusterClient.Exists(context.TODO(), key).Result()
	if err != nil {
		ret = false
	} else if existRet > 0 {
		ret = true
	} else {
		ret = false
	}

	return ret, err
}

func (r *RedisClusterClient) SAdd(key string, members []interface{}) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "SAdd", st, err, key)
	}()

	return clusterClient.SAdd(context.TODO(), key, members...).Result()

}

func (r *RedisClusterClient) SPop(key string) (string, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "SPop", st, err, key)
	}()

	return clusterClient.SPop(context.TODO(), key).Result()
}

func (r *RedisClusterClient) LPop(key string) (string, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "LPop", st, err, key)
	}()

	return clusterClient.LPop(context.TODO(), key).Result()
}

func (r *RedisClusterClient) LIndex(key string, index int64) (string, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "LIndex", st, err, key)
	}()

	return clusterClient.LIndex(context.TODO(), key, index).Result()
}

func (r *RedisClusterClient) LPush(key string, value string) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "LPush", st, err, key)
	}()

	return clusterClient.LPush(context.TODO(), key, value).Result()
}

func (r *RedisClusterClient) RPush(key string, value string) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "RPush", st, err, key)
	}()

	return clusterClient.RPush(context.TODO(), key, value).Result()
}

func (r *RedisClusterClient) HLen(key string) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.redisCluster.clusterName, "HLen", st, err, key)
	}()

	return clusterClient.HLen(context.TODO(), key).Result()
}
