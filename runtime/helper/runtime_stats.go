/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package helper

import (
	"encoding/json"
	"fmt"
	"github.com/chuck1024/gd/utls"
	"runtime"
	"time"
)

func nsToMs(ns uint64) float64 {
	return float64(ns) / float64(time.Millisecond/time.Nanosecond)
}

func gcPause(m1, m2 *runtime.MemStats) (p []uint64) {
	gc := m2.NumGC - m1.NumGC
	if gc > 0 {
		for i := uint32(0); i < gc; i++ {
			p = append(p, m2.PauseNs[(m2.NumGC-i+255)%256])
		}
	}
	return
}

func calcMaxAvg(a []uint64) (uint64, uint64) {
	if len(a) == 0 {
		return 0, 0
	}
	var max, avg uint64
	for _, u := range a {
		if u > max {
			max = u
		}
		avg += u
	}
	return max, avg / uint64(len(a))
}

func formatGCPause(m1, m2 *runtime.MemStats) string {
	p := gcPause(m1, m2)
	max, avg := calcMaxAvg(p)
	return fmt.Sprintf("%.2f+%.2fms", nsToMs(max), nsToMs(avg))
}

func FormatMemStats(m1, m2 *runtime.MemStats) string {
	return fmt.Sprintf("Goroutines: %d, GCPause: %s, HeapObjects: %d, HeapAlloc: %s, Mallocs: %d",
		runtime.NumGoroutine(), formatGCPause(m1, m2), m2.HeapObjects, utls.HumanSize(m2.HeapAlloc), m2.Mallocs)
}

// get runtime stats,contained goroutine numbers,gc pause,heap objects,heap alloc and malloc
func RuntimeStats(d time.Duration, f func(string)) {
	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)
	for {
		time.Sleep(d)
		runtime.ReadMemStats(&m2)
		s := FormatMemStats(&m1, &m2)
		if f == nil {
			fmt.Println(s)
		} else {
			f(s)
		}
		m1 = m2
	}
}

func GcStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// in ms
	lastGC := time.Unix(int64(m.LastGC/1000/1000/1000), 0)
	pauseNs := m.PauseNs
	numGC := m.NumGC
	lastPause := int64(pauseNs[(numGC+255)%256] / 1000 / 1000)
	gcStats := fmt.Sprintf("%dms,%s", lastPause, lastGC)
	return gcStats
}

func MemStats2() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// in mb
	memStats, _ := json.Marshal(m)
	return string(memStats)
}

func MemStats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// in mb
	alloc := int(m.Alloc / 1024 / 1024)
	sys := int(m.Sys / 1024 / 1024)
	heapSys := int(m.HeapSys / 1024 / 1024)
	heapAlloc := int(m.HeapAlloc / 1024 / 1024)
	heapIdle := int(m.HeapIdle / 1024 / 1024)
	heapReleased := int(m.HeapReleased / 1024 / 1024)
	memStats := fmt.Sprintf("%d,%d,%d,%d,%d,%d", sys, alloc, heapSys, heapAlloc, heapIdle, heapReleased)
	return memStats
}
