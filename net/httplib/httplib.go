/**
 * Copyright 2018 godog Author. All Rights Reserved.
 * Author: Chuck1024
 */

package httplib

import (
	"bytes"
	"encoding/json"
	me "github.com/chuck1024/godog/error"
	"github.com/xuyu/logging"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

var (
	HttpGet    = "GET"
	HttpPut    = "PUT"
	HttpPost   = "POST"
	HttpPatch  = "PATCH"
	HttpDelete = "DELETE"

	CONTENT_NONE = ""
	CONTENT_JSON = "application/json"
	CONTENT_YAML = "application/yaml"
	CONTENT_ALL  = "*"
)

type IRequest interface {
	DoRequest(resp interface{}) error
}

type Request struct {
	URL    string
	Method string
	Body   string
	Req    *http.Request
}

type Response struct {
	Status     string // e.g. "200 OK"
	StatusCode int    // e.g. 200
	Proto      string // e.g. "HTTP/1.1"
	ProtoMajor int    // e.g. 1
	ProtoMinor int    // e.g. 1

	Body   string
	Header http.Header

	ContentLength    int64
	TransferEncoding []string
}

// Http client operation
func newRequest(method, url string, body string) (*Request, error) {
	req := &Request{
		Method: method,
		URL:    url,
		Body:   body,
	}

	request, err := http.NewRequest(req.Method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		logging.Error("[newRequest] Fatal error when create request, error = %s, url = %s", err.Error(), url)
		return nil, err
	}

	req.Req = request

	return req, nil
}

func (req *Request) addHeader(key string, value string) {
	req.Req.Header.Add(key, value)
}

func (req *Request) doRequest() (*Response, error) {
	logging.Debug("[doRequest] start connection[to: %s, method: %s, content: %s]", req.URL, req.Method, req.Body)

	client := &http.Client{}
	client.Timeout = time.Duration(10 * time.Second)

	response, err := client.Do(req.Req)
	if err != nil {
		logging.Error("[doRequest] Failed to talk with remote server, error = %s ", err.Error())
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logging.Error("[doRequest] Read Response Body, error = %s", err.Error())
		return nil, err
	}

	return &Response{
		Status:           response.Status,
		StatusCode:       response.StatusCode,
		Proto:            response.Proto,
		ProtoMajor:       response.ProtoMajor,
		ProtoMinor:       response.ProtoMinor,
		Body:             string(body),
		Header:           response.Header,
		ContentLength:    response.ContentLength,
		TransferEncoding: response.TransferEncoding,
	}, nil
}

func JsonSerialize(source map[string]interface{}) (string, error) {
	text, err := json.Marshal(&source)
	if err != nil {
		logging.Error("[Serialize] Failed to convert map to json byte, error: %s", err.Error())
		return "", err
	}

	return string(text), err
}

func Call(method, url string, body string, headers, params map[string][]string) (string, error) {
	if len(params) > 0 {
		url += "?"
	}
	first := true
	for k, v := range params {
		if !first {
			url += "&"
		}
		if len(v) > 0 {
			url = url + k + "=" + v[0]
		} else {
			url = url + k
		}

		first = false
	}

	req, err := newRequest(method, url, body)
	if err != nil {
		logging.Error("[Call] Failed to create Request")
		return "", err
	}

	for k, v := range headers {
		var headContent string
		for i, item := range v {
			if i != 0 {
				headContent += ", "
			}
			headContent += item
		}
		req.addHeader(k, headContent)
	}

	resp, err := req.doRequest()
	if err != nil {
		logging.Error("[Call] Failed to do Request, error = %s", err.Error())
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", err
	}

	return resp.Body, nil
}

func SendToServer(method, url string, headers, params map[string][]string, req, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	contentType := false
	if headers == nil {
		headers = map[string][]string{}
	}

	for k := range headers {
		k = strings.ToLower(k)
		if k == "content-type" {
			contentType = true
		}
	}

	if !contentType {
		headers["Content-Type"] = []string{CONTENT_JSON}
	}

	logging.Debug("[SendToServer] send to server req:%#v", req)

	response, err := Call(method, url, string(body), headers, params)
	if err != nil {
		logging.Error("[SendToServer] occur error:%s", err.Error())
		return err
	}

	if err = json.Unmarshal([]byte(response), resp); err != nil {
		return err
	}

	return nil
}

// Http server operation
type ResponseData struct {
	Result int         `json:"code"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
}

type LogRequestInfo struct {
	Method       string
	URL          string
	User_id      string
	X_Auth_Token string
	Form         interface{}
	Body         interface{}
}

func getResponseInfo(err *me.CodeError, data interface{}) []byte {
	response := &ResponseData{}
	if err == nil {
		response.Result = me.Success
		response.Msg = "ok"
	} else {
		response.Result = err.Code()
		response.Msg = err.Detail()
	}
	response.Data = data

	ret, ee := json.Marshal(response)
	if ee != nil {
		logging.Error("[getResponseInfo] Failed, %s, data = %v", err.ToString(), data)
		panic("[getResponseInfo] Failed, " + ee.Error())
	}

	return ret
}

func LogGetResponseInfo(req *http.Request, err *me.CodeError, data interface{}) []byte {
	ret := getResponseInfo(err, data)

	body := ""
	b_body, ee := ioutil.ReadAll(req.Body)
	if ee == nil {
		body = string(b_body)
	}

	logReq := LogRequestInfo{
		Method: req.Method,
		URL:    req.RequestURI,
		//User_id:      userid,
		//X_Auth_Token: req.Header.Get("X-Auth-Token"),
		Form: req.Form,
		Body: body,
	}

	logging.Debug("[LogGetResponseInfo] HANDLE_LOG:request=%+v,response=%s", logReq, string(ret))

	return ret
}

func GetRequestBody(req *http.Request, v interface{}) error {
	buff := bytes.NewBufferString("")
	_, err := io.Copy(buff, req.Body)
	if err != nil {
		logging.Error("[GetRequestBody] copy req body error, error=%s", err.Error())
		return err
	}
	text := buff.String()
	req.Body = ioutil.NopCloser(strings.NewReader(text))

	err = json.Unmarshal(buff.Bytes(), v)
	if err != nil {
		logging.Error("[GetRequestBody] decode json body error, text=%s, error=%s", text, err.Error())
		return err
	}

	return nil
}
