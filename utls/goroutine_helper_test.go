package utls

import (
	"sync"
	"testing"

	"sync/atomic"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGoRoutinePoolWithConfig_ExecuteKey(t *testing.T) {
	Convey("test execute key", t, func() {
		g := &GoRoutinePoolWithConfig{DefaultSize: 10, ReservedConfig: map[string]int64{
			"test1": 5,
			"test2": 5,
		}}
		err := g.Start()
		So(err, ShouldBeNil)
		counter := int64(0)
		g.ExecuteKey("test", func() {
			atomic.AddInt64(&counter, 1)
		})
		g.ExecuteKey("test1", func() {
			atomic.AddInt64(&counter, 2)
		})
		g.ExecuteKey("test2", func() {
			atomic.AddInt64(&counter, 3)
		})
		g.Close()

		So(counter, ShouldEqual, 6)
	})
	Convey("test default", t, func() {
		g := &GoRoutinePoolWithConfig{DefaultSize: 10, ReservedConfig: map[string]int64{
			"test1": 5,
			"test2": 5,
		}}
		err := g.Start()
		So(err, ShouldBeNil)
		counter := int64(0)
		g.ExecuteDefault(func() {
			atomic.AddInt64(&counter, 5)
		})
		g.Close()
		So(counter, ShouldEqual, 5)
	})
	Convey("test parallel 1", t, func() {
		wg := sync.WaitGroup{}
		p := &FixedGoroutinePool{Size: 1}
		p.Start()
		c := 0
		for i := 0; i < 10; i++ {
			wg.Add(1)
			p.Execute(func() { c++; wg.Done() })
		}
		wg.Wait()
		So(c, ShouldEqual, 10)
	})
	Convey("test parallel 5", t, func() {
		wg := sync.WaitGroup{}
		l := sync.Mutex{}
		p := &FixedGoroutinePool{Size: 5}
		p.Start()
		c := 0
		for i := 0; i < 10; i++ {
			wg.Add(1)
			p.Execute(func() {
				l.Lock()
				c++
				l.Unlock()
				wg.Done()
			})
		}
		wg.Wait()
		So(c, ShouldEqual, 10)
	})

}
