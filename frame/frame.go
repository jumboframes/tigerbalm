package frame

import (
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/bus"
	"github.com/jumboframes/tigerbalm/frame/capal"
	"github.com/jumboframes/tigerbalm/frame/capal/tbhttp"
	"github.com/jumboframes/tigerbalm/frame/capal/tbkafka"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
)

const (
	ExtJS  = ".js"
	ExtLog = ".log"
)

type Frame struct {
	namePlugins   map[string]*Plugin
	httpPlugins   map[string]*Plugin
	kafkaPlugins  map[string]*Plugin
	pluginMux     sync.RWMutex
	pluginWatcher *fsnotify.Watcher

	bus   bus.Bus
	capal *capal.Capal
}

func NewFrame(bus bus.Bus) (*Frame, error) {
	frame := &Frame{
		httpPlugins: make(map[string]*Plugin),
		namePlugins: make(map[string]*Plugin),
		bus:         bus,
	}
	if tigerbalm.Conf.Kafka.Enable {
		frame.kafkaPlugins = make(map[string]*Plugin)
	}
	frame.capal = capal.NewCapal(frame.httpFactory, frame.logFactory)
	err := frame.loadPlugins()
	if err != nil {
		return nil, err
	}
	if tigerbalm.Conf.Plugin.WatchPath {
		frame.pluginWatcher, err = fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		err = frame.pluginWatcher.Add(tigerbalm.Conf.Plugin.Path)
		if err != nil {
			return nil, err
		}
		go frame.watchPlugin()
	}
	return frame, nil
}

func (frame *Frame) Notify(os.Signal) {
	err := frame.unloadPlugins()
	if err != nil {
		tblog.Errorf("frame::notify | unload plugins err: %s", err)
		return
	}
	err = frame.loadPlugins()
	if err != nil {
		tblog.Errorf("frame::notify | load plugins err: %s", err)
		return
	}
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
	if tigerbalm.Conf.Plugin.WatchPath {
		frame.pluginWatcher.Close()
	}
	frame.pluginMux.RLock()
	defer frame.pluginMux.RUnlock()

	for _, plugin := range frame.namePlugins {
		plugin.Fini()
		frame.unregisterHttp(plugin)
		frame.unregisterKafka(plugin)
	}
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
	files, err := ioutil.ReadDir(tigerbalm.Conf.Plugin.Path)
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
	frame.pluginMux.Lock()
	defer frame.pluginMux.Unlock()

	for _, plugin := range frame.namePlugins {
		plugin.Fini()
		delete(frame.namePlugins, plugin.Name())
		frame.unregisterHttp(plugin)
		frame.unregisterKafka(plugin)
	}
	return nil
}

func (frame *Frame) loadPlugin(file string) error {
	name := strings.TrimSuffix(file, ExtJS)
	pluginName, err := filepath.Abs(filepath.Join(tigerbalm.Conf.Plugin.Path, file))
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
	plugin, err := NewPlugin(name, pluginCnt, frame.capal)
	if err != nil {
		tblog.Errorf("frame::loadplugin | new plugin: %s err: %s",
			pluginName, err)
		return err
	}
	frame.pluginMux.Lock()
	frame.namePlugins[name] = plugin
	frame.pluginMux.Unlock()

	// load must be called after NewPlugin, since vm require log from log factory
	err = plugin.Load()
	if err != nil {
		tblog.Errorf("frame::loadplugin | plugin: %s load err: %s",
			pluginName, err)
		return err
	}
	frame.registerHttp(plugin)
	frame.registerKafka(plugin)
	return nil
}

func (frame *Frame) unloadPlugin(file string) error {
	name := strings.TrimSuffix(file, ExtJS)
	frame.pluginMux.Lock()
	defer frame.pluginMux.Unlock()

	plugin, ok := frame.namePlugins[name]
	if ok {
		plugin.Fini()
		delete(frame.namePlugins, name)
		frame.unregisterHttp(plugin)
		frame.unregisterKafka(plugin)
	} else {
		tblog.Errorf("frame::unloadplugin | unload a non-exist plugin: %s", name)
	}
	return nil
}

func (frame *Frame) reloadPlugin(file string) error {
	name := strings.TrimSuffix(file, ExtJS)
	pluginName := filepath.Join(tigerbalm.Conf.Plugin.Path, file)
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
		frame.unregisterHttp(plugin)
		frame.unregisterKafka(plugin)
		err = plugin.Reload(pluginCnt)
		if err != nil {
			return err
		}
		frame.registerHttp(plugin)
		frame.registerKafka(plugin)
	} else {
		tblog.Errorf("frame::unloadplugin | unload a non-exist plugin: %s", name)
	}
	return nil
}

func (frame *Frame) registerHttp(plugin *Plugin) {
	if !plugin.Http() {
		return
	}
	route := plugin.HttpPath() + plugin.HttpMethod()
	frame.httpPlugins[route] = plugin
	if frame.bus != nil {
		frame.bus.AddSlotHandler(bus.SlotHttp, httpHandlerFactory(plugin),
			plugin.HttpMethod(), plugin.HttpPath())
		tblog.Debugf("frame::registerhttp | plugin: %s, method: %s, path: %s",
			plugin.Name(), plugin.HttpMethod(), plugin.HttpPath())
	}
}

func (frame *Frame) unregisterHttp(plugin *Plugin) {
	if !plugin.Http() {
		return
	}
	route := plugin.HttpPath() + plugin.HttpMethod()
	delete(frame.httpPlugins, route)
	if frame.bus != nil {
		frame.bus.DelSlotHandler(bus.SlotHttp,
			plugin.HttpMethod(), plugin.HttpPath())
		tblog.Debugf("frame::unregisterhttp | plugin: %s, method: %s, path: %s",
			plugin.Name(), plugin.HttpMethod(), plugin.HttpPath())
	}
}

func (frame *Frame) registerKafka(plugin *Plugin) {
	if !plugin.Kafka() || !tigerbalm.Conf.Kafka.Enable {
		return
	}
	tp := plugin.KafkaTopic() + plugin.KafkaGroup()
	frame.kafkaPlugins[tp] = plugin
	if frame.bus != nil {
		frame.bus.AddSlotHandler(bus.SlotKafka, kafkaHandlerFactory(plugin),
			plugin.KafkaTopic(), plugin.KafkaGroup())
		tblog.Debugf("frame::registerkafka | plugin: %s, topic: %s, group: %s",
			plugin.Name(), plugin.KafkaTopic(), plugin.KafkaGroup())
	}
}

func (frame *Frame) unregisterKafka(plugin *Plugin) {
	if !plugin.Kafka() || !tigerbalm.Conf.Kafka.Enable {
		return
	}
	tp := plugin.KafkaTopic() + plugin.KafkaGroup()
	delete(frame.kafkaPlugins, tp)
	if frame.bus != nil {
		frame.bus.DelSlotHandler(bus.SlotKafka,
			plugin.KafkaTopic(), plugin.KafkaGroup())
		tblog.Debugf("frame::unregisterkafka | plugin: %s, topic: %s, group: %s",
			plugin.Name(), plugin.KafkaTopic(), plugin.KafkaGroup())
	}
}

func httpHandlerFactory(plugin *Plugin) func(data interface{}) {
	return func(data interface{}) {
		ctx, ok := data.(*bus.ContextHttp)
		if !ok {
			return
		}

		reqJS, err := tbhttp.HttpReq2TbReq(ctx.Request())
		if err != nil {
			ctx.ResponseWriter().WriteHeader(http.StatusBadRequest)
			return
		}
		rsp, err := plugin.HttpHandle(reqJS)
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
}

func kafkaHandlerFactory(plugin *Plugin) func(data interface{}) {
	return func(data interface{}) {
		cgmsg, ok := data.(*tbkafka.ConsumerGroupMessage)
		if !ok {
			return
		}
		tbmsg, _ := tbkafka.CGMessage2TbCGMessage(cgmsg)
		plugin.KafkaHandle(tbmsg)
	}
}
