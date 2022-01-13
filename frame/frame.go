package frame

import (
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"

	"github.com/jumboframes/tigerbalm/errors"
	"github.com/jumboframes/tigerbalm/frame/capability"
	"github.com/robertkrimen/otto"
)

const (
	ExtJS = ".js"
)

const (
	RegisterFunc = "register"
)

type Frame struct {
	plugins   map[string]*Plugin
	pluginDir string
}

func NewFrame(pluginDir string) (*Frame, error) {
	frame := &Frame{
		plugins:   make(map[string]*Plugin),
		pluginDir: pluginDir,
	}
	err := frame.loadPlugins()
	if err != nil {
		return nil, err
	}
	return frame, nil
}

func (frame *Frame) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reqJS, err := httpReq2Req(req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := req.URL.Path + req.Method
	plugin, ok := frame.plugins[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	rsp, err := plugin.Handle(reqJS)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(rsp)
}

func (frame *Frame) loadPlugins() error {
	files, err := ioutil.ReadDir(frame.pluginDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && path.Ext(file.Name()) == ExtJS {
			err = frame.loadPlugin(file.Name())
			if err != nil {
				log.Printf("Frame::loadPlugins | load plugin err: %s", err)
			}
		}
	}
	return nil
}

func (frame *Frame) loadPlugin(file string) error {
	pluginCnt, err := readPlugin(frame.pluginDir, file)
	if err != nil {
		return err
	}
	vm := otto.New()
	_, err = vm.Run(pluginCnt)
	if err != nil {
		return err
	}
	routeValue, err := vm.Call(RegisterFunc, nil)
	if err != nil {
		return err
	}
	if !routeValue.IsObject() {
		return errors.ErrIllegalRegister
	}
	route, err := getRoute(routeValue.Object())
	if err != nil {
		return err
	}
	err = vm.Set("doRequest", capability.DoRequest)
	if err != nil {
		return err
	}
	plugin := &Plugin{
		Name:    file,
		Url:     route.url,
		Method:  route.method,
		Handler: route.handler,
	}
	frame.plugins[route.url+route.method] = plugin
	return nil
}

type route struct {
	url, method string
	handler     otto.Value
}

func getRoute(obj *otto.Object) (*route, error) {
	matchValue, err := obj.Get("match")
	if err != nil {
		return nil, err
	}
	urlValue, err := matchValue.Object().Get("url")
	if err != nil {
		return nil, err
	}
	url, err := urlValue.ToString()
	if err != nil {
		return nil, err
	}
	methodValue, err := matchValue.Object().Get("method")
	if err != nil {
		return nil, err
	}
	method, err := methodValue.ToString()
	if err != nil {
		return nil, err
	}

	handler, err := obj.Get("handler")
	if err != nil {
		return nil, err
	}
	if !handler.IsFunction() {
		return nil, errors.ErrRegisterNotFunction
	}
	route := &route{
		url:     url,
		method:  method,
		handler: handler,
	}
	return route, nil
}

func readPlugin(dir, file string) ([]byte, error) {
	filepath := filepath.Join(dir, file)
	return ioutil.ReadFile(filepath)
}
