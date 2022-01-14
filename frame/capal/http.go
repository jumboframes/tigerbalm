package capal

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/robertkrimen/otto"
)

func DoRequest(call otto.FunctionCall) otto.Value {
	host := call.ArgumentList[0].String()
	uri := call.ArgumentList[1].String()
	method := call.ArgumentList[2].String()
	body := call.ArgumentList[3].String()

	url := fmt.Sprintf("http://%s%s", host, uri)
	bodier := io.Reader(nil)
	if body != "" {
		bodier = bytes.NewBuffer([]byte(body))
	}
	req, err := http.NewRequest(method, url, bodier)
	if err != nil {
		log.Printf("DoRequest | new request err: %s", err)
		return otto.NullValue()
	}

	client := &http.Client{}
	rowRsp, err := client.Do(req)
	if err != nil {
		log.Printf("DoRequest | client do err: %s", err)
		return otto.NullValue()
	}
	defer rowRsp.Body.Close()
	rsp, err := HttpRsp2Rsp(rowRsp)
	if err != nil {
		log.Printf("DoRequest | http rsp to rsp err: %s", err)
		return otto.NullValue()
	}

	value, err := otto.New().ToValue(rsp)
	if err != nil {
		log.Printf("DoRequest | otto to value err: %s", err)
		return otto.NullValue()
	}
	return value
}