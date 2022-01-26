package tbhttp

import (
	"net/http"

	"github.com/robertkrimen/otto"
)

const (
	ProtoHttp  = "http://"
	ProtoHttps = "https://"
)

type TbHttp struct{}

func (tbhttp *TbHttp) DoRequest(call otto.FunctionCall) otto.Value {
	argc := len(call.ArgumentList)
	if argc != 1 {
		return otto.NullValue()
	}
	req, err := OttoValue2HttpReq(call.ArgumentList[0])
	if err != nil {
		return otto.NullValue()
	}

	client := &http.Client{}
	rowRsp, err := client.Do(req)
	if err != nil {
		return otto.NullValue()
	}
	defer rowRsp.Body.Close()

	rsp, err := HttpRsp2TbRsp(rowRsp)
	if err != nil {
		return otto.NullValue()
	}

	value, err := otto.New().ToValue(rsp)
	if err != nil {
		return otto.NullValue()
	}
	return value
}
