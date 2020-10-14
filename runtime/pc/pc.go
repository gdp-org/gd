/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package pc

import (
	"bytes"
	"fmt"
	js "github.com/bitly/go-simplejson"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/utls"
	cMap "github.com/orcaman/concurrent-map"
	"github.com/rcrowley/go-metrics"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultShardCount    = 1024
	DefaultSendCountOnce = 35

	defaultWorkerCount      = 512000
	defaultCostHandlerCount = 2

	GoProjectsGoroutineNum = "go_projects_goroutine_num"
)

var (
	closeOnce         sync.Once
	initOnce          sync.Once
	stop              = make(chan bool)
	kMap              cMap.ConcurrentMap
	updaterLock       sync.RWMutex
	_updater          updater
	suffixDeciderLock sync.RWMutex
	suffixDecider     DecideSuffix
	costTimerC        chan *costTimer
	perfCounterC      chan *pcReq
	closed            int32

	mRegistry = metrics.NewRegistry()

	falconAgentUrl = "http://127.0.0.1:1988/v1/push"
)

type DecideSuffix func(key string) string
type updater func() map[string]*int64

type pcReq struct {
	Key string
	Val int64
}

type costTimer struct {
	Name string
	Cost time.Duration
}

func init() {
	cMap.SHARD_COUNT = defaultShardCount
	kMap = cMap.New()
}

func Init() {
	InitPerfCounter("", nil, []string{})
}

func InitPerfCounter(tar string, upd updater, initKeys []string) {
	decideSuffix := func(key string) string {
		return ""
	}
	initOnce.Do(func() {
		initPerfCounter(tar, upd, initKeys, decideSuffix, defaultWorkerCount, defaultCostHandlerCount)
		atomic.AddInt32(&closed, 1)
	})
	SetSuffixDecider(decideSuffix)
	SetUpdater(upd)
}

func initPerfCounter(tar string, upd updater, initKeys []string, decideSuffix DecideSuffix, workerCount, costHandlerCount int) {
	if workerCount < defaultWorkerCount {
		workerCount = defaultWorkerCount
	}

	if costHandlerCount <= defaultCostHandlerCount {
		costHandlerCount = defaultCostHandlerCount
	}

	costTimerC = make(chan *costTimer, workerCount)
	perfCounterC = make(chan *pcReq, workerCount)

	if tar != "" {
		falconAgentUrl = tar
	}

	updaterLock.Lock()
	_updater = upd
	updaterLock.Unlock()

	for _, k := range initKeys {
		tmp := new(int64)
		kMap.Set(k, tmp)
	}

	suffixDeciderLock.Lock()
	suffixDecider = decideSuffix
	suffixDeciderLock.Unlock()

	go handlePc()
	for i := 0; i < costHandlerCount; i++ {
		go handleCostsNoPool()
	}
	go walk()
}

func SetRunPort(runPort int) {
	SetSuffixDecider(
		func(key string) string {
			return fmt.Sprintf("-%d", runPort)
		},
	)
}

func SetSuffixDecider(d DecideSuffix) {
	suffixDeciderLock.Lock()
	defer suffixDeciderLock.Unlock()
	suffixDecider = d
}

func SetUpdater(upd updater) {
	updaterLock.Lock()
	defer updaterLock.Unlock()
	_updater = upd
}

func handlePc() {
	for {
		select {
		case c, isOpen := <-perfCounterC:
			if !isOpen {
				return
			}
			_, ok := kMap.Get(c.Key)
			if !ok {
				tmp := new(int64)
				kMap.Set(c.Key, tmp)
			}
			incr(c.Key, c.Val)
		}
	}
}

func handleCostsNoPool() {
	for {
		select {
		case c, isOpen := <-costTimerC:
			if !isOpen {
				return
			}
			t := metrics.GetOrRegisterTimer(c.Name, mRegistry)
			t.Update(c.Cost)
		}
	}
}

func walk() {
	tc := time.NewTicker(60 * time.Second)
	defer tc.Stop()
	for {
		select {
		case <-tc.C:
			report()
			continue
		case <-stop:
			return
		}
	}
}

func report() {
	//costs from metrics
	snapMaps := make(map[string]metrics.Timer)
	timerMap := make(map[string]*int64)
	mRegistry.Each(func(name string, ifc interface{}) {
		if name == "" {
			return
		}

		timer, ok := ifc.(metrics.Timer)
		if !ok {
			return
		}
		snap := timer.Snapshot()
		snapMaps[name] = snap
	})

	//defer mRegistry.UnregisterAll()
	for name, snap := range snapMaps {
		count := int64(snap.Rate1() * 60)
		skipP99 := false
		if count < 30 {
			//not worth reporting
			skipP99 = true
		}

		floats := snap.Percentiles([]float64{0.95, 0.99, 0.995, 0.999})
		p95Float := floats[0]
		p99Float := floats[1]
		p995Float := floats[2]
		p999Float := floats[3]

		p95Ms := int64(p95Float) / int64(time.Millisecond)
		p99Ms := int64(p99Float) / int64(time.Millisecond)
		p995Ms := int64(p995Float) / int64(time.Millisecond)
		p999Ms := int64(p999Float) / int64(time.Millisecond)
		timerMap[name+",sum=count"] = &count

		if !skipP99 {
			timerMap[name+",sum=cost_ms_p95"] = &p95Ms
			timerMap[name+",sum=cost_ms_p99"] = &p99Ms
			timerMap[name+",sum=cost_ms_p995"] = &p995Ms
			timerMap[name+",sum=cost_ms_p999"] = &p999Ms
		} else {
			timerMap[name+",sum=skipCost_ms_p95"] = &p95Ms
			timerMap[name+",sum=skipCost_ms_p99"] = &p99Ms
			timerMap[name+",sum=skipCost_ms_p995"] = &p995Ms
			timerMap[name+",sum=skipCost_ms_p999"] = &p999Ms
		}

		failCount, ok := kMap.Get(name + ",sum=fail")
		if ok {
			fCip, ok := failCount.(*int64)
			if ok && fCip != nil {
				fci := atomic.LoadInt64(fCip)
				if count > 0 {
					failRate := (float64(fci) / float64(count)) * float64(10000)
					failRateInt := int64(failRate)
					if !skipP99 {
						timerMap[name+",sum=failRate"] = &failRateInt
					} else {
						timerMap[name+",sum=skipFailRate"] = &failRateInt
					}
				}
			}
		}
	}

	for tk, tv := range timerMap {
		kMap.Set(tk, tv)
	}

	// perf counters from updaters
	var upds updater
	updaterLock.RLock()
	upds = _updater
	updaterLock.RUnlock()
	if upds != nil {
		ups := _updater()
		if ups != nil {
			for k, v := range ups {
				kMap.Set(k, v)
			}
		}
	}

	goNums := int64(runtime.NumGoroutine())
	kMap.Set(GoProjectsGoroutineNum, &goNums)

	suffixDeciderLock.RLock()
	sufDecider := suffixDecider
	suffixDeciderLock.RUnlock()

	hn, err := os.Hostname()
	if err != nil {
		dlog.Error("get host name fail! no report send")
		return
	}

	// -1 means go process
	// hn = hn + "-1"
	var send []*js.Json
	ct := time.Now().Unix()
	runPort := -1
	if sufDecider != nil {
		suffix := sufDecider("")
		if strings.HasPrefix(suffix, "-") {
			rp, err := strconv.Atoi(suffix[1:])
			if err == nil && rp > 0 {
				runPort = rp
			}
		}
	}
	for ele := range kMap.IterBuffered() {
		k := ele.Key
		pv, ok := ele.Val.(*int64)
		if !ok {
			dlog.Error("pc val type invalid,k=%s,v=%v", k, ele.Val)
			continue
		}
		finalHostName := hn

		v := atomic.SwapInt64(pv, 0)
		j := js.New()
		j.Set("metric", "gd")
		j.Set("endpoint", finalHostName)
		j.Set("timestamp", ct)
		j.Set("step", 60)
		j.Set("value", v)
		j.Set("counterType", "GAUGE")
		if runPort > 0 {
			k = fmt.Sprintf("%s,port=%d", k, runPort)
		}
		j.Set("tags", "attr="+k)
		send = append(send, j)
	}

	reports, err := utls.SliceCutter(send[:], DefaultSendCountOnce)
	if err != nil {
		dlog.Error("cut report fail,send=%v,err=%s", send, err)
		return
	}
	for _, v := range reports {
		go sendPcReport(v)
	}
}

func sendPcReport(send interface{}) {
	paramsBytes, err := utls.Marshal(send)
	if err != nil {
		dlog.Error("json marshal fail,send=%v,err=%s", send, err)
		return
	}

	req, _ := http.NewRequest("POST", falconAgentUrl, bytes.NewReader(paramsBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	pcReportSt := time.Now()
	resp, err := client.Do(req)
	pcReportCost := time.Now().Sub(pcReportSt) / time.Millisecond

	var bodyStr string
	var httpStatus int

	if resp != nil {
		httpStatus = resp.StatusCode
		body := make([]byte, resp.ContentLength)
		n, _ := resp.Body.Read(body)
		defer resp.Body.Close()
		if n > 0 {
			bodyStr = string(body[:n])
		}
	}

	if err != nil {
		dlog.Error("report perfCounter fail,params=%s,status=%d,body=%s,err=%s,reportCost=%d", string(paramsBytes), httpStatus, bodyStr, err, pcReportCost)
	} else {
		dlog.Debug("report perfCounter ok,params=%s,status=%d,body=%s,reportCost=%d", string(paramsBytes), httpStatus, bodyStr, pcReportCost)
	}
}

func ClosePerfCounter() {
	closeOnce.Do(func() {
		atomic.StoreInt32(&closed, 0)
		close(stop)
	})
}

func CostFail(key string, val int64) {
	key = key + ",sum=fail"
	Incr(key, val)
}

func ErrorIncr(key string, val int64) {
	key = key + ",type=exceptionReport"
	Incr(key, val)
}

func Incr(key string, val int64) {
	req := &pcReq{
		Key: key,
		Val: val,
	}

	if atomic.LoadInt32(&closed) == 0 {
		return
	}

	select {
	case perfCounterC <- req:
	default:
		utls.FatalWithSmsAlert(fmt.Sprintf("perfcounter chan full, drop kv=%v", req))
	}
}

func incr(key string, val int64) {
	cv, ok := kMap.Get(key)
	if !ok || cv == nil {
		dlog.Warn("key %s not inited!", key)
		return
	}
	vl, ok := cv.(*int64)
	if !ok {
		dlog.Warn("key %s not int64!", key)
		return
	}
	atomic.AddInt64(vl, val)
}

// in ms
func Cost(name string, cost time.Duration) {
	r := &costTimer{
		Name: name,
		Cost: cost,
	}
	if atomic.LoadInt32(&closed) == 0 {
		return
	}
	select {
	case costTimerC <- r:
	default:
		utls.FatalWithSmsAlert(fmt.Sprintf("cost timer chan full with len=%v, name=%s, drop cost=%v", len(costTimerC), name, cost))
	}
}
