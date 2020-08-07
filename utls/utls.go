/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Chuck1024
 */

package utls

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/chuck1024/dlog"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func loadFile(filename string) ([]byte, error) {
	file, err := os.Open(filename)
	defer file.Close()

	if err != nil {
		return nil, err
	}

	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func LoadJsonToObject(filename string, t interface{}) error {
	buf, e := loadFile(filename)
	if buf == nil || e != nil {
		return e
	}

	if err := json.Unmarshal(buf, &t); err != nil {
		return err
	}

	return nil
}

func FuncName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	funcName := filepath.Ext(runtime.FuncForPC(pc).Name())
	return strings.TrimPrefix(funcName, ".")
}

func HumanSize(s uint64) string {
	const (
		b = 1
		k = 1024 * b
		m = 1024 * k
		g = 1024 * m
	)
	switch {
	case s/g > 0:
		return fmt.Sprintf("%.1fGB", float64(s)/float64(g))
	case s/m > 0:
		return fmt.Sprintf("%.1fMB", float64(s)/float64(m))
	case s/k > 0:
		return fmt.Sprintf("%.1fKB", float64(s)/float64(k))
	default:
		return fmt.Sprintf("%dB", s)
	}
}

// k, m, g
func ParseMemorySize(size string) (uint64, error) {
	sizeSuffix := size[len(size)-1:]
	sizeNum, err := strconv.ParseInt(size[:len(size)-1], 10, 0)
	if err != nil {
		return uint64(0), nil
	}

	switch sizeSuffix {
	case "k":
		sizeNum = sizeNum * 1024
	case "K":
		sizeNum = sizeNum * 1024
	case "m":
		sizeNum = sizeNum * 1024 * 1024
	case "M":
		sizeNum = sizeNum * 1024 * 1024
	case "g":
		sizeNum = sizeNum * 1024 * 1024 * 1024
	case "G":
		sizeNum = sizeNum * 1024 * 1024 * 1024
	default:
		return uint64(0), fmt.Errorf("unsupport suffix:%s", sizeSuffix)
	}

	return uint64(sizeNum), nil
}

//no escape html
func Marshal(v interface{}) ([]byte, error) {
	var bbuf bytes.Buffer
	enc := json.NewEncoder(&bbuf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	bodyBts := bbuf.Bytes()
	l := len(bodyBts)
	if l > 0 && bodyBts[l-1] == '\n' {
		return bodyBts[:l-1], nil
	} else {
		return bodyBts, nil
	}
}

func WithRecover(fn func(), errHandler func(interface{})) (err interface{}) {
	defer func() {
		if err = recover(); err != nil {
			fmt.Fprintln(os.Stderr, "panic_recovered")
			if errHandler != nil {
				errHandler(err)
			}
		}
	}()

	fn()
	return
}

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success.
// copy the file contents from src to dst.
func CopyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	err = copyFileContents(src, dst)
	return
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}

var letterRunes = []rune("0123456789abcdefghipqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}

func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func Store2File(file string, content string) error {
	if content == "" {
		dlog.Error("write empty to file? file=%s,content=%s", file, content)
	}
	dir, err := filepath.Abs(filepath.Dir(file))
	if err != nil {
		return err
	}
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

func TraceId() string {
	h := md5.New()
	rand.Seed(time.Now().UnixNano())
	h.Write([]byte(strconv.FormatInt(rand.Int63(), 10)))
	h.Write([]byte("-"))
	h.Write([]byte(strconv.FormatInt(time.Now().UnixNano(), 10)))
	h.Write([]byte("-"))
	h.Write([]byte(strconv.FormatInt(int64(rand.Int31()), 10)))
	return hex.EncodeToString(h.Sum([]byte("godog")))
}