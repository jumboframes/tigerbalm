package tbhttp

import (
	"io"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Method string
	Host   string
	Url    string
	Query  map[string]string
	Header map[string]string
	Body   string
}

func HttpReq2Req(req *http.Request) (*Request, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	newReq := &Request{
		Method: req.Method,
		Url:    req.URL.Path,
		Query:  map[string]string{},
		Header: map[string]string{},
		Host:   req.Host,
		Body:   string(body),
	}
	for k, v := range req.Header {
		newReq.Header[k] = v[0]
	}
	for k, v := range req.URL.Query() {
		newReq.Query[k] = v[0]
	}
	return newReq, nil
}

type Response struct {
	Status int
	Header map[string]string
	Body   string
}

func HttpRsp2Rsp(rsp *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	newRsp := &Response{
		Status: rsp.StatusCode,
		Header: map[string]string{},
	}
	for k, v := range rsp.Header {
		newRsp.Header[k] = v[0]
	}
	if body != nil {
		newRsp.Body = string(body)
	} else {
		newRsp.Body = ""
	}
	return newRsp, nil
}
