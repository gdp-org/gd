/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package helper

import (
	"bufio"
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/utls"
	"io"
	"net"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strconv"
	"strings"
	"time"
)

const ProfDumpDir = "prof"
const ProfDumpBackupDir = "./prof"

type UpdateFunc func(Key, Value, Type string, Offset int64) bool
type FindFunc func(key string) (string, error)
type PerfFunc func() string
type StatusFunc func() string
type DeployFunc func([]string) string

type Helper struct {
	listener net.Listener

	Host     string
	Updater  UpdateFunc
	Finder   FindFunc
	Perfer   PerfFunc
	Stater   StatusFunc
	Deployer DeployFunc
}

func (helper *Helper) Start() error {
	l, err := net.Listen("tcp", helper.Host)
	if err != nil {
		return fmt.Errorf("host=%s,%v", helper.Host, err)
	}
	helper.listener = l

	err = os.MkdirAll(ProfDumpDir, 0755)
	if err != nil {
		dlog.Info("Helper create dir fail, path:%s, err:%v", ProfDumpDir, err)
	}

	err = os.MkdirAll(ProfDumpBackupDir, 0755)
	if err != nil {
		dlog.Info("Helper create dir fail, path:%s, err:%v", ProfDumpDir, err)
	}

	go helper.waitTcp()
	return nil
}

func (helper *Helper) waitTcp() {
	for {
		if c, err := helper.listener.Accept(); err == nil {
			go utls.WithRecover(func() {
				helper.dealCommand(c)
			}, nil)
		} else {
			if x, ok := err.(*net.OpError); ok && x.Op == "accept" {
				dlog.Info("Stoping Tcp Accept")
				break
			}

			dlog.Warn("Accept failed: %v", err)
			continue
		}
	}
}

func (helper *Helper) Close() {
	helper.listener.Close()
}

func WriteHeap() {
	writeHeap(true)
}

func writeHeap(forceGc bool) {
	if forceGc {
		runtime.GC()
	}

	fn0 := fmt.Sprintf("heap_%s.prof", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write heap fail, %s", err)
		return
	}
	defer f.Close()
	pprof.Lookup("heap").WriteTo(f, 1)
	dlog.Info("write heap to file %s", f)
}

func writeTrace(s int64) {
	fn0 := fmt.Sprintf("trace_%s.out", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write trace fail, %s", err)
		return
	}
	defer f.Close()
	err = trace.Start(f)
	if err != nil {
		dlog.Error("start trace fail, %s", err)
		return
	}
	defer trace.Stop()
	time.Sleep(time.Duration(s) * time.Second)
	dlog.Info("write trace to file %s", fn)
}

func WriteMutex() {
	fn0 := fmt.Sprintf("mutex_%s.prof", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write mutex fail, %s", err)
		return
	}
	defer f.Close()
	pprof.Lookup("mutex").WriteTo(f, 1)
	dlog.Info("write mutex to file %s", f)
}

func WriteThreadCreate() {
	fn0 := fmt.Sprintf("threadCreate_%s.prof", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write threadCreate fail, %s", err)
		return
	}
	defer f.Close()
	pprof.Lookup("threadcreate").WriteTo(f, 1)
	dlog.Info("write threadCreate to file %s", f)
}

func WriteBlock() {
	fn0 := fmt.Sprintf("block_%s.prof", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write block fail, %s", err)
		return
	}
	defer f.Close()
	pprof.Lookup("block").WriteTo(f, 1)
	dlog.Info("write block to file %s", f)
}

func writeGoroutine() {
	fn0 := fmt.Sprintf("goroutine_%s.prof", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write goroutine fail, %s", err)
		return
	}
	defer f.Close()
	pprof.Lookup("goroutine").WriteTo(f, 1)
	dlog.Info("write gouroutine to file %s", f)
}

func CpuProfiling(s int64) {
	cpuProfiling(s)
}

func cpuProfiling(s int64) {
	fn0 := fmt.Sprintf("cpuprofile_%s.prof", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write cpuprofile fail, %s", err)
		return
	}
	defer f.Close()
	if err := pprof.StartCPUProfile(f); err != nil {
		dlog.Error("write cpuprofile fail, %s", err)
		return
	}
	time.Sleep(time.Duration(s) * time.Second)
	pprof.StopCPUProfile()
}

func WriteGoroutine() {
	fn0 := fmt.Sprintf("goroutine_%s.prof", time.Now().Format("2006_01_02_03_04_05"))
	fn := "prof/" + fn0
	f, err := os.Create(fn)
	if err != nil {
		dlog.Error("write goroutine fail, %s", err)
		return
	}
	defer f.Close()
	pprof.Lookup("goroutine").WriteTo(f, 1)
	dlog.Info("write goroutine to file %s", f)
}

func (helper *Helper) help(client net.Conn) {
	client.Write([]byte("THIS IS HELP\n--------------------------------------------------------------\n"))
	client.Write([]byte("  add\t\t<id> <key> <sec> <min> <hour> <day> <month> <week> <year>\n"))
	client.Write([]byte("  del\t\t<id> <key> <sec> <min> <hour> <day> <month> <week> <year>\n"))
	client.Write([]byte("  status\treturn status\n"))
	client.Write([]byte("  find\t\t<id> <key>\n"))
	client.Write([]byte("  trace n \t\tdump n second trace info to prof/trace_xxx.out\n"))
	client.Write([]byte("  heap <nogc>\t\tdump current heap to prof/heap_xxx.prof\n"))
	client.Write([]byte("  dlog\t\t-1:CURRENT,0:FNST,1:FINE,2:DEBG,3:TRAC,4:INFO,5:WARN,6:EROR,7:CRIT\n"))
	client.Write([]byte("  deploy\t<file1> <file2> ;\"deploy all\" means deploy all accessable file\n"))
}

func (helper *Helper) dealCommand(client net.Conn) {
	defer client.Close()

	bio := bufio.NewReader(client)
	for {
		data, err := bio.ReadString('\n')
		if err == io.EOF {
			client.Write([]byte("<end\n"))
			break
		} else if err != nil {
			client.Write([]byte("<" + err.Error() + "\n"))
			break
		}
		if data == "\r\n" || data == "\n" {
			client.Write([]byte("<<\n"))
			break
		}
		data = strings.TrimSpace(data)
		arr := strings.Split(data, " ")
		tpe := arr[0]
		ret := true
		if tpe == "add" || tpe == "del" {
			if len(arr) != 10 {
				helper.help(client)
				client.Write([]byte("<invalid len\n"))
				ret = false
				continue
			}
			id := arr[1]
			key := arr[2]
			value := strings.Join(arr[3:], " ")
			ret = helper.Updater(id+" "+key, value, tpe, -1)
		} else if tpe == "log" {
			if len(arr) != 2 {
				helper.help(client)
				ret = false
				continue
			}
			lvl, err := strconv.Atoi(arr[1])
			if err != nil {
				client.Write([]byte("<" + err.Error() + "\n"))
				ret = false
			} else if lvl == -1 {
				client.Write([]byte("<log:" + dlog.GetLevel() + "\n"))
			} else {
				dlog.Debug("SetLogLevel:%d", lvl)
				dlog.SetLevel(lvl)
			}
		} else if tpe == "heap" {
			forceGc := true
			if len(arr) == 2 && arr[1] == "nogc" {
				forceGc = false
			}
			writeHeap(forceGc)
			writeGoroutine()
		} else if tpe == "trace" {
			var s int64
			if len(arr) < 2 {
				s = 10
			} else {
				st, err := strconv.ParseInt(arr[1], 10, 0)
				if err != nil {
					s = 30
				} else {
					s = st
				}
			}
			writeTrace(s)
		} else if tpe == "status" {
			mem := MemStats()
			mem2 := MemStats2()
			gc := GcStats()
			numGos := runtime.NumGoroutine()
			lc := ""
			if helper.Stater != nil {
				lc = helper.Stater()
			}

			perf := ""
			if helper.Perfer != nil {
				perf = helper.Perfer()
			}

			status := fmt.Sprintf("mem:%s\nmem2:%s\ngc:%s\nnumGos:%d\nstater:%s\n%s\n", mem, mem2, gc, numGos, lc, perf)
			client.Write([]byte(status))
		} else if tpe == "find" {
			key := strings.Join(arr[1:3], " ")
			fvalue, ferr := helper.Finder(key)
			if ferr != nil {
				client.Write([]byte(fmt.Sprintf("err:%s,key=%s\n", ferr, key)))
			} else {
				client.Write([]byte(fmt.Sprintf("found:%s,key=%s\n", fvalue, key)))
			}
		} else if tpe == "deploy" {
			if helper.Deployer != nil {
				deployresult := helper.Deployer(arr[1:])
				client.Write([]byte(fmt.Sprintf("deploy: finished result is %s", deployresult)))
			} else {
				client.Write([]byte(fmt.Sprintf("err: not support deploy for helper project")))
			}
		} else if tpe == "profile" {
			var s int64
			if len(arr) < 2 {
				s = 30
			} else {
				st, err := strconv.ParseInt(arr[1], 10, 0)
				if err != nil {
					s = 30
				} else {
					s = st
				}
			}
			cpuProfiling(s)
			client.Write([]byte("profile: finished"))
		} else {
			helper.help(client)
		}
		if ret == true {
			client.Write([]byte("<ok\n"))
		} else {
			client.Write([]byte("<err\n"))
		}
	}
}
