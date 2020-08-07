package utls

import (
	"fmt"
	"sync"
	"time"
)

type FixedGoroutinePool struct {
	Size          int64
	semaphoreChan chan bool
	wg            sync.WaitGroup
}

func (f *FixedGoroutinePool) Start() error {
	f.semaphoreChan = make(chan bool, f.Size)
	return nil
}

func (f *FixedGoroutinePool) Execute(function func()) {
	f.semaphoreChan <- true
	f.wg.Add(1)
	go func() {
		defer func() {
			<-f.semaphoreChan
			f.wg.Done()
		}()
		function()
	}()
}

func (f *FixedGoroutinePool) ExecuteWithArg(function func(args ...interface{}), arg ...interface{}) {
	f.semaphoreChan <- true
	f.wg.Add(1)
	go func() {
		defer func() {
			<-f.semaphoreChan
			f.wg.Done()
		}()
		function(arg...)
	}()
}

func (f *FixedGoroutinePool) Close() {
	close(f.semaphoreChan)
	f.wg.Wait()
}

var TIMTOUR_Err = fmt.Errorf("insert into gouroutine pool timeout")

type FixedGoroutinePoolTimeout struct {
	Size          int64
	Timeout       time.Duration
	semaphoreChan chan bool
	wg            sync.WaitGroup
}

func (f *FixedGoroutinePoolTimeout) Start() error {
	f.semaphoreChan = make(chan bool, f.Size)
	return nil
}

func (f *FixedGoroutinePoolTimeout) Execute(function func()) error {
	if f.Timeout > 0 {
		select {
		case f.semaphoreChan <- true:
		case <-time.After(f.Timeout):
			return TIMTOUR_Err
		}

	} else {
		f.semaphoreChan <- true
	}
	f.wg.Add(1)
	go func() {
		defer func() {
			<-f.semaphoreChan
			f.wg.Done()
		}()
		function()
	}()
	return nil
}

func (f *FixedGoroutinePoolTimeout) Close() {
	close(f.semaphoreChan)
	f.wg.Wait()
}

type GoRoutinePoolWithConfig struct {
	DefaultSize    int64
	ReservedConfig map[string]int64
	reservedPools  map[string]*FixedGoroutinePool
	defaultPool    *FixedGoroutinePool
}

func (g *GoRoutinePoolWithConfig) Start() error {
	g.reservedPools = make(map[string]*FixedGoroutinePool)

	if g.ReservedConfig != nil {
		for k, v := range g.ReservedConfig {
			p := &FixedGoroutinePool{Size: v}
			if err := p.Start(); err != nil {
				return err
			}
			g.reservedPools[k] = p
		}
	}

	if g.DefaultSize < 1 {
		return fmt.Errorf("invalid pool default size %v", g.DefaultSize)
	}

	g.defaultPool = &FixedGoroutinePool{Size: g.DefaultSize}
	g.defaultPool.Start()

	return nil
}

func (g *GoRoutinePoolWithConfig) ExecuteDefault(function func()) {
	g.defaultPool.Execute(function)
}

func (g *GoRoutinePoolWithConfig) ExecuteKey(key string, function func()) {
	pool := g.reservedPools[key]
	if pool != nil {
		pool.Execute(function)
		return
	}
	g.ExecuteDefault(function)
}

func (g *GoRoutinePoolWithConfig) Close() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.defaultPool.Close()
	}()

	for _, p := range g.reservedPools {
		wg.Add(1)
		pool := p
		go func() {
			defer wg.Done()
			pool.Close()
		}()
	}
	wg.Wait()
}
