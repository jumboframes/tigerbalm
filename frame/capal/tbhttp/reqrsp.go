package tbhttp

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/robertkrimen/otto"
)

var (
	ErrNoMethod = errors.New("no method")
)

type Request struct {
	Method string
	Host   string
	Url    string
	Query  map[string]string
	Header map[string]string
	Body   string
}

func HttpReq2TbReq(req *http.Request) (*Request, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	tbReq := &Request{
		Method: req.Method,
		Url:    req.URL.Path,
		Query:  map[string]string{},
		Header: map[string]string{},
		Host:   req.Host,
		Body:   string(body),
	}
	for k, v := range req.Header {
		tbReq.Header[k] = v[0]
	}
	for k, v := range req.URL.Query() {
		tbReq.Query[k] = v[0]
	}
	return tbReq, nil
}

type Response struct {
	Status int
	Header map[string]string
	Body   string
}

func HttpRsp2TbRsp(rsp *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	tbRsp := &Response{
		Status: rsp.StatusCode,
		Header: map[string]string{},
	}
	for k, v := range rsp.Header {
		tbRsp.Header[k] = v[0]
	}
	if body != nil {
		tbRsp.Body = string(body)
	} else {
		tbRsp.Body = ""
	}
	return tbRsp, nil
}

/*
{
	"Method": "GET",
	"Host": "www.baidu.com",
	"Path": "/",
	"Query": {
		"type": "movie",
	},
	"Header": {
		"X-REAL-IP": "192.168.180.56"
	},
	"Body": ""
}
*/
func OttoValue2HttpReq(req otto.Value) (*http.Request, error) {
	// method
	value, err := req.Object().Get("Method")
	if err != nil {
		return nil, err
	}
	method, err := value.ToString()
	if err != nil {
		return nil, err
	}
	// host
	value, err = req.Object().Get("Host")
	if err != nil {
		return nil, err
	}
	host, err := value.ToString()
	if err != nil {
		return nil, err
	}
	// path
	value, err = req.Object().Get("Path")
	if err != nil {
		return nil, err
	}
	path, err := value.ToString()
	if err != nil {
		return nil, err
	}
	url := concat(ProtoHttp, host, path)

	// query
	value, err = req.Object().Get("Query")
	if err != nil {
		return nil, err
	}
	if value.IsDefined() {
		object := value.Object()
		for index, key := range object.Keys() {
			query, err := object.Get(key)
			if err != nil {
				continue
			}
			if query.IsString() {
				if index == 0 {
					url += "?" + key + "=" + query.String()
				} else {
					url += "&" + key + "=" + query.String()
				}
			}
		}
	}
	// header
	header := http.Header{}
	value, err = req.Object().Get("Header")
	if err != nil {
		return nil, err
	}
	if value.IsDefined() {
		object := value.Object()
		for _, key := range object.Keys() {
			hdr, err := object.Get(key)
			if err != nil {
				continue
			}
			if hdr.IsString() {
				header.Set(key, hdr.String())
			}
		}
	}
	// body
	body := io.Reader(nil)
	value, err = req.Object().Get("Body")
	if err != nil {
		return nil, err
	}
	if value.IsDefined() {
		body = bytes.NewBuffer([]byte(value.String()))
	}

	// request
	httpReq, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	httpReq.Header = header
	return httpReq, nil
}

/*
{
	"status": 200,
	"header": {
		"Content-Type": "text/json"
	},
	"body": "{'foo': "bar"}"
}
*/
func OttoValue2TbRsp(rsp otto.Value) (*Response, error) {
	// status
	status := http.StatusOK
	value, err := rsp.Object().Get("Status")
	if err != nil {
		return nil, err
	}
	if value.IsDefined() {
		status64, err := value.ToInteger()
		if err != nil {
			return nil, err
		}
		status = int(status64)
	}
	// header
	header := map[string]string{}
	value, err = rsp.Object().Get("Header")
	if err != nil {
		return nil, err
	}
	if value.IsDefined() {
		for _, key := range value.Object().Keys() {
			v, err := value.Object().Get(key)
			if err != nil {
				continue
			}
			elem, err := v.ToString()
			if err != nil {
				continue
			}
			header[key] = elem
		}
	}
	// body
	body := ""
	value, err = rsp.Object().Get("Body")
	if err != nil {
		return nil, err
	}
	if value.IsDefined() {
		body, err = value.ToString()
		if err != nil {
			return nil, err
		}
	}
	return &Response{status, header, body}, nil
}

func TbRsp2OttoValue(rsp *Response) (otto.Value, error) {
	return otto.New().ToValue(rsp)
}
