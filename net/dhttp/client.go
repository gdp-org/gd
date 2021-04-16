// Package HttpClient is a simplified http client.
// Its initial codes are cloned from [HttpRequest](https://github.com/parnurzeal/gorequest). I have refactored the codes and make it more friendly to programmers.
// HttpClient makes http thing more simple for you, using fluent styles to make http client more awesome. You can control headers, timeout, query parameters, binding response and others in one line:
//
// Before
//
// client := &http.Client{
// 	 CheckRedirect: redirectPolicyFunc,
// }
// req, err := http.NewRequest("GET", "http://example.com", nil)
// req.Header.Add("If-None-Match", `W/"wyzzy"`)
// resp, err := client.Do(req)
//
// Using HttpClient
//
// resp, body, errs := dhttp.New().Get("http://example.com").
//   RedirectPolicy(redirectPolicyFunc).
//   SetHeader("If-None-Match", `W/"wyzzy"`).
//   End()
package dhttp

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/chuck1024/gd/dlog"
	"github.com/chuck1024/gd/runtime/gl"
	"github.com/chuck1024/gd/runtime/pc"
	"github.com/chuck1024/gd/utls"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/publicsuffix"
	"moul.io/http2curl"
)

type Request *http.Request
type Response *http.Response

const (
	glHttpClientCall     = "http_call_%v"
	glHttpClientCost     = "http_cost_%v"
	glHttpClientCallFail = "http_call_fail_%v"
)

// HTTP methods we support
const (
	POST    = "POST"
	GET     = "GET"
	HEAD    = "HEAD"
	PUT     = "PUT"
	DELETE  = "DELETE"
	PATCH   = "PATCH"
	OPTIONS = "OPTIONS"
)

// Types we support.
const (
	TypeJSON       = "json"
	TypeXML        = "xml"
	TypeUrlencoded = "urlencoded"
	TypeForm       = "form"
	TypeFormData   = "form-data"
	TypeHTML       = "html"
	TypeText       = "text"
	TypeMultipart  = "multipart"
)

type HttpClientRetryable struct {
	RetryableStatus []int
	RetryTime       time.Duration
	RetryCount      int
	Attempt         int
	Enable          bool
}

// A HttpClient is a object storing all request data for client.
type HttpClient struct {
	Url                  string
	Method               string
	Header               http.Header
	TargetType           string
	ForceType            string
	Data                 map[string]interface{}
	SliceData            []interface{}
	FormData             url.Values
	QueryData            url.Values
	FileData             []File
	BounceToRawString    bool
	RawString            string
	Client               *http.Client
	Transport            *http.Transport
	Cookies              []*http.Cookie
	Errors               []error
	BasicAuth            struct{ Username, Password string }
	Debug                bool
	CurlCommand          bool
	Retryable            HttpClientRetryable
	DoNotClearHttpClient bool
	isClone              bool
}

var DisableTransportSwap = false

// Used to create a new HttpClient object.
func New() *HttpClient {
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)

	debug := os.Getenv("HttpClient_DEBUG") == "1"

	dhc := &HttpClient{
		TargetType:        TypeJSON,
		Data:              make(map[string]interface{}),
		Header:            http.Header{},
		RawString:         "",
		SliceData:         []interface{}{},
		FormData:          url.Values{},
		QueryData:         url.Values{},
		FileData:          make([]File, 0),
		BounceToRawString: false,
		Client:            &http.Client{Jar: jar},
		Transport:         &http.Transport{},
		Cookies:           make([]*http.Cookie, 0),
		Errors:            nil,
		BasicAuth:         struct{ Username, Password string }{},
		Debug:             debug,
		CurlCommand:       false,
		isClone:           false,
	}
	dhc.Transport.DisableKeepAlives = true
	return dhc
}

func cloneMapArray(old map[string][]string) map[string][]string {
	newMap := make(map[string][]string, len(old))
	for k, vals := range old {
		newMap[k] = make([]string, len(vals))
		for i := range vals {
			newMap[k][i] = vals[i]
		}
	}
	return newMap
}

func shallowCopyData(old map[string]interface{}) map[string]interface{} {
	if old == nil {
		return nil
	}
	newData := make(map[string]interface{})
	for k, val := range old {
		newData[k] = val
	}
	return newData
}

func shallowCopyDataSlice(old []interface{}) []interface{} {
	if old == nil {
		return nil
	}
	newData := make([]interface{}, len(old))
	for i := range old {
		newData[i] = old[i]
	}
	return newData
}

func shallowCopyFileArray(old []File) []File {
	if old == nil {
		return nil
	}
	newData := make([]File, len(old))
	for i := range old {
		newData[i] = old[i]
	}
	return newData
}

func shallowCopyCookies(old []*http.Cookie) []*http.Cookie {
	if old == nil {
		return nil
	}
	newData := make([]*http.Cookie, len(old))
	for i := range old {
		newData[i] = old[i]
	}
	return newData
}

func shallowCopyErrors(old []error) []error {
	if old == nil {
		return nil
	}
	newData := make([]error, len(old))
	for i := range old {
		newData[i] = old[i]
	}
	return newData
}

func (dhc *HttpClient) setJSONHeaders(headers string) *HttpClient {
	var val map[string]string
	if err := json.Unmarshal([]byte(headers), &val); err == nil {
		for k, v := range val {
			dhc.SetHeader(k, v)
		}
	} else {
		dhc.Errors = append(dhc.Errors, err)
	}
	return dhc
}

func (dhc *HttpClient) setStructHeaders(headers interface{}) *HttpClient {
	if marshalContent, err := json.Marshal(headers); err != nil {
		dhc.Errors = append(dhc.Errors, err)
	} else {
		var val map[string]string
		if err := json.Unmarshal(marshalContent, &val); err != nil {
			dhc.Errors = append(dhc.Errors, err)
		} else {
			for k, v := range val {
				dhc.SetHeader(k, v)
			}
		}
	}
	return dhc
}

// just need to change the array pointer?
func copyRetryable(old HttpClientRetryable) HttpClientRetryable {
	newRetryable := old
	newRetryable.RetryableStatus = make([]int, len(old.RetryableStatus))
	for i := range old.RetryableStatus {
		newRetryable.RetryableStatus[i] = old.RetryableStatus[i]
	}
	return newRetryable
}

// Returns a copy of this HttpClient. Useful if you want to reuse the client/settings
// concurrently.
// Note: This does a shallow copy of the parent. So you will need to be
// careful of Data provided
// Note: It also directly re-uses the client and transport. If you modify the Timeout,
// or RedirectPolicy on a clone, the clone will have a new http.client. It is recommended
// that the base request set your timeout and redirect polices, and no modification of
// the client or transport happen after cloning.
// Note: DoNotClearHttpClient is forced to "true" after Clone
func (dhc *HttpClient) Clone() *HttpClient {
	clone := &HttpClient{
		Url:                  dhc.Url,
		Method:               dhc.Method,
		Header:               http.Header(cloneMapArray(dhc.Header)),
		TargetType:           dhc.TargetType,
		ForceType:            dhc.ForceType,
		Data:                 shallowCopyData(dhc.Data),
		SliceData:            shallowCopyDataSlice(dhc.SliceData),
		FormData:             url.Values(cloneMapArray(dhc.FormData)),
		QueryData:            url.Values(cloneMapArray(dhc.QueryData)),
		FileData:             shallowCopyFileArray(dhc.FileData),
		BounceToRawString:    dhc.BounceToRawString,
		RawString:            dhc.RawString,
		Client:               dhc.Client,
		Transport:            dhc.Transport,
		Cookies:              shallowCopyCookies(dhc.Cookies),
		Errors:               shallowCopyErrors(dhc.Errors),
		BasicAuth:            dhc.BasicAuth,
		Debug:                dhc.Debug,
		CurlCommand:          dhc.CurlCommand,
		Retryable:            copyRetryable(dhc.Retryable),
		DoNotClearHttpClient: true,
		isClone:              true,
	}
	return clone
}

// Enable the debug mode which logs request/response detail
func (dhc *HttpClient) SetDebug(enable bool) *HttpClient {
	dhc.Debug = enable
	return dhc
}

// Enable the curl command mode which display a CURL command line
func (dhc *HttpClient) SetCurlCommand(enable bool) *HttpClient {
	dhc.CurlCommand = enable
	return dhc
}

// Enable the DoNotClear mode for not clearing super agent and reuse for the next request
func (dhc *HttpClient) SetDoNotClearHttpClient(enable bool) *HttpClient {
	dhc.DoNotClearHttpClient = enable
	return dhc
}

// Clear HttpClient data for another new request.
func (dhc *HttpClient) ClearHttpClient() {
	if dhc.DoNotClearHttpClient {
		return
	}
	dhc.Url = ""
	dhc.Method = ""
	dhc.Header = http.Header{}
	dhc.Data = make(map[string]interface{})
	dhc.SliceData = []interface{}{}
	dhc.FormData = url.Values{}
	dhc.QueryData = url.Values{}
	dhc.FileData = make([]File, 0)
	dhc.BounceToRawString = false
	dhc.RawString = ""
	dhc.ForceType = ""
	dhc.TargetType = TypeJSON
	dhc.Cookies = make([]*http.Cookie, 0)
	dhc.Errors = nil
}

// http timeout
func (dhc *HttpClient) Timeout(timeout time.Duration) *HttpClient {
	dhc.Client.Timeout = timeout
	return dhc
}

// Just a wrapper to initialize HttpClient instance by method string
func (dhc *HttpClient) CustomMethod(method, targetUrl string) *HttpClient {
	switch method {
	case POST:
		return dhc.Post(targetUrl)
	case GET:
		return dhc.Get(targetUrl)
	case HEAD:
		return dhc.Head(targetUrl)
	case PUT:
		return dhc.Put(targetUrl)
	case DELETE:
		return dhc.Delete(targetUrl)
	case PATCH:
		return dhc.Patch(targetUrl)
	case OPTIONS:
		return dhc.Options(targetUrl)
	default:
		dhc.ClearHttpClient()
		dhc.Method = method
		dhc.Url = targetUrl
		dhc.Errors = nil
		return dhc
	}
}

func (dhc *HttpClient) Get(targetUrl string) *HttpClient {
	dhc.ClearHttpClient()
	dhc.Method = GET
	dhc.Url = targetUrl
	dhc.Errors = nil
	return dhc
}

func (dhc *HttpClient) Post(targetUrl string) *HttpClient {
	dhc.ClearHttpClient()
	dhc.Method = POST
	dhc.Url = targetUrl
	dhc.Errors = nil
	return dhc
}

func (dhc *HttpClient) Head(targetUrl string) *HttpClient {
	dhc.ClearHttpClient()
	dhc.Method = HEAD
	dhc.Url = targetUrl
	dhc.Errors = nil
	return dhc
}

func (dhc *HttpClient) Put(targetUrl string) *HttpClient {
	dhc.ClearHttpClient()
	dhc.Method = PUT
	dhc.Url = targetUrl
	dhc.Errors = nil
	return dhc
}

func (dhc *HttpClient) Delete(targetUrl string) *HttpClient {
	dhc.ClearHttpClient()
	dhc.Method = DELETE
	dhc.Url = targetUrl
	dhc.Errors = nil
	return dhc
}

func (dhc *HttpClient) Patch(targetUrl string) *HttpClient {
	dhc.ClearHttpClient()
	dhc.Method = PATCH
	dhc.Url = targetUrl
	dhc.Errors = nil
	return dhc
}

func (dhc *HttpClient) Options(targetUrl string) *HttpClient {
	dhc.ClearHttpClient()
	dhc.Method = OPTIONS
	dhc.Url = targetUrl
	dhc.Errors = nil
	return dhc
}

// Set is used for setting header fields,
// this will overwrite the existed values of Header through AppendHeader().
// Example. To set `Accept` as `application/json`
//
//    New().
//      Post("/gamelist").
//      SetHeader("Accept", "application/json").
//      End()
func (dhc *HttpClient) SetHeader(param string, value string) *HttpClient {
	dhc.Header.Set(param, value)
	return dhc
}

// AppendHeader is used for setting header fileds with multiple values,
// Example. To set `Accept` as `application/json, text/plain`
//
//    New().
//      Post("/gamelist").
//      AppendHeader("Accept", "application/json").
//      AppendHeader("Accept", "text/plain").
//      End()
func (dhc *HttpClient) AppendHeader(param string, value string) *HttpClient {
	dhc.Header.Add(param, value)
	return dhc
}

// SetHeaders is used to set headers with multiple fields.
// it accepts structs or json strings:
// for example:
//    New().Get(ts.URL).
//    SetHeaders(`{'Content-Type' = 'text/plain','X-Test-Tag'='test'}`).
//    End()
//or
//    headers := struct {
//        ContentType string `json:"Content-Type"`
//        XTestTag string `json:"X-Test-Tag"`
//    } {ContentType:"text/plain",XTestTag:"test"}
//
//    New().Get(ts.URL).
//    SetHeaders(headers).
//    End()
//
func (dhc *HttpClient) SetHeaders(headers interface{}) *HttpClient {
	switch v := reflect.ValueOf(headers); v.Kind() {
	case reflect.String:
		dhc.setJSONHeaders(v.String())
	case reflect.Struct:
		dhc.setStructHeaders(v.Interface())
	default:
	}
	return dhc
}

// Retryable is used for setting a Retry policy
// Example. To set Retry policy with 5 seconds between each attempt.
//          3 max attempt.
//          And StatusBadRequest and StatusInternalServerError as RetryableStatus

//    New().
//      Post("/gamelist").
//      Retry(3, 5 * time.seconds, http.StatusBadRequest, http.StatusInternalServerError).
//      End()
func (dhc *HttpClient) Retry(retryCount int, retryTime time.Duration, statusCode ...int) *HttpClient {
	for _, code := range statusCode {
		statusText := http.StatusText(code)
		if len(statusText) == 0 {
			dhc.Errors = append(dhc.Errors, errors.New("StatusCode '"+strconv.Itoa(code)+"' doesn't exist in http package"))
		}
	}

	dhc.Retryable = struct {
		RetryableStatus []int
		RetryTime       time.Duration
		RetryCount      int
		Attempt         int
		Enable          bool
	}{
		statusCode,
		retryTime,
		retryCount,
		0,
		true,
	}
	return dhc
}

// SetBasicAuth sets the basic authentication header
// Example. To set the header for username "myuser" and password "mypass"
//
//    New()
//      Post("/gamelist").
//      SetBasicAuth("myuser", "mypass").
//      End()
func (dhc *HttpClient) SetBasicAuth(username string, password string) *HttpClient {
	dhc.BasicAuth = struct{ Username, Password string }{username, password}
	return dhc
}

// AddCookie adds a cookie to the request. The behavior is the same as AddCookie on Request from net/http
func (dhc *HttpClient) AddCookie(c *http.Cookie) *HttpClient {
	dhc.Cookies = append(dhc.Cookies, c)
	return dhc
}

// AddCookies is a convenient method to add multiple cookies
func (dhc *HttpClient) AddCookies(cookies []*http.Cookie) *HttpClient {
	dhc.Cookies = append(dhc.Cookies, cookies...)
	return dhc
}

var Types = map[string]string{
	TypeJSON:       "application/json",
	TypeXML:        "application/xml",
	TypeForm:       "application/x-www-form-urlencoded",
	TypeFormData:   "application/x-www-form-urlencoded",
	TypeUrlencoded: "application/x-www-form-urlencoded",
	TypeHTML:       "text/html",
	TypeText:       "text/plain",
	TypeMultipart:  "multipart/form-data",
}

// Type is a convenience function to specify the data type to send.
// For example, to send data as `application/x-www-form-urlencoded` :
//
//    New().
//      Post("/recipe").
//      Type("form").
//      Send(`{ "name": "egg benedict", "category": "brunch" }`).
//      End()
//
// This will POST the body "name=egg benedict&category=brunch" to url /recipe
//
// GoRequest supports
//
//    "text/html" uses "html"
//    "application/json" uses "json"
//    "application/xml" uses "xml"
//    "text/plain" uses "text"
//    "application/x-www-form-urlencoded" uses "urlencoded", "form" or "form-data"
//
func (dhc *HttpClient) Type(typeStr string) *HttpClient {
	if _, ok := Types[typeStr]; ok {
		dhc.ForceType = typeStr
	} else {
		dhc.Errors = append(dhc.Errors, errors.New("Type func: incorrect type \""+typeStr+"\""))
	}
	return dhc
}

// Query function accepts either json string or strings which will form a query-string in url of GET method or body of POST method.
// For example, making "/search?query=bicycle&size=50x50&weight=20kg" using GET method:
//
//      New().
//        Get("/search").
//        Query(`{ query: 'bicycle' }`).
//        Query(`{ size: '50x50' }`).
//        Query(`{ weight: '20kg' }`).
//        End()
//
// Or you can put multiple json values:
//
//      New().
//        Get("/search").
//        Query(`{ query: 'bicycle', size: '50x50', weight: '20kg' }`).
//        End()
//
// Strings are also acceptable:
//
//      New().
//        Get("/search").
//        Query("query=bicycle&size=50x50").
//        Query("weight=20kg").
//        End()
//
// Or even Mixed! :)
//
//      New().
//        Get("/search").
//        Query("query=bicycle").
//        Query(`{ size: '50x50', weight:'20kg' }`).
//        End()
//
func (dhc *HttpClient) Query(content interface{}) *HttpClient {
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		dhc.queryString(v.String())
	case reflect.Struct:
		dhc.queryStruct(v.Interface())
	case reflect.Map:
		dhc.queryMap(v.Interface())
	default:
	}
	return dhc
}

func (dhc *HttpClient) queryStruct(content interface{}) *HttpClient {
	if marshalContent, err := json.Marshal(content); err != nil {
		dhc.Errors = append(dhc.Errors, err)
	} else {
		var val map[string]interface{}
		if err := json.Unmarshal(marshalContent, &val); err != nil {
			dhc.Errors = append(dhc.Errors, err)
		} else {
			for k, v := range val {
				k = strings.ToLower(k)
				var queryVal string
				switch t := v.(type) {
				case string:
					queryVal = t
				case float64:
					queryVal = strconv.FormatFloat(t, 'f', -1, 64)
				case time.Time:
					queryVal = t.Format(time.RFC3339)
				default:
					j, err := json.Marshal(v)
					if err != nil {
						continue
					}
					queryVal = string(j)
				}
				dhc.QueryData.Add(k, queryVal)
			}
		}
	}
	return dhc
}

func (dhc *HttpClient) queryString(content string) *HttpClient {
	var val map[string]string
	if err := json.Unmarshal([]byte(content), &val); err == nil {
		for k, v := range val {
			dhc.QueryData.Add(k, v)
		}
	} else {
		if queryData, err := url.ParseQuery(content); err == nil {
			for k, queryValues := range queryData {
				for _, queryValue := range queryValues {
					dhc.QueryData.Add(k, string(queryValue))
				}
			}
		} else {
			dhc.Errors = append(dhc.Errors, err)
		}
		// TODO: need to check correct format of 'field=val&field=val&...'
	}
	return dhc
}

func (dhc *HttpClient) queryMap(content interface{}) *HttpClient {
	return dhc.queryStruct(content)
}

// As Go conventions accepts ; as a synonym for &. (https://github.com/golang/go/issues/2210)
// Thus, Query won't accept ; in a querystring if we provide something like fields=f1;f2;f3
// This Param is then created as an alternative method to solve this.
func (dhc *HttpClient) Param(key string, value string) *HttpClient {
	dhc.QueryData.Add(key, value)
	return dhc
}

// Set TLSClientConfig for underling Transport.
// One example is you can use it to disable security check (https):
//
//      New().TLSClientConfig(&tls.Config{ InsecureSkipVerify: true}).
//        Get("https://disable-security-check.com").
//        End()
//
func (dhc *HttpClient) TLSClientConfig(config *tls.Config) *HttpClient {
	dhc.Transport.TLSClientConfig = config
	return dhc
}

// Proxy function accepts a proxy url string to setup proxy url for any request.
// It provides a convenience way to setup proxy which have advantages over usual old ways.
// One example is you might try to set `http_proxy` environment. This means you are setting proxy up for all the requests.
// You will not be able to send different request with different proxy unless you change your `http_proxy` environment again.
// Another example is using Golang proxy setting. This is normal prefer way to do but too verbase compared to GoRequest'dhc Proxy:
//
//      New().Proxy("http://myproxy:9999").
//        Post("http://www.google.com").
//        End()
//
// To set no_proxy, just put empty string to Proxy func:
//
//      New().Proxy("").
//        Post("http://www.google.com").
//        End()
//
func (dhc *HttpClient) Proxy(proxyUrl string) *HttpClient {
	parsedProxyUrl, err := url.Parse(proxyUrl)
	if err != nil {
		dhc.Errors = append(dhc.Errors, err)
	} else if proxyUrl == "" {
		dhc.Transport.Proxy = nil
	} else {
		dhc.Transport.Proxy = http.ProxyURL(parsedProxyUrl)
	}
	return dhc
}

// RedirectPolicy accepts a function to define how to handle redirects. If the
// policy function returns an error, the next Request is not made and the previous
// request is returned.
//
// The policy function'dhc arguments are the Request about to be made and the
// past requests in order of oldest first.
func (dhc *HttpClient) RedirectPolicy(policy func(req Request, via []Request) error) *HttpClient {
	dhc.Client.CheckRedirect = func(r *http.Request, v []*http.Request) error {
		vv := make([]Request, len(v))
		for i, r := range v {
			vv[i] = Request(r)
		}
		return policy(Request(r), vv)
	}
	return dhc
}

// Send function accepts either json string or query strings which is usually used to assign data to POST or PUT method.
// Without specifying any type, if you give Send with json data, you are doing requesting in json format:
//
//      New().
//        Post("/search").
//        Send(`{ query: 'sushi' }`).
//        End()
//
// While if you use at least one of querystring, GoRequest understands and automatically set the Content-Type to `application/x-www-form-urlencoded`
//
//      New().
//        Post("/search").
//        Send("query=tonkatsu").
//        End()
//
// So, if you want to strictly send json format, you need to use Type func to set it as `json` (Please see more details in Type function).
// You can also do multiple chain of Send:
//
//      New().
//        Post("/search").
//        Send("query=bicycle&size=50x50").
//        Send(`{ wheel: '4'}`).
//        End()
//
// From v0.2.0, Send function provide another convenience way to work with Struct type. You can mix and match it with json and query string:
//
//      type BrowserVersionSupport struct {
//        Chrome string
//        Firefox string
//      }
//      ver := BrowserVersionSupport{ Chrome: "37.0.2041.6", Firefox: "30.0" }
//      New().
//        Post("/update_version").
//        Send(ver).
//        Send(`{"Safari":"5.1.10"}`).
//        End()
//
// If you have set Type to text or Content-Type to text/plain, content will be sent as raw string in body instead of form
//
//      New().
//        Post("/greet").
//        Type("text").
//        Send("hello world").
//        End()
//
func (dhc *HttpClient) Send(content interface{}) *HttpClient {
	// TODO: add normal text mode or other mode to Send func
	switch v := reflect.ValueOf(content); v.Kind() {
	case reflect.String:
		dhc.SendString(v.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // includes rune
		dhc.SendString(strconv.FormatInt(v.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // includes byte
		dhc.SendString(strconv.FormatUint(v.Uint(), 10))
	case reflect.Float64:
		dhc.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 64))
	case reflect.Float32:
		dhc.SendString(strconv.FormatFloat(v.Float(), 'f', -1, 32))
	case reflect.Bool:
		dhc.SendString(strconv.FormatBool(v.Bool()))
	case reflect.Struct:
		dhc.SendStruct(v.Interface())
	case reflect.Slice:
		dhc.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Array:
		dhc.SendSlice(makeSliceOfReflectValue(v))
	case reflect.Ptr:
		dhc.Send(v.Elem().Interface())
	case reflect.Map:
		dhc.SendMap(v.Interface())
	default:
		return dhc
	}
	return dhc
}

func makeSliceOfReflectValue(v reflect.Value) (slice []interface{}) {
	kind := v.Kind()
	if kind != reflect.Slice && kind != reflect.Array {
		return slice
	}

	slice = make([]interface{}, v.Len())
	for i := 0; i < v.Len(); i++ {
		slice[i] = v.Index(i).Interface()
	}

	return slice
}

// SendSlice (similar to SendString) returns HttpClient'dhc itself for any next chain and takes content []interface{} as a parameter.
// Its duty is to append slice of interface{} into dhc.SliceData ([]interface{}) which later changes into json array in the End() func.
func (dhc *HttpClient) SendSlice(content []interface{}) *HttpClient {
	dhc.SliceData = append(dhc.SliceData, content...)
	return dhc
}

func (dhc *HttpClient) SendMap(content interface{}) *HttpClient {
	return dhc.SendStruct(content)
}

// SendStruct (similar to SendString) returns HttpClient'dhc itself for any next chain and takes content interface{} as a parameter.
// Its duty is to transfrom interface{} (implicitly always a struct) into dhc.Data (map[string]interface{}) which later changes into appropriate format such as json, form, text, etc. in the End() func.
func (dhc *HttpClient) SendStruct(content interface{}) *HttpClient {
	if marshalContent, err := json.Marshal(content); err != nil {
		dhc.Errors = append(dhc.Errors, err)
	} else {
		var val map[string]interface{}
		d := json.NewDecoder(bytes.NewBuffer(marshalContent))
		d.UseNumber()
		if err := d.Decode(&val); err != nil {
			dhc.Errors = append(dhc.Errors, err)
		} else {
			for k, v := range val {
				dhc.Data[k] = v
			}
		}
	}
	return dhc
}

// SendString returns HttpClient'dhc itself for any next chain and takes content string as a parameter.
// Its duty is to transform String into dhc.Data (map[string]interface{}) which later changes into appropriate format such as json, form, text, etc. in the End func.
// Send implicitly uses SendString and you should use Send instead of this.
func (dhc *HttpClient) SendString(content string) *HttpClient {
	if !dhc.BounceToRawString {
		var val interface{}
		d := json.NewDecoder(strings.NewReader(content))
		d.UseNumber()
		if err := d.Decode(&val); err == nil {
			switch v := reflect.ValueOf(val); v.Kind() {
			case reflect.Map:
				for k, v := range val.(map[string]interface{}) {
					dhc.Data[k] = v
				}
			// add to SliceData
			case reflect.Slice:
				dhc.SendSlice(val.([]interface{}))
			// bounce to rawstring if it is arrayjson, or others
			default:
				dhc.BounceToRawString = true
			}
		} else if formData, err := url.ParseQuery(content); err == nil {
			for k, formValues := range formData {
				for _, formValue := range formValues {
					// make it array if already have key
					if val, ok := dhc.Data[k]; ok {
						var strArray []string
						strArray = append(strArray, string(formValue))
						// check if previous data is one string or array
						switch oldValue := val.(type) {
						case []string:
							strArray = append(strArray, oldValue...)
						case string:
							strArray = append(strArray, oldValue)
						}
						dhc.Data[k] = strArray
					} else {
						// make it just string if does not already have same key
						dhc.Data[k] = formValue
					}
				}
			}
			dhc.TargetType = TypeForm
		} else {
			dhc.BounceToRawString = true
		}
	}
	// Dump all contents to RawString in case in the end user doesn't want json or form.
	dhc.RawString += content
	return dhc
}

type File struct {
	Filename  string
	Fieldname string
	Data      []byte
}

// SendFile function works only with type "multipart". The function accepts one mandatory and up to two optional arguments. The mandatory (first) argument is the file.
// The function accepts a path to a file as string:
//
//      New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile("./example_file.ext").
//        End()
//
// File can also be a []byte slice of a already file read by eg. ioutil.ReadFile:
//
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b).
//        End()
//
// Furthermore file can also be a os.File:
//
//      f, _ := os.Open("./example_file.ext")
//      New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(f).
//        End()
//
// The first optional argument (second argument overall) is the filename, which will be automatically determined when file is a string (path) or a os.File.
// When file is a []byte slice, filename defaults to "filename". In all cases the automatically determined filename can be overwritten:
//
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b, "my_custom_filename").
//        End()
//
// The second optional argument (third argument overall) is the fieldname in the multipart/form-data request. It defaults to fileNUMBER (eg. file1), where number is ascending and starts counting at 1.
// So if you send multiple files, the fieldnames will be file1, file2, ... unless it is overwritten. If fieldname is set to "file" it will be automatically set to fileNUMBER, where number is the greatest exsiting number+1.
//
//      b, _ := ioutil.ReadFile("./example_file.ext")
//      New().
//        Post("http://example.com").
//        Type("multipart").
//        SendFile(b, "", "my_custom_fieldname"). // filename left blank, will become "example_file.ext"
//        End()
//
func (dhc *HttpClient) SendFile(file interface{}, args ...string) *HttpClient {

	filename := ""
	fieldname := "file"

	if len(args) >= 1 && len(args[0]) > 0 {
		filename = strings.TrimSpace(args[0])
	}
	if len(args) >= 2 && len(args[1]) > 0 {
		fieldname = strings.TrimSpace(args[1])
	}
	if fieldname == "file" || fieldname == "" {
		fieldname = "file" + strconv.Itoa(len(dhc.FileData)+1)
	}

	switch v := reflect.ValueOf(file); v.Kind() {
	case reflect.String:
		pathToFile, err := filepath.Abs(v.String())
		if err != nil {
			dhc.Errors = append(dhc.Errors, err)
			return dhc
		}
		if filename == "" {
			filename = filepath.Base(pathToFile)
		}
		data, err := ioutil.ReadFile(v.String())
		if err != nil {
			dhc.Errors = append(dhc.Errors, err)
			return dhc
		}
		dhc.FileData = append(dhc.FileData, File{
			Filename:  filename,
			Fieldname: fieldname,
			Data:      data,
		})
	case reflect.Slice:
		slice := makeSliceOfReflectValue(v)
		if filename == "" {
			filename = "filename"
		}
		f := File{
			Filename:  filename,
			Fieldname: fieldname,
			Data:      make([]byte, len(slice)),
		}
		for i := range slice {
			f.Data[i] = slice[i].(byte)
		}
		dhc.FileData = append(dhc.FileData, f)
	case reflect.Ptr:
		if len(args) == 1 {
			return dhc.SendFile(v.Elem().Interface(), args[0])
		}
		if len(args) >= 2 {
			return dhc.SendFile(v.Elem().Interface(), args[0], args[1])
		}
		return dhc.SendFile(v.Elem().Interface())
	default:
		if v.Type() == reflect.TypeOf(os.File{}) {
			osfile := v.Interface().(os.File)
			if filename == "" {
				filename = filepath.Base(osfile.Name())
			}
			data, err := ioutil.ReadFile(osfile.Name())
			if err != nil {
				dhc.Errors = append(dhc.Errors, err)
				return dhc
			}
			dhc.FileData = append(dhc.FileData, File{
				Filename:  filename,
				Fieldname: fieldname,
				Data:      data,
			})
			return dhc
		}

		dhc.Errors = append(dhc.Errors, errors.New("SendFile currently only supports either a string (path/to/file), a slice of bytes (file content itself), or a os.File!"))
	}

	return dhc
}

func changeMapToURLValues(data map[string]interface{}) url.Values {
	var newUrlValues = url.Values{}
	for k, v := range data {
		switch val := v.(type) {
		case string:
			newUrlValues.Add(k, val)
		case bool:
			newUrlValues.Add(k, strconv.FormatBool(val))
		// if a number, change to string
		// json.Number used to protect against a wrong (for GoRequest) default conversion
		// which always converts number to float64.
		// This type is caused by using Decoder.UseNumber()
		case json.Number:
			newUrlValues.Add(k, string(val))
		case int:
			newUrlValues.Add(k, strconv.FormatInt(int64(val), 10))
		// TODO add all other int-Types (int8, int16, ...)
		case float64:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		case float32:
			newUrlValues.Add(k, strconv.FormatFloat(float64(val), 'f', -1, 64))
		// following slices are mostly needed for tests
		case []string:
			for _, element := range val {
				newUrlValues.Add(k, element)
			}
		case []int:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatInt(int64(element), 10))
			}
		case []bool:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatBool(element))
			}
		case []float64:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		case []float32:
			for _, element := range val {
				newUrlValues.Add(k, strconv.FormatFloat(float64(element), 'f', -1, 64))
			}
		// these slices are used in practice like sending a struct
		case []interface{}:

			if len(val) <= 0 {
				continue
			}

			switch val[0].(type) {
			case string:
				for _, element := range val {
					newUrlValues.Add(k, element.(string))
				}
			case bool:
				for _, element := range val {
					newUrlValues.Add(k, strconv.FormatBool(element.(bool)))
				}
			case json.Number:
				for _, element := range val {
					newUrlValues.Add(k, string(element.(json.Number)))
				}
			}
		default:
			// TODO add ptr, arrays, ...
		}
	}
	return newUrlValues
}

// End is the most important function that you need to call when ending the chain. The request won't proceed without calling it.
// End function returns Response which matchs the structure of Response type in Golang'dhc http package (but without Body data). The body data itself returns as a string in a 2nd return value.
// Lastly but worth noticing, error array (NOTE: not just single error value) is returned as a 3rd value and nil otherwise.
//
// For example:
//
//    resp, body, err := New().Get("http://www.baidu.com").End()
//    if err != nil {
//      fmt.Println(err)
//    }
//    fmt.Println(resp, body)
//
// Moreover, End function also supports callback which you can put as a parameter.
// This extends the flexibility and makes GoRequest fun and clean! You can use GoRequest in whatever style you love!
//
// For example:
//
//    func printBody(resp Response, body string, errs []error){
//      fmt.Println(resp.Status)
//    }
//    New().Get("http://www.baidu.com").End(printBody)
//
func (dhc *HttpClient) End(callback ...func(response Response, body string, err error)) (Response, string, error) {
	var bytesCallback []func(response Response, body []byte, err error)
	if len(callback) > 0 {
		bytesCallback = []func(response Response, body []byte, err error){
			func(response Response, body []byte, err error) {
				callback[0](response, string(body), err)
			},
		}
	}

	resp, body, err := dhc.EndBytes(bytesCallback...)
	bodyString := string(body)

	return resp, bodyString, err
}

// EndBytes should be used when you want the body as bytes. The callbacks work the same way as with `End`, except that a byte array is used instead of a string.
func (dhc *HttpClient) EndBytes(callback ...func(response Response, body []byte, err error)) (Response, []byte, error) {
	var (
		err  error
		resp Response
		body []byte
	)

	for {
		resp, body, err = dhc.getResponseBytes()
		if err != nil {
			return nil, nil, err
		}
		if dhc.isRetryableRequest(resp) {
			resp.Header.Set("Retry-Count", strconv.Itoa(dhc.Retryable.Attempt))
			break
		}
	}

	respCallback := *resp
	if len(callback) != 0 {
		callback[0](&respCallback, body, dhc.marshalErrors())
	}
	return resp, body, nil
}

func (dhc *HttpClient) isRetryableRequest(resp Response) bool {
	if dhc.Retryable.Enable && dhc.Retryable.Attempt < dhc.Retryable.RetryCount && contains(resp.StatusCode, dhc.Retryable.RetryableStatus) {
		time.Sleep(dhc.Retryable.RetryTime)
		dhc.Retryable.Attempt++
		return false
	}
	return true
}

func contains(respStatus int, statuses []int) bool {
	for _, status := range statuses {
		if status == respStatus {
			return true
		}
	}
	return false
}

// EndStruct should be used when you want the body as a struct. The callbacks work the same way as with `End`, except that a struct is used instead of a string.
func (dhc *HttpClient) EndStruct(v interface{}, callback ...func(response Response, v interface{}, body []byte, err error)) (Response, []byte, error) {
	resp, body, errs := dhc.EndBytes()
	if errs != nil {
		return nil, body, errs
	}
	err := json.Unmarshal(body, &v)
	if err != nil {
		dhc.Errors = append(dhc.Errors, err)
		return resp, body, dhc.marshalErrors()
	}
	respCallback := *resp
	if len(callback) != 0 {
		callback[0](&respCallback, v, body, dhc.marshalErrors())
	}
	return resp, body, nil
}

func (dhc *HttpClient) getResponseBytes() (Response, []byte, error) {
	var (
		req  *http.Request
		err  error
		resp Response
	)

	sTime := time.Now()
	defer func() {
		cost := time.Now().Sub(sTime)
		pc.Cost(fmt.Sprintf("http_call_cost_%s", dhc.Url), cost)

		gl.Incr(fmt.Sprintf(glHttpClientCall, dhc.Url), 1)
		gl.IncrCost(fmt.Sprintf(glHttpClientCost, dhc.Url), cost)

		if err != nil {
			pc.CostFail(fmt.Sprintf("http_call_cost_fail_%v", dhc.Url), 1)
			gl.Incr(fmt.Sprintf(glHttpClientCallFail, dhc.Url), 1)
		}
	}()

	// check whether there is an error. if yes, return all errors
	if len(dhc.Errors) != 0 {
		return nil, nil, dhc.marshalErrors()
	}
	// check if there is forced type
	switch dhc.ForceType {
	case TypeJSON, TypeForm, TypeXML, TypeText, TypeMultipart:
		dhc.TargetType = dhc.ForceType
		// If force type is not set, check whether user set Content-Type header.
		// If yes, also bounce to the correct supported TargetType automatically.
	default:
		contentType := dhc.Header.Get("Content-Type")
		for k, v := range Types {
			if contentType == v {
				dhc.TargetType = k
			}
		}
	}

	traceId, ok := gl.Get(gl.LogId)
	if ok {
		dhc.SetHeader(TraceID, traceId.(string))
	}

	server, sok := gl.Get(gl.Server)
	key, kok := gl.Get(gl.SecretKey)
	if sok && kok {
		dstUrl, _ := url.ParseRequestURI(dhc.Url)
		gdToken := utls.GdEncode([]byte(fmt.Sprintf("%s__%s__%d", server, dstUrl.Path, time.Now().UnixNano()/1e6)), key.(string))
		dhc.SetHeader(GdTokenRaw, gdToken)
	}

	// if slice and map get mixed, let'dhc bounce to rawstring
	if len(dhc.Data) != 0 && len(dhc.SliceData) != 0 {
		dhc.BounceToRawString = true
	}

	// Make Request
	req, err = dhc.MakeRequest()
	if err != nil {
		dhc.Errors = append(dhc.Errors, err)
		return nil, nil, dhc.marshalErrors()
	}

	// Set Transport
	if !DisableTransportSwap {
		dhc.Client.Transport = dhc.Transport
	}

	// Log details of this request
	if dhc.Debug {
		dump, err := httputil.DumpRequest(req, true)
		if err != nil {
			dlog.Error("[http] Error:%v", err)
		} else {
			dlog.Info("[http] HTTP Request: %s", string(dump))
		}
	}

	// Display CURL command line
	if dhc.CurlCommand {
		curl, err := http2curl.GetCurlCommand(req)
		if err != nil {
			dlog.Error("getResponseBytes CURL command occur error:%s", err)
		} else {
			dlog.Info("CURL command :%v", curl)
		}
	}

	// Send request
	resp, err = dhc.Client.Do(req)
	if err != nil {
		dhc.Errors = append(dhc.Errors, err)
		return nil, nil, dhc.marshalErrors()
	}
	defer func() { _ = resp.Body.Close() }()

	// Log details of this response
	if dhc.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if nil != err {
			dlog.Error("http Error:%v", err)
		} else {
			dlog.Info("http HTTP Response: %dhc", string(dump))
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	// Reset resp.Body so it can be use again
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	if err != nil {
		return nil, nil, err
	}
	return resp, body, nil
}

func (dhc *HttpClient) MakeRequest() (*http.Request, error) {
	var (
		req           *http.Request
		contentType   string // This is only set when the request body content is non-empty.
		contentReader io.Reader
		err           error
	)

	if dhc.Method == "" {
		return nil, errors.New("No method specified")
	}

	// !!! Important Note !!!
	//
	// Throughout this region, contentReader and contentType are only set when
	// the contents will be non-empty.
	// This is done avoid ever sending a non-nil request body with nil contents
	// to http.NewRequest, because it contains logic which dependends on
	// whether or not the body is "nil".
	//
	switch dhc.TargetType {
	case TypeJSON:
		// If-case to give support to json array. we check if
		// 1) Map only: send it as json map from dhc.Data
		// 2) Array or Mix of map & array or others: send it as rawstring from dhc.RawString
		var contentJson []byte
		if dhc.BounceToRawString {
			contentJson = []byte(dhc.RawString)
		} else if len(dhc.Data) != 0 {
			contentJson, _ = json.Marshal(dhc.Data)
		} else if len(dhc.SliceData) != 0 {
			contentJson, _ = json.Marshal(dhc.SliceData)
		}
		if contentJson != nil {
			contentReader = bytes.NewReader(contentJson)
			contentType = "application/json"
		}
	case TypeForm, TypeFormData, TypeUrlencoded:
		var contentForm []byte
		if dhc.BounceToRawString || len(dhc.SliceData) != 0 {
			contentForm = []byte(dhc.RawString)
		} else {
			formData := changeMapToURLValues(dhc.Data)
			contentForm = []byte(formData.Encode())
		}
		if len(contentForm) != 0 {
			contentReader = bytes.NewReader(contentForm)
			contentType = "application/x-www-form-urlencoded"
		}
	case TypeText:
		if len(dhc.RawString) != 0 {
			contentReader = strings.NewReader(dhc.RawString)
			contentType = "text/plain"
		}
	case TypeXML:
		if len(dhc.RawString) != 0 {
			contentReader = strings.NewReader(dhc.RawString)
			contentType = "application/xml"
		}
	case TypeMultipart:
		var (
			buf = &bytes.Buffer{}
			mw  = multipart.NewWriter(buf)
		)

		if dhc.BounceToRawString {
			fieldName := dhc.Header.Get("data_fieldname")
			if fieldName == "" {
				fieldName = "data"
			}
			fw, _ := mw.CreateFormField(fieldName)
			fw.Write([]byte(dhc.RawString))
			contentReader = buf
		}

		if len(dhc.Data) != 0 {
			formData := changeMapToURLValues(dhc.Data)
			for key, values := range formData {
				for _, value := range values {
					fw, _ := mw.CreateFormField(key)
					fw.Write([]byte(value))
				}
			}
			contentReader = buf
		}

		if len(dhc.SliceData) != 0 {
			fieldName := dhc.Header.Get("json_fieldname")
			if fieldName == "" {
				fieldName = "data"
			}
			// copied from CreateFormField() in mime/multipart/writer.go
			h := make(textproto.MIMEHeader)
			fieldName = strings.Replace(strings.Replace(fieldName, "\\", "\\\\", -1), `"`, "\\\"", -1)
			h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%dhc"`, fieldName))
			h.Set("Content-Type", "application/json")
			fw, _ := mw.CreatePart(h)
			contentJson, err := json.Marshal(dhc.SliceData)
			if err != nil {
				return nil, err
			}
			fw.Write(contentJson)
			contentReader = buf
		}

		// add the files
		if len(dhc.FileData) != 0 {
			for _, file := range dhc.FileData {
				fw, _ := mw.CreateFormFile(file.Fieldname, file.Filename)
				fw.Write(file.Data)
			}
			contentReader = buf
		}

		// close before call to FormDataContentType ! otherwise its not valid multipart
		mw.Close()

		if contentReader != nil {
			contentType = mw.FormDataContentType()
		}
	default:
		// let'dhc return an error instead of an nil pointer exception here
		return nil, errors.New("TargetType '" + dhc.TargetType + "' could not be determined")
	}

	if req, err = http.NewRequest(dhc.Method, dhc.Url, contentReader); err != nil {
		return nil, err
	}
	for k, vals := range dhc.Header {
		for _, v := range vals {
			req.Header.Add(k, v)
		}

		if strings.EqualFold(k, "Host") {
			req.Host = vals[0]
		}
	}

	// Don't infer the content type header if an overrride is already provided.
	if len(contentType) != 0 && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", contentType)
	}

	// Add all querystring from Query func
	q := req.URL.Query()
	for k, v := range dhc.QueryData {
		for _, vv := range v {
			q.Add(k, vv)
		}
	}
	req.URL.RawQuery = q.Encode()

	// Add basic auth
	if dhc.BasicAuth != struct{ Username, Password string }{} {
		req.SetBasicAuth(dhc.BasicAuth.Username, dhc.BasicAuth.Password)
	}

	// Add cookies
	for _, cookie := range dhc.Cookies {
		req.AddCookie(cookie)
	}

	return req, nil
}

// AsCurlCommand returns a string representing the runnable `curl' command
// version of the request.
func (dhc *HttpClient) AsCurlCommand() (string, error) {
	req, err := dhc.MakeRequest()
	if err != nil {
		return "", err
	}
	cmd, err := http2curl.GetCurlCommand(req)
	if err != nil {
		return "", err
	}
	return cmd.String(), nil
}

func (dhc *HttpClient) marshalErrors() error {
	var err error
	el := len(dhc.Errors)
	if el > 0 {
		if el == 1 {
			err = dhc.Errors[0]
		} else {
			var errStrS []string
			for _, e := range dhc.Errors {
				if e != nil {
					errStrS = append(errStrS, e.Error())
				}
			}
			if len(errStrS) > 0 {
				err = errors.New(strings.Join(errStrS, "|"))
			}
		}
	}
	return err
}
