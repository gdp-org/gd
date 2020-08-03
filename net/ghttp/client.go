/**
 * Copyright 2019 godog Author. All rights reserved.
 * Author: Chuck1024
 */

package ghttp

import (
	"errors"
	"fmt"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"strings"
	"time"
)

type HttpClient struct {
	Timeout time.Duration
	Domain  string
}

func (c *HttpClient) Start() error {
	if c.Timeout <= 0 {
		c.Timeout = 3 * time.Second
	}
	if c.Domain == "" {
		return errors.New("http client need domain")
	}
	if !strings.HasPrefix(c.Domain, "http") {
		return fmt.Errorf("need protocol %v", c.Domain)
	}
	return nil
}

func (c *HttpClient) Method(method string, path string, header map[string]string, params interface{}) (*http.Response, string, error) {
	return c.MethodTimeout(method, path, header, params, c.Timeout)
}

func (c *HttpClient) MethodTimeout(method string, path string, header map[string]string, params interface{}, timeout time.Duration) (*http.Response, string, error) {
	dm := c.Domain
	if !strings.HasSuffix(dm, "/") && !strings.HasPrefix(path, "/") {
		dm = dm + "/"
	}

	target := dm + path
	a := gorequest.New().Timeout(timeout).CustomMethod(strings.ToUpper(method), target)
	for k, v := range header {
		a.Set(k, v)
	}

	if params != nil {
		if strings.ToUpper(method) == "GET" {
			a.Query(params)
		} else {
			a.Send(params)
		}
	}

	resp, body, errs := a.End()
	var err error
	el := len(errs)
	if el > 0 {
		if el == 1 {
			err = errs[0]
		} else {
			var errStrs []string
			for _, e := range errs {
				if e != nil {
					errStrs = append(errStrs, e.Error())
				}
			}
			if len(errStrs) > 0 {
				err = errors.New(strings.Join(errStrs, "|"))
			}
		}
	}

	return resp, body, err
}
