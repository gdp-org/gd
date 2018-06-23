/**
 * Created by JetBrains GoLand.
 * Author: Chuck Chen
 * Date: 2018/6/22
 * Time: 15:47
 */

package http

import (
	"encoding/json"
	"github.com/xuyu/logging"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"bytes"
	"io"
	me "godog/error"
)

var (
	HTTP_METHOD_GET    = "GET"
	HTTP_METHOD_PUT    = "PUT"
	HTTP_METHOD_POST   = "POST"
	HTTP_METHOD_PATCH  = "PATCH"
	HTTP_METHOD_DELETE = "DELETE"

	CONTENT_NONE = ""
	CONTENT_JSON = "application/json"
	CONTENT_YAML = "application/yaml"
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

type Responser http.ResponseWriter
type Requester http.Request
type HandlerFunc func(http.ResponseWriter, *http.Request)

type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

func Serve(srvPort string, handler http.Handler) {
	logging.Info("[Serve] Try to listen on port: %s", srvPort)
	go func() {
		err := http.ListenAndServe(srvPort, handler)
		if err != nil {
			logging.Error("[Serve] Listen failed, error = %s", err.Error())
			return
		}
	}()
}

func Health(srvPort string, handler http.Handler) {
	logging.Info("[Health] Try to monitor health condition on port: %s", srvPort)
	go func() {
		err := http.ListenAndServe(srvPort, handler)
		if err != nil {
			logging.Error("[Health] monitor failed, error = %s", err.Error())
			return
		}
	}()
}

func HandleFunc(addr string, handler HandlerFunc) {
	http.HandleFunc(addr, handler)
}

func newRequest(method, url string, body string) (*Request, error) {
	req := &Request{
		Method: method,
		URL:    url,
		Body:   body,
	}

	request, err := http.NewRequest(req.Method, req.URL, strings.NewReader(req.Body))
	if err != nil {
		logging.Error("[NewRequest] Fatal error when create request, error = %s, url = %s", err.Error(), url)
		return nil, err
	}

	req.Req = request

	return req, nil
}

func (req *Request) addHeader(key string, value string) {
	req.Req.Header.Add(key, value)
}

func (req *Request) doRequest() (*Response, error) {
	logging.Debug("[Request.DoRequest, id: %s] start connection[to: %s, method: %s, content: %s]", req.URL, req.Method, req.Body)

	client := &http.Client{}
	client.Timeout = time.Duration(10 * time.Second)

	response, err := client.Do(req.Req)
	if err != nil {
		logging.Error("[Request.DoRequest] Failed to talk with remote server, error = %s ", err.Error())
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logging.Error("[Request.DoRequest] Read Response Body, error = %s", err.Error())
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

func Serialize(source map[string]interface{}) (string, error) {
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

func GetResponseInfo(err *me.MError, data interface{}) []byte {
	response := &ResponseData{}
	if err == nil {
		response.Result = me.ERR_CODE_SUCCESS
		response.Msg = "ok"
	} else {
		response.Result = err.Code()
		response.Msg = err.Detail()
	}
	response.Data = data

	ret, ee := json.Marshal(response)
	if ee != nil {
		logging.Error("[GetResponseInfo] Failed, %s, data = %v", err.ToString(), data)
		panic("[GetResponseInfo] Failed, " + ee.Error())
	}

	return ret
}

func LogGetResponseInfo(req *http.Request, err *me.MError, data interface{}) []byte {
	ret := GetResponseInfo(err, data)

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

	logging.Debug("HANDLE_LOG:request=%+v,response=%s", logReq, string(ret))

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
