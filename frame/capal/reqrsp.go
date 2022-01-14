package capal

import (
	"io/ioutil"
	"net/http"
	"net/url"
)

type Request struct {
	Method string
	Host   string
	URL    *url.URL
	Header http.Header
	Body   string
}

func HttpReq2Req(req *http.Request) (*Request, error) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	newReq := &Request{
		Method: req.Method,
		URL:    req.URL,
		Header: req.Header,
		Host:   req.Host,
		Body:   string(body),
	}
	return newReq, nil
}

type Response struct {
	Status int
	Body   string
}

func HttpRsp2Rsp(rsp *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	newRsp := &Response{
		Status: rsp.StatusCode,
		Body:   string(body),
	}
	return newRsp, nil
}
