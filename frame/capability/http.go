package capability

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
	rsp, err := client.Do(req)
	if err != nil {
		log.Printf("DoRequest | client do err: %s", err)
		return otto.NullValue()
	}
	defer rsp.Body.Close()

	data, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		log.Printf("DoRequest | io read all err: %s", err)
		return otto.NullValue()
	}
	value, err := otto.ToValue(string(data))
	if err != nil {
		log.Printf("DoRequest | to value err: %s", err)
		return otto.NullValue()
	}
	return value
}
