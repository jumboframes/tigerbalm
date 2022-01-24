package tbhttp

import (
	"io"
	"io/ioutil"
	"net/http"

	"github.com/robertkrimen/otto"
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
	"status": 200,
	"header": {
		"X-REAL-IP": ["192.168.180.56"]
	},
	"body": "{'foo': "bar"}"
}
*/
func OttoValue2TbRsp(rsp otto.Value) (*Response, error) {
	// status
	status := http.StatusOK
	value, err := rsp.Object().Get("status")
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
	value, err = rsp.Object().Get("header")
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
	value, err = rsp.Object().Get("body")
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
