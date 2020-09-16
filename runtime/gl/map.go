/**
 * Copyright 2020 gl Author. All rights reserved.
 * Author: Chuck1024
 */

package gl

import "sync"

type Shard struct {
	data map[string]interface{}
	sync.RWMutex
}

type ConcurrentMap struct {
	s          []*Shard
	shardCount int
}

func NewShard(shardCount int) *ConcurrentMap {
	if shardCount <= 0 {
		shardCount = 32
	}
	cache := make([]*Shard, shardCount)
	for i := 0; i < shardCount; i++ {
		cache[i] = &Shard{data: make(map[string]interface{})}
	}
	m := &ConcurrentMap{
		s:          cache,
		shardCount: shardCount,
	}
	return m
}

func (m *ConcurrentMap) GetShard(key string) *Shard {
	return m.s[uint(fnv32(key))%uint(m.shardCount)]
}

func (m *ConcurrentMap) Set(key string, value interface{}) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	shard.data[key] = value
	shard.Unlock()
}

func (m *ConcurrentMap) Get(key string) (interface{}, bool) {
	shard := m.GetShard(key)
	shard.RLock()
	val, ok := shard.data[key]
	shard.RUnlock()
	return val, ok
}

func (m *ConcurrentMap) Remove(key string) {
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.data, key)
	shard.Unlock()
}

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	for i := 0; i < len(key); i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
