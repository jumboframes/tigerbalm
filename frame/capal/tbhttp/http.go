package tbhttp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

const (
	ProtocolHttp  = "http://"
	ProtocolHttps = "https://"
)

func DoRequest(call otto.FunctionCall) otto.Value {
	// method host uri query header body
	req := (*http.Request)(nil)
	err := error(nil)
	argc := len(call.ArgumentList)
	if argc < 3 || argc > 6 {
		return otto.NullValue()
	}
	switch argc {
	case 3:
		method := call.ArgumentList[0].String()
		host := call.ArgumentList[1].String()
		uri := call.ArgumentList[2].String()

		url := concat(ProtocolHttp, host, uri)
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			logrus.Errorf("DoRequest | http new request err: %s", err)
			return otto.NullValue()
		}

	case 4:
		method := call.ArgumentList[0].String()
		host := call.ArgumentList[1].String()
		uri := call.ArgumentList[2].String()
		query := call.ArgumentList[3].Object()

		qry := ""
		for index, key := range query.Keys() {
			value, err := query.Get(key)
			if err != nil {
				continue
			}
			if value.IsString() {
				if index == 0 {
					qry += "?" + key + "=" + value.String()
				} else {
					qry += "&" + key + "=" + value.String()
				}
			}
		}
		url := concat(ProtocolHttp, host, uri, qry)
		fmt.Println(url)
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			logrus.Errorf("DoRequest | http new request err: %s", err)
			return otto.NullValue()
		}

	case 5:
		method := call.ArgumentList[0].String()
		host := call.ArgumentList[1].String()
		uri := call.ArgumentList[2].String()
		query := call.ArgumentList[3].Object()

		qry := ""
		for index, key := range query.Keys() {
			value, err := query.Get(key)
			if err != nil {
				continue
			}
			if value.IsString() {
				if index == 0 {
					qry += "?" + key + "=" + value.String()
				} else {
					qry += "&" + key + "=" + value.String()
				}
			}
		}
		url := concat(ProtocolHttp, host, uri, qry)
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			logrus.Errorf("DoRequest | http new request err: %s", err)
			return otto.NullValue()
		}
		header := call.ArgumentList[4].Object()
		for _, key := range header.Keys() {
			value, err := header.Get(key)
			if err != nil {
				continue
			}
			if value.IsString() {
				req.Header.Set(key, value.String())
			}
		}

	case 6:
		method := call.ArgumentList[0].String()
		host := call.ArgumentList[1].String()
		uri := call.ArgumentList[2].String()
		query := call.ArgumentList[3].Object()
		header := call.ArgumentList[4].Object()
		body := call.ArgumentList[5].String()

		qry := ""
		for index, key := range query.Keys() {
			value, err := query.Get(key)
			if err != nil {
				continue
			}
			if value.IsString() {
				if index == 0 {
					qry += "?" + key + "=" + value.String()
				} else {
					qry += "&" + key + "=" + value.String()
				}
			}
		}
		url := concat(ProtocolHttp, host, uri, qry)
		bodier := io.Reader(nil)
		if body != "" {
			bodier = bytes.NewBuffer([]byte(body))
		}
		req, err = http.NewRequest(method, url, bodier)
		if err != nil {
			logrus.Errorf("DoRequest | http new request err: %s", err)
			return otto.NullValue()
		}
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
	data, _ := httputil.DumpRequest(req, false)
	fmt.Println(string(data))

	client := &http.Client{}
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
