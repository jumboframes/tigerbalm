package http

import (
	"bytes"
	"io"
	nhttp "net/http"
	"strings"

	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

const (
	ProtocolHttp  = "http://"
	ProtocolHttps = "https://"
)

func DoRequest(call otto.FunctionCall) otto.Value {
	// method host uri header body
	req := (*nhttp.Request)(nil)
	err := error(nil)
	argc := len(call.ArgumentList)
	if argc < 3 || argc > 5 {
		return otto.NullValue()
	}
	switch argc {
	case 3:
		method := call.ArgumentList[0].String()
		host := call.ArgumentList[1].String()
		uri := call.ArgumentList[2].String()

		url := concat(ProtocolHttp, host, uri)
		req, err = nhttp.NewRequest(method, url, nil)
		if err != nil {
			logrus.Errorf("DoRequest | http new request err: %s", err)
			return otto.NullValue()
		}

	case 4:
		method := call.ArgumentList[0].String()
		host := call.ArgumentList[1].String()
		uri := call.ArgumentList[2].String()

		url := concat(ProtocolHttp, host, uri)
		req, err = nhttp.NewRequest(method, url, nil)
		if err != nil {
			logrus.Errorf("DoRequest | http new request err: %s", err)
			return otto.NullValue()
		}
		header := call.ArgumentList[3].Object()
		for _, key := range header.Keys() {
			value, err := header.Get(key)
			if err != nil {
				continue
			}
			if value.IsString() {
				req.Header.Set(key, value.String())
			}
		}

	case 5:
		method := call.ArgumentList[0].String()
		host := call.ArgumentList[1].String()
		uri := call.ArgumentList[2].String()
		body := call.ArgumentList[4].String()

		bodier := io.Reader(nil)
		if body != "" {
			bodier = bytes.NewBuffer([]byte(body))
		}
		url := concat(ProtocolHttp, host, uri)
		req, err = nhttp.NewRequest(method, url, bodier)
		if err != nil {
			logrus.Errorf("DoRequest | http new request err: %s", err)
			return otto.NullValue()
		}
		header := call.ArgumentList[3].Object()
		for _, key := range header.Keys() {
			value, err := header.Get(key)
			if err != nil {
				continue
			}
			if value.IsString() {
				req.Header.Set(key, value.String())
			}
		}
	}

	client := &nhttp.Client{}
	rowRsp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("DoRequest | client do err: %s", err)
		return otto.NullValue()
	}
	defer rowRsp.Body.Close()

	rsp, err := HttpRsp2Rsp(rowRsp)
	if err != nil {
		logrus.Errorf("DoRequest | http rsp to rsp err: %s", err)
		return otto.NullValue()
	}

	value, err := otto.New().ToValue(rsp)
	if err != nil {
		logrus.Errorf("DoRequest | otto to value err: %s", err)
		return otto.NullValue()
	}
	return value
}

func concat(substrs ...string) string {
	builder := new(strings.Builder)
	for _, substr := range substrs {
		builder.WriteString(substr)
	}
	return builder.String()
}
