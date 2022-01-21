package frame

import (
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/jumboframes/tigerbalm/frame/capal/tbhttp"
	"github.com/sirupsen/logrus"
)

const (
	ExtJS = ".js"
)

type Frame struct {
	pluginDir string
	pluginMux sync.RWMutex
	plugins   map[string]*Plugin
	bus       *bus.Bus

	fsWatcher *fsnotify.Watcher
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

func (frame *Frame) HandleHTTP(ctx *bus.Context) {
	key := ctx.RelativePath + ctx.Method()
	frame.pluginMux.RLock()
	plugin, ok := frame.plugins[key]
	if !ok {
		ctx.ResponseWriter().WriteHeader(http.StatusNotFound)
		frame.pluginMux.RUnlock()
		return
	}
	frame.pluginMux.RUnlock()

	reqJS, err := tbhttp.HttpReq2Req(ctx.Request())
	if err != nil {
		ctx.ResponseWriter().WriteHeader(http.StatusBadRequest)
		return
	}
	rsp, err := plugin.Handle(reqJS)
	if err != nil {
		logrus.Errorf("Frame::HandleHTTP | plugin handle err: %s", err)
		ctx.ResponseWriter().WriteHeader(http.StatusInternalServerError)
		return
	}
	header := ctx.ResponseWriter().Header()
	for k, v := range rsp.Header {
		if strings.ToLower(k) != strings.ToLower("Content-Length") {
			header.Set(k, v)
		}
	}
	ctx.ResponseWriter().WriteHeader(rsp.Status)
	ctx.ResponseWriter().Write([]byte(rsp.Body))
}

func (frame *Frame) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	key := req.URL.Path + req.Method
	plugin, ok := frame.plugins[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	reqJS, err := tbhttp.HttpReq2Req(req)
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

	plugin, err := NewPlugin(pluginCnt)
	if err != nil {
		return err
	}
	frame.pluginMux.Lock()
	frame.plugins[plugin.Url()+plugin.Method()] = plugin
	frame.pluginMux.Unlock()
	frame.bus.RegisterHttp(plugin.Method(), plugin.Url(), frame.HandleHTTP)
	return nil
}

func readPlugin(dir, file string) ([]byte, error) {
	filepath := filepath.Join(dir, file)
	return ioutil.ReadFile(filepath)
}
