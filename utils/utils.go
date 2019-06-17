/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

// IsLocalIP, 检查是否为本地 IP
func IsLocalIP(request *http.Request) (bool, error) {
	ip, err := GetRealIP(request)
	if err != nil {
		return false, err
	}
	if (len(ip) >= 3 && ip[0:3] == "10.") || ip == "127.0.0.1" {
		return true, nil
	}
	return false, nil
}

func GetRealIP(request *http.Request) (string, error) {
	var ip string

	if len(request.Header.Get("X-Forwarded-For")) > 0 {
		// Reference: http://en.wikipedia.org/wiki/X-Forwarded-For#Format
		xForwardedFor := strings.Split(request.Header.Get("X-Forwarded-For"), ", ")
		if len(xForwardedFor) > 0 && net.ParseIP(xForwardedFor[0]) != nil {
			ip = xForwardedFor[0]
		}
	}

	//nginx remoteAddr "X-Real-IP"
	if len(ip) == 0 && len(request.Header.Get("X-Real-IP")) > 0 {
		xRealIP := request.Header.Get("X-Real-IP")
		if len(xRealIP) > 0 && net.ParseIP(xRealIP) != nil {
			ip = xRealIP
		}
	}
	if len(ip) == 0 && len(request.RemoteAddr) > 0 {
		remoteAddr := strings.Split(request.RemoteAddr, ":")
		if len(remoteAddr) == 2 && net.ParseIP(remoteAddr[0]) != nil {
			ip = remoteAddr[0]
		}
	}

	if len(ip) == 0 {
		return "", errors.New(fmt.Sprintf("cannot get real ip from request %+v", request))
	}
	return ip, nil
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

func ParseJSON(path string, v interface{}) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	mode := info.Mode()
	if mode.IsDir() {
		return errors.New("Invalid config file.it is directory. ")
	}

	if !mode.IsRegular() {
		return errors.New("Invalid config file,it is not a regular file. ")
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var lines [][]byte
	buf := bytes.NewBuffer(data)
	for {
		line, err := buf.ReadBytes('\n')
		line = bytes.Trim(line, " \t\r\n")
		if len(line) > 0 && !bytes.HasPrefix(line, []byte("//")) {
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	data = bytes.Join(lines, []byte{})
	if err = json.Unmarshal(data, v); err != nil {
		return err
	}

	return nil
}

func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		os.Stderr.WriteString("Oops: " + err.Error() + "\n")
	} else {
		for _, a := range addrs {
			if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP.String()
				}
			}
		}
	}
	return ""
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
