/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package stat

import (
	"bytes"
	"fmt"
	dogErr "github.com/gdp-org/gd/derror"
	"github.com/gdp-org/gd/dlog"
	"github.com/gdp-org/gd/utls"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type StatValue struct {
	Total  int
	SumVal int
	Avg    int64
	Max    int64
	Min    int64
	Gt10   int
	Gt100  int
	Gt500  int
	TotalD int64
}

type Stat struct {
	cmd string
	b   time.Time
	e   time.Time
	ret int
}

func NewStat() *Stat {
	return &Stat{}
}

func (st *Stat) Begin(cmd string) *Stat {
	st.cmd = cmd
	st.b = time.Now()
	return st
}

// BeginAt begin is micro seconds
func (st *Stat) BeginAt(cmd string, begin time.Time) *Stat {
	st.cmd = cmd
	st.b = begin
	return st
}

// BeginFuncCmd use funcName as cmd
func (st *Stat) BeginFuncCmd() *Stat {
	st.cmd = utls.FuncName(2)
	st.b = time.Now()
	return st
}

// BeginFuncCmdAt use funcName as cmd but set begin
func (st *Stat) BeginFuncCmdAt(begin time.Time) *Stat {
	st.cmd = utls.FuncName(2)
	st.b = begin
	return st
}

// beginFuncAt use funcName as cmd but set begin
func (st *Stat) beginFuncAt(pre string, skip int, begin time.Time) *Stat {
	st.cmd = pre + "." + utls.FuncName(skip)
	st.b = begin
	return st
}

func (st *Stat) End(ret int) {
	st.ret = ret
	st.e = time.Now()

	if statMgr == nil {
		dlog.Error("StatMgr is not init.")
	} else {
		statMgr.addStat(st)
	}
}

func (st *Stat) EndErr(err error) {
	if err == nil {
		st.End(0)
	} else {
		switch err.(type) {
		case *dogErr.CodeError:
			err0 := err.(*dogErr.CodeError)
			if err0 != nil {
				st.End(err0.Code())
			} else {
				st.End(0)
			}
		case error:
			st.End(-1)
		default:
			st.End(-2)
		}
	}
}

// Elapse return Duration
func (st *Stat) Elapse() time.Duration {
	return st.e.Sub(st.b)
}

func DeferStat(b time.Time, err **dogErr.CodeError) {
	NewStat().beginFuncAt("", 3, b).EndErr(*err)
}

func DeferAppStat(appid uint32, b time.Time, err **dogErr.CodeError) {
	NewStat().beginFuncAt(strconv.FormatUint(uint64(appid), 10), 3, b).EndErr(*err)
}

type StatMgr struct {
	m         map[string]*StatValue
	c         chan *Stat
	statFile  *os.File
	statGap   time.Duration
	maxCmdLen int
}

var statMgr *StatMgr
var once sync.Once

func StatMgrInstance() *StatMgr {
	once.Do(func() {
		statMgr = &StatMgr{
			m: make(map[string]*StatValue),
			c: make(chan *Stat, 4096),
		}
	})
	return statMgr
}

func (mgr *StatMgr) Init(statPath string, statGap time.Duration) {
	var err error
	if mgr.statFile, err = os.OpenFile(statPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		dlog.Error("init stat file failed, %s", err.Error())
		return
	}

	mgr.statGap = statGap
	ticker := time.NewTicker(statGap)
	go func() {
		for {
			select {
			case <-ticker.C:
				mgr.dump()
			case st := <-mgr.c:
				mgr.stat(st)
			}
		}
	}()
}

func (mgr *StatMgr) addStat(st *Stat) {
	mgr.c <- st
}

const sep = "$$$"

var buf = bytes.NewBufferString("")
var tt = int64(time.Millisecond / time.Microsecond)

func (mgr *StatMgr) dump() {
	defer buf.Truncate(0)

	if len(mgr.m) < 1 {
		return
	}

	title := fmt.Sprintf("===============PID %d, Statistic in %ds, %s=====================\n", os.Getpid(), int(mgr.statGap.Seconds()), time.Now().Format("2006-01-02 15:04:05"))
	buf.WriteString(title)

	headFormat := fmt.Sprintf("%%-%ds|%%8s|%%8s|%%8s|%%9s|%%9s|%%9s|%%-18s|%%11s|%%11s|%%11s|\n", mgr.maxCmdLen)

	head := fmt.Sprintf(headFormat, "", "RESULT", "TOTAL", "SUMVAL", "AVG(ms)", "MAX(ms)", "MIN(ms)", "RECATMAX", ">10.000ms", ">100.000ms", ">500.000ms")
	buf.WriteString(head)

	contentFormat := fmt.Sprintf("%%-%ds|%%8d|%%8d|%%8d|%%9.3f|%%9.3f|%%9.3f|%%-18s|%%11d|%%11d|%%11d|\n", mgr.maxCmdLen)

	total := 0
	max := int64(0)
	min := int64(math.MaxInt64)
	gt10 := 0
	gt100 := 0
	gt500 := 0

	var err error

	keys := make([]string, 0)
	for k := range mgr.m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := mgr.m[k]
		delete(mgr.m, k)

		sps := strings.Split(k, sep)
		if len(sps) != 2 {
			dlog.Error("invalid stat key, ", k)
			continue
		}
		cmd := sps[0]
		ret := 0

		if ret, err = strconv.Atoi(sps[1]); err != nil {
			ret = math.MaxInt32
		}

		content := fmt.Sprintf(contentFormat, cmd, ret, v.Total, 0, float64(v.TotalD)/float64(int64(v.Total)*tt), float64(v.Max)/float64(tt), float64(v.Min)/float64(tt), "", v.Gt10, v.Gt100, v.Gt500)
		buf.WriteString(content)

		total += v.Total
		if v.Max > max {
			max = v.Max
		}
		if v.Min < min {
			min = v.Min
		}
		if v.Gt10 > 0 {
			gt10 += v.Gt10
		}
		if v.Gt100 > 0 {
			gt100 += v.Gt100
		}
		if v.Gt500 > 0 {
			gt500 += v.Gt500
		}
	}
	buf.WriteString("----------------------------------------------------------------------------------------\n")

	if min == math.MaxInt64 {
		min = 0
	}
	tailFormat := fmt.Sprintf("%%-%ds|%%8d|%%8d|%%8d|%%9.3f|%%9.3f|%%9.3f|                  |%%11d|%%11d|%%11d|\n", mgr.maxCmdLen)
	tail := fmt.Sprintf(tailFormat, "ALL", 0, total, 0, 0.0, float64(max)/float64(tt), float64(min)/float64(tt), gt10, gt100, gt500)
	buf.WriteString(tail)
	buf.WriteString("\n")

	if _, err = mgr.statFile.Write(buf.Bytes()); err != nil {
		dlog.Error("write stat failed, %s", err.Error())
	} else {
		mgr.rotateFile()
	}
}

func (mgr *StatMgr) stat(st *Stat) {
	cmd := st.cmd
	ret := st.ret
	cmdRet := cmd + sep + strconv.Itoa(ret)
	d := st.e.Sub(st.b).Nanoseconds() / int64(time.Microsecond)
	var v *StatValue
	var ok bool
	if v, ok = mgr.m[cmdRet]; ok {
		if d > v.Max {
			v.Max = d
		}
		if d < v.Min {
			v.Min = d
		}

		v.Total++
		v.TotalD += d
	} else {
		v = &StatValue{
			Total:  1,
			TotalD: d,
			Max:    d,
			Min:    d,
		}

		mgr.m[cmdRet] = v
	}

	if len(cmd) > mgr.maxCmdLen {
		mgr.maxCmdLen = len(cmd)
	}

	ms10 := int64(time.Microsecond * 10)
	ms100 := int64(time.Microsecond * 100)
	ms500 := int64(time.Microsecond * 500)

	if ms10 <= d && d < ms100 {
		v.Gt10++
	} else if ms100 <= d && d < ms500 {
		v.Gt100++
	} else if ms500 <= d {
		v.Gt500++
	} else {
		//ignore
	}
}

const MaxStatFileSize = 100 * 1024 * 1024
const MaxStatFileCount = 10

func (mgr *StatMgr) rotateFile() {
	f, err := mgr.statFile.Stat()
	if err != nil {
		dlog.Error("stat %s failed.", mgr.statFile.Name())
		return
	}
	if f.Size() < int64(MaxStatFileSize) {
		return
	}

	mgr.statFile.Close()

	statFileName := mgr.statFile.Name()
	for i := MaxStatFileCount - 1; i > 0; i-- {
		oldFileName := fmt.Sprintf("%s.%d", statFileName, i)
		newFileName := fmt.Sprintf("%s.%d", statFileName, i+1)

		os.Rename(oldFileName, newFileName)
	}

	newFileName := fmt.Sprintf("%s.1", statFileName)
	os.Rename(statFileName, newFileName)

	mgr.statFile, err = os.OpenFile(statFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		dlog.Error("rotate stat file failed: %s", err.Error())
	}
}
