package capal

import (
	"io"
	"io/ioutil"
	"net/http"
)

type Request struct {
	Method string
	Host   string
	Url    string
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
		Header: map[string]string{},
		Host:   req.Host,
		Body:   string(body),
	}
	for k, v := range req.Header {
		newReq.Header[k] = v[0]
	}
	return newReq, nil
}

type Response struct {
	Status int
	Body   string
}

func HttpRsp2Rsp(rsp *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil && err != io.EOF {
		return nil, err
	}

	newRsp := &Response{
		Status: rsp.StatusCode,
	}
	if body != nil {
		newRsp.Body = string(body)
	} else {
		newRsp.Body = ""
	}
	return newRsp, nil
}
