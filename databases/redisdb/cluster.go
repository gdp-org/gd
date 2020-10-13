/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package redisdb

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	log "github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/runtime/gl"
	"github.com/chuck1024/gd/runtime/gr"
	"github.com/chuck1024/gd/runtime/pc"
	"github.com/chuck1024/gd/utls"
	"github.com/go-redis/redis"
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
}

var ErrNil = errors.New("redis: nil returned")

type RedisClusterCustomError string

func (e RedisClusterCustomError) Error() string { return string(e) }

type RedisCluster struct {
	clusterName   string
	clusterClient *redis.ClusterClient
	stop          chan bool
}

func NewRedisCluster(clusterConf *RedisClusterConf) (*RedisCluster, error) {
	if clusterConf == nil {
		return nil, errors.New("redisClusterConf is nil")
	}

	addrs := clusterConf.Addrs
	if addrs == nil || len(addrs) == 0 {
		return nil, errors.New("addrs not set")
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
	}

	clusterClient := redis.NewClusterClient(clusterOptions)

	_, err := clusterClient.Ping().Result()
	if err != nil {
		log.Error("init cluster client fail, %v", err)
	}

	ret := &RedisCluster{
		clusterName:   clusterConf.ClusterName,
		clusterClient: clusterClient,
	}

	ret.stop = make(chan bool)

	return ret, nil
}

func (r *RedisCluster) getClusterClient() *redis.ClusterClient {
	return r.clusterClient
}

func (r *RedisCluster) Close() error {
	if r.stop != nil {
		close(r.stop)
	}

	if r.clusterClient != nil {
		err := r.clusterClient.Close()
		if err != nil {
			return err
		}
	}

	return nil
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
func (r *RedisCluster) Get(key string) (string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "Get", st, err, key)
	}()

	ret, err := clusterClient.Get(key).Result()
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
func (r *RedisCluster) Set(key string, value string, expire time.Duration) (string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "Set", st, err, key)
	}()

	ret, err := clusterClient.Set(key, value, expire).Result()
	return ret, err
}

/**
1. 若expire为0, 表示不设置过期
2. 如果 err 不为空, 则发生异常;
3. 在 err 为空的情况下, bool=false 表示key已存在set无效, bool=true表示key不存在set成功
https://redis.io/commands/setnx
*/
func (r *RedisCluster) SetNX(key string, value string, expire time.Duration) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "SetNX", st, err, key)
	}()

	isNew, err := clusterClient.SetNX(key, value, expire).Result()
	return isNew, err
}

/**
1. 若正常, 返回 (num,nil), num为删除的key个数
2. 若异常, 返回 (0,error)
*/
func (r *RedisCluster) Del(key string) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "Del", st, err, key)
	}()

	ret, err := clusterClient.Del(key).Result()
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
func (r *RedisCluster) MGet(keys []string) ([]interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "MGet", st, err, keys)
	}()

	retMap := make(map[string]interface{}, len(keys))
	errMsg := make([]error, 0, len(keys))

	//or use : clusterClient.Pipelined()
	pipeline := clusterClient.Pipeline()
	defer pipeline.Close()
	for _, k := range keys {
		pipeline.Get(k)
	}
	cmds, err := pipeline.Exec()
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
func (r *RedisCluster) MGet2(keys []string) ([]interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "MGet2", st, err, keys)
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
				pipeline.Get(keyItem)
			}
			cmds, err := pipeline.Exec()
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

func (r *RedisCluster) MSet(kvs map[string]string, expire time.Duration) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "MSet", st, err, kvs)
	}()

	keys := make([]string, 0, len(kvs))
	retMap := make(map[string]bool, len(kvs))
	errMsg := make([]error, 0, len(kvs))

	pipeline := clusterClient.Pipeline()
	defer pipeline.Close()
	for k, v := range kvs {
		keys = append(keys, k)
		pipeline.Set(k, v, expire)
	}
	cmds, err := pipeline.Exec()
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
func (r *RedisCluster) MSet2(kvs map[string]string, expire time.Duration) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "MSet2", st, err, kvs)
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
				pipeline.Set(keyItem, kvs[keyItem], expire)
			}
			cmds, err := pipeline.Exec()
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

func (r *RedisCluster) MDel(keys []string) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "MDel", st, err, keys)
	}()

	retMap := make(map[string]bool, len(keys))
	errMsg := make([]error, 0, len(keys))

	//or use : clusterClient.Pipelined()
	pipeline := clusterClient.Pipeline()
	defer pipeline.Close()
	for _, k := range keys {
		pipeline.Del(k)
	}
	cmds, err := pipeline.Exec()
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
func (r *RedisCluster) MDel2(keys []string) (map[string]bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "MDel2", st, err, keys)
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
				pipeline.Del(keyItem)
			}
			cmds, err := pipeline.Exec()
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

func (r *RedisCluster) getNodeAndKeyMap(keys []string) map[string][]string {
	nodeAndKeyMap := make(map[string][]string)
	return nodeAndKeyMap
}

/**
1. 异常, 返回 ("",error)
2. 正常, 但key不存在或者field不存在, 返回 ("",redisCluster.ErrNil)
3. 正常, 且key存在, filed存在, 返回 (string, nil)
*/
func (r *RedisCluster) HGet(key string, field string) (string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HGet", st, err, key)
	}()

	ret, err := clusterClient.HGet(key, field).Result()
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
func (r *RedisCluster) HMGet(key string, fields []string) ([]interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HMGet", st, err, key)
	}()

	if fields == nil || len(fields) == 0 {
		return make([]interface{}, 0), nil
	}

	ret, err := clusterClient.HMGet(key, fields...).Result()
	return ret, err
}

/**
1. 若异常, 返回 (nil, error)
2. 若正常, 返回 (map[string]string, nil)

Notice: 这个函数并不能严格的限制返回count个，简单测试看起来主要和redis用的存储结构有关。
        Hmap redis在存储的时候如果数据比较少（看文章是512）会使用ziplist，测试了下，在ziplist存储的状态下，会都返回，count不生效
        如果数据超过512之后会使用hmap来存储，这时基本就是准确的了。所以这个函数只能保证返回 >= count
*/
func (r *RedisCluster) HScan(key string, count int64) (map[string]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HScan", st, err, key)
	}()

	ret, _, err := clusterClient.HScan(key, 0, "", count).Result()
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
func (r *RedisCluster) HGetAll(key string) (map[string]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HGetAll", st, err, key)
	}()

	ret, err := clusterClient.HGetAll(key).Result()
	return ret, err
}

/**
1. 若异常, 返回 (false, error)
2. 若正常, 返回 (bool, nil), 其中true: field在hash中不存在且新增成功; false: field已在hash中存在且更新成功.
*/
func (r *RedisCluster) HSet(key string, field string, value string) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HSet", st, err, key)
	}()

	ret, err := clusterClient.HSet(key, field, value).Result()
	return ret, err
}

/**
1. 若异常, 返回 (false, error)
2. 若正常, 返回 (bool, nil), 其中true: field在hash中不存在且新增成功; false: field已在hash中存在, 不做更新.
*/
func (r *RedisCluster) HSetNx(key string, field string, value string) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HSetNx", st, err, key)
	}()

	ret, err := clusterClient.HSetNX(key, field, value).Result()
	return ret, err
}

/**
1. 若异常, 返回 (false, error)
2. 若正常, 返回 (true, nil)
*/
func (r *RedisCluster) HMSet(key string, fields map[string]string) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HMSet", st, err, key)
	}()

	if fields == nil || len(fields) == 0 {
		return true, nil
	}

	tmpFields := make(map[string]interface{}, len(fields))
	for fieldName, fieldValue := range fields {
		tmpFields[fieldName] = fieldValue
	}

	_, err = clusterClient.HMSet(key, tmpFields).Result()
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

func (r *RedisCluster) HMDel(key string, fields []string) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HMDel", st, err, key)
	}()

	return clusterClient.HDel(key, fields...).Result()
}

/**
1. 若正常, 返回 (true,nil)
2. 若异常, 返回 (false,error)
  key不存在返回(false,nil)
*/
func (r *RedisCluster) Expire(key string, expiration time.Duration) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "Expire", st, err, key)
	}()

	return clusterClient.Expire(key, expiration).Result()
}

/**
1. 若正常, 返回 (true,nil)
2. 若异常, 返回 (false,error)
  key不存在返回(false,nil)
*/
func (r *RedisCluster) PExpire(key string, expiration time.Duration) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "PExpire", st, err, key)
	}()

	return clusterClient.PExpire(key, expiration).Result()
}

/**
1. 若正常, 返回 (true,nil)
2. 若异常, 返回 (false,error)
  key不存在返回(false,nil)
*/
func (r *RedisCluster) PTtl(key string) (time.Duration, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "PTtl", st, err, key)
	}()

	return clusterClient.PTTL(key).Result()
}

/**
1. 若设置成功, 返回(int64,nil)
2. 若异常, 返回(-1,error)
*/
func (r *RedisCluster) HIncrBy(key string, field string, value int64) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HIncrBy", st, err, key)
	}()

	ret, err := clusterClient.HIncrBy(key, field, value).Result()
	return ret, err
}

func (r *RedisCluster) IncrBy(key string, value int64) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "IncrBy", st, err, key)
	}()

	ret, err := clusterClient.IncrBy(key, value).Result()
	return ret, err
}

func (r *RedisCluster) Eval(script string, keys []string, args []interface{}) (interface{}, error) {
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
		reportPerf(r.clusterName, "Eval", st, err, keys)
	}()

	ret, err = clusterClient.Eval(script, keys, args...).Result()
	return ret, err
}

func (r *RedisCluster) EvalSha(scriptSha string, keys []string, args []interface{}) (interface{}, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "IncrBy", st, err, keys)
	}()

	ret, err := clusterClient.EvalSha(scriptSha, keys, args...).Result()
	return ret, err
}

func getScriptSha(script string) string {
	hash := sha1.New()
	io.WriteString(hash, script)
	return hex.EncodeToString(hash.Sum(nil))
}

func (r *RedisCluster) ZAdd(key string, members []redis.Z) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return -1, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZAdd", st, err, key)
	}()

	return clusterClient.ZAdd(key, members...).Result()
}

type ZSetResult struct {
	Member string
	Score  float64
}

func (r *RedisCluster) ZRangeByScoreWithScores(key string, min, max string, offset, limit int64) ([]*ZSetResult, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZRangeByScoreWithScores", st, err, key)
	}()

	opt := redis.ZRangeBy{
		Count:  limit,
		Max:    max,
		Min:    min,
		Offset: offset,
	}
	result, err := clusterClient.ZRangeByScoreWithScores(key, opt).Result()
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

func (r *RedisCluster) ZRevRangeByScoreWithScores(key string, min, max string, offset, limit int64) ([]*ZSetResult, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZRevRangeByScoreWithScores", st, err, key)
	}()

	opt := redis.ZRangeBy{
		Count:  limit,
		Max:    max,
		Min:    min,
		Offset: offset,
	}
	log.Debug("ZRevRangeByScoreWithScores key=%v,opt=%v", key, opt)
	result, err := clusterClient.ZRevRangeByScoreWithScores(key, opt).Result()
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

func (r *RedisCluster) ZRange(key string, start, stop int64) ([]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZRange", st, err, key)
	}()

	return clusterClient.ZRange(key, start, stop).Result()
}

func (r *RedisCluster) ZRemRangeByRank(key string, start, stop int64) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZRemRangeByRank", st, err, key)
	}()

	return clusterClient.ZRemRangeByRank(key, start, stop).Result()
}

func (r *RedisCluster) ZRemRangeByScore(key string, min, max string) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZRemRangeByScore", st, err, key)
	}()

	return clusterClient.ZRemRangeByScore(key, min, max).Result()
}

func (r *RedisCluster) ZRem(key string, members ...interface{}) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZRem", st, err, key)
	}()

	return clusterClient.ZRem(key, members...).Result()
}

func (r *RedisCluster) ZRevRange(key string, start, stop int64) ([]string, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return nil, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "ZRevRange", st, err, key)
	}()

	return clusterClient.ZRevRange(key, start, stop).Result()
}

func (r *RedisCluster) SetBit(key string, offset int64, value int) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "SetBit", st, err, key)
	}()

	return clusterClient.SetBit(key, offset, value).Result()
}

func (r *RedisCluster) Exist(key string) (bool, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return false, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "Exist", st, err, key)
	}()

	var ret bool
	var existRet int64
	existRet, err = clusterClient.Exists(key).Result()
	if err != nil {
		ret = false
	} else if existRet > 0 {
		ret = true
	} else {
		ret = false
	}

	return ret, err
}

func (r *RedisCluster) SAdd(key string, members []interface{}) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "SAdd", st, err, key)
	}()

	return clusterClient.SAdd(key, members...).Result()

}

func (r *RedisCluster) SPop(key string) (string, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "SPop", st, err, key)
	}()

	return clusterClient.SPop(key).Result()
}

func (r *RedisCluster) LPop(key string) (string, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "LPop", st, err, key)
	}()

	return clusterClient.LPop(key).Result()
}

func (r *RedisCluster) LIndex(key string, index int64) (string, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return "", RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "LIndex", st, err, key)
	}()

	return clusterClient.LIndex(key, index).Result()
}

func (r *RedisCluster) LPush(key string, value string) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "LPush", st, err, key)
	}()

	return clusterClient.LPush(key, value).Result()
}

func (r *RedisCluster) RPush(key string, value string) (int64, error) {
	clusterClient := r.getClusterClient()
	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "RPush", st, err, key)
	}()

	return clusterClient.RPush(key, value).Result()
}

func (r *RedisCluster) HLen(key string) (int64, error) {
	clusterClient := r.getClusterClient()

	if clusterClient == nil {
		return 0, RedisClusterCustomError("redis cluster not init")
	}

	var err error
	st := time.Now()
	defer func() {
		reportPerf(r.clusterName, "HLen", st, err, key)
	}()

	return clusterClient.HLen(key).Result()
}
