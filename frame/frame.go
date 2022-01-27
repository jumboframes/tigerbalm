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

	bus   bus.Bus
	capal *capal.Capal
}

func NewFrame(bus bus.Bus, config *tigerbalm.Config) (*Frame, error) {
	frame := &Frame{
		pathPlugins: make(map[string]*Plugin),
		namePlugins: make(map[string]*Plugin),
		bus:         bus,
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
		go frame.watchPlugin()
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
			file := filepath.Base(event.Name)
			if !strings.HasSuffix(file, ExtJS) {
				continue
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				// modification
				err := frame.reloadPlugin(file)
				if err != nil {
					tblog.Errorf("frame::watchplugin | plugin: %s modified, reload err: %s",
						file, err)
					continue
				}
				tblog.Debugf("frame::watchplugin | plugin: %s reload success", file)
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				// creation
				err := frame.loadPlugin(file)
				if err != nil {
					tblog.Errorf("frame::watchplugin | plugin: %s created, load err: %s",
						file, err)
					continue
				}
				tblog.Debugf("frame::watchplugin | plugin: %s load success", file)
			} else if event.Op&fsnotify.Remove == fsnotify.Remove ||
				event.Op&fsnotify.Rename == fsnotify.Rename {
				// remove or rename
				err := frame.unloadPlugin(file)
				if err != nil {
					tblog.Errorf("frame::watchplugin | plugin: %s removed, unload err: %s",
						file, err)
					continue
				}
				tblog.Debugf("frame::watchplugin | plugin: %s unload success", file)
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

func (frame *Frame) handleHTTP(data interface{}) {
	ctx, ok := data.(*bus.ContextHttp)
	if !ok {
		return
	}
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

func (frame *Frame) unloadPlugins() error {
	return nil
}

func (frame *Frame) loadPlugin(file string) error {
	name := strings.TrimSuffix(file, ExtJS)
	pluginName, err := filepath.Abs(filepath.Join(frame.config.Plugin.Path, file))
	if err != nil {
		return err
	}
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
	route := plugin.Path() + plugin.Method()
	frame.pluginMux.Lock()
	frame.pathPlugins[route] = plugin
	frame.pluginMux.Unlock()

	tblog.Debugf("frame::loadplugin | new plugin: %s success, method: %s, path: %s",
		file, plugin.method, plugin.path)
	if frame.bus != nil {
		frame.bus.AddSlotHandler(bus.SlotHttp, frame.handleHTTP,
			plugin.Method(), plugin.Path())
	}
	return nil
}

func (frame *Frame) unloadPlugin(file string) error {
	name := strings.TrimSuffix(file, ExtJS)
	frame.pluginMux.Lock()
	defer frame.pluginMux.Unlock()

	plugin, ok := frame.namePlugins[name]
	if ok {
		route := plugin.Path() + plugin.Method()
		delete(frame.namePlugins, name)
		delete(frame.pathPlugins, route)
		if frame.bus != nil {
			frame.bus.DelSlotHandler(bus.SlotHttp, plugin.Method(), plugin.Path())
		}
		plugin.Fini()
	} else {
		tblog.Errorf("frame::unloadplugin | unload a non-exist plugin: %s", name)
	}
	return nil
}

func (frame *Frame) reloadPlugin(file string) error {
	name := strings.TrimSuffix(file, ExtJS)
	pluginName := filepath.Join(frame.config.Plugin.Path, file)
	pluginCnt, err := ioutil.ReadFile(pluginName)
	if err != nil {
		tblog.Errorf("frame::reloadplugin | read plugin: %s err: %s",
			pluginName, err)
		return err
	}
	frame.pluginMux.Lock()
	defer frame.pluginMux.Unlock()

	plugin, ok := frame.namePlugins[name]
	if ok {
		// del slot before reload
		frame.pluginMux.Lock()
		route := plugin.Path() + plugin.Method()
		delete(frame.pathPlugins, route)
		frame.pluginMux.Unlock()
		if frame.bus != nil {
			frame.bus.DelSlotHandler(bus.SlotHttp, plugin.Method(), plugin.Path())
		}
		err = plugin.Reload(pluginCnt)
		if err != nil {
			return err
		}
		frame.pluginMux.Lock()
		route = plugin.Path() + plugin.Method()
		frame.pathPlugins[route] = plugin
		frame.pluginMux.Unlock()
		frame.bus.AddSlotHandler(bus.SlotHttp, frame.handleHTTP,
			plugin.Method(), plugin.Path())
	} else {
		tblog.Errorf("frame::unloadplugin | unload a non-exist plugin: %s", name)
	}
	return nil
}
