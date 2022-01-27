package frame

import (
	"net/http"
	"testing"

	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/frame/capal/tbhttp"
)

func (frame *Frame) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Path + req.Method
	plugin, ok := frame.pathPlugins[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	reqJS, err := tbhttp.HttpReq2TbReq(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rsp, err := plugin.Handle(reqJS)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(rsp.Status)
	w.Write([]byte(rsp.Body))
}

func TestFrame(t *testing.T) {
	config := &tigerbalm.Config{}
	config.Plugin.Path = "../js"
	config.Plugin.Log.Path = "/tmp/tigerbalm/log/plugin"
	config.Plugin.Log.Level = "debug"
	config.Plugin.Log.MaxSize = 10485760
	config.Plugin.Log.MaxRolls = 10
	config.Plugin.WatchPath = true
	config.Log.Level = "debug"
	config.Log.MaxSize = 10485760
	config.Log.MaxRolls = 10
	config.Log.File = "/tmp/tigerbalm/log/tigerbalm.log"

	frame, err := NewFrame(nil, config)
	if err != nil {
		t.Error(err)
		return
	}

	http.ListenAndServe("127.0.0.1:1202", frame)
}
