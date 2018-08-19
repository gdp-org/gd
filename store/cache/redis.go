/**
 * Copyright 2018 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package cache

import (
	redisCluster "github.com/chasex/redis-go-cluster"
	"github.com/garyburd/redigo/redis"
	"github.com/xuyu/logging"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultNetwork   = "tcp"
	DefaultAddress   = ":6379"
	DefaultCluster   = false
	DefaultDatabases = 0
	DefaultTimeout   = 15 * time.Second
	DefaultMaxIdle   = 1
	DefaultMaxActive = -1
)

var (
	client *RedisClient
)

type RedisConfig struct {
	Network     string
	Database    int
	Cluster     bool
	Host        string
	Password    string
	MaxIdle     int
	MaxActive   int
	IdleTimeout time.Duration
}

func redisConfigFromURLString(rawUrl string) (*RedisConfig, error) {
	ul, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}

	scheme := DefaultNetwork
	if ul.Scheme != "" {
		scheme = ul.Scheme
	}

	host := DefaultAddress
	if ul.Host != "" {
		host = ul.Host
	}

	password := ""
	if ul.User != nil {
		if pw, set := ul.User.Password(); set {
			password = pw
		}
	}

	db := DefaultDatabases
	path := strings.Trim(ul.Path, "/")
	if path != "" {
		db, err = strconv.Atoi(path)
		if err != nil {
			return nil, err
		}
	}

	cluster := DefaultCluster
	if ul.Query().Get("cluster") != "" {
		cluster, err = strconv.ParseBool(ul.Query().Get("cluster"))
		if err != nil {
			return nil, err
		}
	}

	timeout := DefaultTimeout
	if ul.Query().Get("idleTimeout") != "" {
		timeout, err = time.ParseDuration(ul.Query().Get("idleTimeout"))
		if err != nil {
			return nil, err
		}
	}

	maxIdle := DefaultMaxIdle
	if ul.Query().Get("maxIdle") != "" {
		maxIdle, err = strconv.Atoi(ul.Query().Get("maxIdle"))
		if err != nil {
			return nil, err
		}
	}

	maxActive := DefaultMaxActive
	if ul.Query().Get("maxActive") != "" {
		maxActive, err = strconv.Atoi(ul.Query().Get("maxActive"))
		if err != nil {
			return nil, err
		}
	}

	return &RedisConfig{
		Network:     scheme,
		Database:    db,
		Cluster:     cluster,
		Host:        host,
		Password:    password,
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: timeout,
	}, nil
}

func Init(URL string) {
	client = &RedisClient{}
	redisConfig, err := redisConfigFromURLString(URL)
	if err != nil {
		logging.Error("[DialURL] redisConfigFromURLString occur error: %s", err)
		return
	}

	if redisConfig.Cluster {
		client.client = initCluster(redisConfig)
	} else {
		client.client = initPool(redisConfig)
	}
}

func GetClient() RedisHandle {
	return client.client
}

func initPool(conf *RedisConfig) *RedisPool {
	return &RedisPool{
		pool: initRedisPools(conf.Host, conf),
	}
}

func initRedisPools(host string, conf *RedisConfig) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     conf.MaxIdle,
		MaxActive:   conf.MaxActive,
		IdleTimeout: conf.IdleTimeout * time.Second,
		Dial: func() (redis.Conn, error) {
			c, e := redis.Dial(conf.Network, host)
			if e != nil {
				return nil, e
			}
			if conf.Password != "" {
				if _, e := c.Do("AUTH", conf.Password); e != nil {
					c.Close()
					return nil, e
				}
			}
			c.Do("SELECT", conf.Database)
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, e := c.Do("PING")
			return e
		},
	}
}

func initCluster(conf *RedisConfig) *ClusterClient {
	host := strings.Split(conf.Host, ",")

	c, err := redisCluster.NewCluster(
		&redisCluster.Options{
			StartNodes:   host,
			ConnTimeout:  500 * time.Millisecond,
			ReadTimeout:  500 * time.Millisecond,
			WriteTimeout: 500 * time.Millisecond,
			KeepAlive:    conf.MaxActive,
			AliveTime:    conf.IdleTimeout * time.Second,
		})

	if err != nil {
		logging.Error("[initCluster] occur error: ", err)
		return nil
	}

	return &ClusterClient{cluster: c}
}
