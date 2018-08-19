/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package cache

import (
	"fmt"
	"github.com/chasex/redis-go-cluster"
)

const (
	incr             = "incr"
	decr             = "decr"
	incrBy           = "incrBy"
	decrBy           = "decrBy"
	get              = "get"
	set              = "set"
	setEx            = "setEx"
	setNx            = "setNx"
	del              = "del"
	expire           = "expire"
	rPop             = "rPop"
	rPush            = "rPush"
	lPop             = "lPop"
	lPush            = "lPush"
	lLen             = "lLen"
	lRange           = "lRange"
	lTrim            = "lTrim"
	sRem             = "sRem"
	sAdd             = "sAdd"
	sPop             = "sPop"
	sCARD            = "sCard"
	sIsMembers       = "sIsMember"
	sMembers         = "sMembers"
	hDel             = "hDel"
	hGet             = "hGet"
	hGetAll          = "hGetAll"
	hSet             = "hSet"
	hIncrBy          = "hIncrBy"
	sCan             = "sCan"
	zAdd             = "zAdd"
	zRange           = "zRange"
	zRevRange        = "zRevRange"
	zRemRangeByScore = "zRemRangeByScore"
	zScore           = "zScore"
	zCard            = "zCard"
	zRem             = "zRem"
	zRemRangeByRank  = "zRemRangeByRank"
)

func Incr(key string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(incr, key))
}

func Decr(key string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(decr, key))
}

func IncrBy(key string, add int) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(incrBy, key, add))
}

func DecrBy(key string, sub int) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(decrBy, key, sub))
}

func Get(key string) (string, error) {
	redisCluster := GetClient()
	return redis.String(redisCluster.Do(get, key))
}

func Set(key string, value string) error {
	redisCluster := GetClient()

	_, err := redis.String(redisCluster.Do(set, key, value))

	return err
}

func SetEx(key string, second int, value string) error {
	redisCluster := GetClient()

	_, err := redis.String(redisCluster.Do(setEx, key, second, value))

	return err
}

func SetNx(key string, value string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(setNx, key, value))
}

func Del(key string) (int64, error) {
	redisCluster := GetClient()
	return redis.Int64(redisCluster.Do(del, key))
}

func Expire(key string, second int) (int, error) {
	redisCluster := GetClient()
	secondStr := fmt.Sprintf("%d", second)
	return redis.Int(redisCluster.Do(expire, key, secondStr))
}

func RPop(key string) (string, error) {
	redisCluster := GetClient()
	return redis.String(redisCluster.Do(rPop, key))
}

func RPush(key string, value string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(rPush, key, value))
}

func LPush(key string, value string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(lPush, key, value))
}

func LPop(key string) (string, error) {
	redisCluster := GetClient()
	return redis.String(redisCluster.Do(lPop, key))
}

func LLen(key string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(lLen, key))
}

func LTrim(key string, start, stop int) error {
	redisCluster := GetClient()
	_, err := redisCluster.Do(lTrim, key, start, stop)
	return err
}

func LRange(key string, begin int, end int) ([]string, error) {
	redisCluster := GetClient()
	return redis.Strings(redisCluster.Do(lRange, key, begin, end))
}

func SAdd(key string, value string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(sAdd, key, value))
}

func SPop(key string) (string, error) {
	redisCluster := GetClient()
	return redis.String(redisCluster.Do(sPop, key))
}

func SRem(key string, value string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(sRem, key, value))
}

func SCard(key string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(sCARD, key))
}

func SIsMember(key, value string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(sIsMembers, key, value))
}

func SMembers(key string) ([]string, error) {
	redisCluster := GetClient()
	return redis.Strings(redisCluster.Do(sMembers, key))
}

func HDel(key string, hashKey ...string) (int, error) {
	redisCluster := GetClient()

	args := make([]interface{}, 0)
	args = append(args, key)
	for _, v := range hashKey {
		args = append(args, v)
	}

	return redis.Int(redisCluster.Do(hDel, args...))
}

func HGet(key string, hashKey string) (string, error) {
	redisCluster := GetClient()
	return redis.String(redisCluster.Do(hGet, key, hashKey))
}

func HGetAll(key string) (map[string]string, error) {
	redisCluster := GetClient()
	return redis.StringMap(redisCluster.Do(hGetAll, key))
}

func HSet(key string, hashKey string, value string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(hSet, key, hashKey, value))
}

func HIncrBy(key string, hashKey string, add int) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(hIncrBy, key, hashKey, add))
}

func ZAdd(key string, score int64, value string) error {
	redisCluster := GetClient()
	_, err := redis.Int(redisCluster.Do(zAdd, key, score, value))
	return err
}

func ZRange(key string, begin int, end int) ([]string, error) {
	redisCluster := GetClient()
	return redis.Strings(redisCluster.Do(zRange, key, begin, end))
}

func ZRevRange(key string, begin int, end int) ([]string, error) {
	redisCluster := GetClient()
	return redis.Strings(redisCluster.Do(zRevRange, key, begin, end))
}

func ZCard(key string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(zCard, key))
}

func ZRem(key string, member string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(zRem, key, member))
}

func ZScore(key string, member string) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(zScore, key, member))
}

func ZRemRangeByScore(key string, scoreBegin int, scoreEnd int) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(zRemRangeByScore, key, scoreBegin, scoreEnd))
}

func ZRemRangeByRank(key string, begin int, end int) (int, error) {
	redisCluster := GetClient()
	return redis.Int(redisCluster.Do(zRemRangeByRank, key, begin, end))
}
