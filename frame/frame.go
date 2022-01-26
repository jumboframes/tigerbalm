package frame

import (
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/jumboframes/tigerbalm/frame/capal"
	"github.com/jumboframes/tigerbalm/frame/capal/tbhttp"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
)

const (
	ExtJS  = ".js"
	ExtLog = ".log"
)

type Frame struct {
	config        *tigerbalm.Config
	pathPlugins   map[string]*Plugin
	namePlugins   map[string]*Plugin
	pluginMux     sync.RWMutex
	pluginWatcher *fsnotify.Watcher

	bus   *bus.Bus
	capal *capal.Capal
}

func NewFrame(config *tigerbalm.Config) (*Frame, error) {
	frame := &Frame{
		pathPlugins: make(map[string]*Plugin),
		namePlugins: make(map[string]*Plugin),
		config:      config,
	}
	frame.capal = capal.NewCapal(frame.httpFactory, frame.logFactory)
	err := frame.loadPlugins()
	if err != nil {
		return nil, err
	}
	if config.Plugin.WatchPath {
		frame.pluginWatcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		err = frame.pluginWatcher.Add(config.Plugin.Path)
		if err != nil {
			return nil, err
		}
	}
	return frame, nil
}

func (frame *Frame) watchPlugin() {
	for {
		select {
		case event, ok := <-frame.pluginWatcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// modification
			}
			if event.Op&fsnotify.Create == fsnotify.Create {
				// creation
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				// remove
			}
		}
	}
}

func (frame *Frame) Fini() {
	frame.pluginWatcher.Close()
	frame.pluginMux.RLock()
	defer frame.pluginMux.RUnlock()
	for _, plugin := range frame.namePlugins {
		plugin.Fini()
	}
}

func (frame *Frame) HandleHTTP(ctx *bus.Context) {
	key := ctx.RelativePath + ctx.Method()
	frame.pluginMux.RLock()
	plugin, ok := frame.pathPlugins[key]
	if !ok {
		ctx.ResponseWriter().WriteHeader(http.StatusNotFound)
		frame.pluginMux.RUnlock()
		return
	}
	frame.pluginMux.RUnlock()

	reqJS, err := tbhttp.HttpReq2TbReq(ctx.Request())
	if err != nil {
		ctx.ResponseWriter().WriteHeader(http.StatusBadRequest)
		return
	}
	rsp, err := plugin.Handle(reqJS)
	if err != nil {
		tblog.Errorf("frame::handlehttp | plugin handle err: %s", err)
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

func (frame *Frame) httpFactory(ctx *capal.PluginContext) *tbhttp.TbHttp {
	return &tbhttp.TbHttp{}
}

func (frame *Frame) logFactory(ctx *capal.PluginContext) *tblog.TbLog {
	frame.pluginMux.RLock()
	defer frame.pluginMux.RUnlock()

	plugin, ok := frame.namePlugins[ctx.Name]
	if ok {
		return plugin.Log()
	}
	tblog.Errorf("frame::logfactory | get nil log from factory, name: %s", ctx.Name)
	return nil
}

func (frame *Frame) loadPlugins() error {
	files, err := ioutil.ReadDir(frame.config.Plugin.Path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if !file.IsDir() && path.Ext(file.Name()) == ExtJS {
			err = frame.loadPlugin(file.Name())
			if err != nil {
				tblog.Errorf("frame::loadplugins | load plugin err: %s", err)
			}
		}
	}
	return nil
}

func (frame *Frame) loadPlugin(file string) error {
	name := strings.TrimSuffix(file, ExtJS)
	pluginName := filepath.Join(frame.config.Plugin.Path, file)
	pluginCnt, err := ioutil.ReadFile(pluginName)
	if err != nil {
		tblog.Errorf("frame::loadplugin | read plugin: %s err: %s",
			pluginName, err)
		return err
	}
	// new plugin
	plugin, err := NewPlugin(name, pluginCnt, frame.capal, frame.config)
	if err != nil {
		tblog.Errorf("frame::loadplugin | new plugin: %s err: %s",
			pluginName, err)
		return err
	}
	frame.pluginMux.Lock()
	frame.namePlugins[name] = plugin
	frame.pluginMux.Unlock()

	err = plugin.Load()
	if err != nil {
		tblog.Errorf("frame::loadplugin | plugin: %s load err: %s",
			pluginName, err)
		return err
	}
	frame.pluginMux.Lock()
	frame.pathPlugins[plugin.Path()+plugin.Method()] = plugin
	frame.pluginMux.Unlock()

	tblog.Debugf("frame::loadplugin | new plugin: %s success, method: %s, path: %s",
		pluginName, plugin.method, plugin.path)
	// frame.bus.RegisterHttp(plugin.Method(), plugin.Url(), frame.HandleHTTP)
	return nil
}
