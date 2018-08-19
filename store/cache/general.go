/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package cache

import (
	redisCluster "github.com/chasex/redis-go-cluster"
	"github.com/garyburd/redigo/redis"
	"sync"
	"time"
)

type RedisHandle interface {
	Get() RedisHandle
	Close()
	Do(cmd string, args ...interface{}) (interface{}, error)
}

type RedisPool struct {
	pool  *redis.Pool
	mutex sync.Mutex
}

func (c *RedisPool) Get() RedisHandle {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for {
		r := c.pool.Get()
		if r.Err() != nil {
			return nil
		}
		return &Connector{Pool: c, Conn: r}
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

func (c *RedisPool) Close() {

}

func (c *RedisPool) Do(cmd string, args ...interface{}) (interface{}, error) {
	return nil,nil
}

type Connector struct {
	Pool *RedisPool
	Conn redis.Conn
}

func (c *Connector) Get() RedisHandle {
	return c
}

func (c *Connector) Close() {
	c.Pool.mutex.Lock()
	defer c.Pool.mutex.Unlock()

	c.Conn.Close()
}

func (c *Connector) Do(cmd string, args ...interface{}) (interface{}, error) {
	return c.Conn.Do(cmd, args...)
}

type ClusterClient struct {
	cluster *redisCluster.Cluster
}

func (c *ClusterClient) Get() RedisHandle {
	return c
}

func (c *ClusterClient) Close() {
	c.cluster.Close()
}

func (c *ClusterClient) Do(cmd string, args ...interface{}) (interface{}, error) {
	return c.cluster.Do(cmd, args...)
}

type RedisClient struct {
	client RedisHandle
	config *RedisConfig
}

func (c *RedisClient) Get() RedisHandle {
	return c.client.Get()
}

func (c *RedisClient) Close() {
	if c.config.Cluster {
		return
	}
	c.client.Close()
}

func (c *RedisClient) Do(cmd string, args ...interface{}) (interface{}, error) {
	return c.client.Do(cmd, args...)
}
