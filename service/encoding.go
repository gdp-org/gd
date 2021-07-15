/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Chuck1024
 */

package service

type NodeInfo interface {
	GetIp() string
	GetPort() int
	GetOffline() bool
	GetWeight() uint64
}

type DefaultNodeInfo struct {
	Ip      string `json:"ip"`
	Port    int    `json:"port"`
	Offline bool   `json:"offline"`
	Weight  uint64 `json:"weight"`
}

func (d *DefaultNodeInfo) GetIp() string {
	return d.Ip
}

func (d *DefaultNodeInfo) GetPort() int {
	return d.Port
}

func (d *DefaultNodeInfo) GetOffline() bool {
	return d.Offline
}

func (d *DefaultNodeInfo) GetWeight() uint64 {
	return d.Weight
}
