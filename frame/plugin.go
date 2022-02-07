package frame

import (
	"path/filepath"
	"sync"

	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/frame/capal"
	"github.com/jumboframes/tigerbalm/frame/capal/tbhttp"
	"github.com/jumboframes/tigerbalm/frame/capal/tbkafka"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/robertkrimen/otto"
)

type Plugin struct {
	mu sync.RWMutex
	// metas
	http       bool
	httpPath   string
	httpMethod string

	kafka      bool
	kafkaTopic string
	kafkaGroup string

	name    string
	content []byte

	// conf
	config *tigerbalm.Config

	// capals
	capal *capal.Capal
	ctx   *capal.PluginContext

	// runtimes
	httpPool  sync.Pool
	kafkaPool sync.Pool

	// logs
	log       *tblog.TbLog
	rotateLog *rotatelogs.RotateLogs
}

// plugin uses name as runtime context indexing for looking-up resources,
// like logging instance. the changing of name must then recreate ctx.
func NewPlugin(name string, content []byte,
	cpl *capal.Capal, config *tigerbalm.Config) (*Plugin, error) {

	plugin := &Plugin{
		name:    name,
		content: content,
		capal:   cpl,
		config:  config,
		ctx:     &capal.PluginContext{name},
	}
	err := plugin.newLog()
	if err != nil {
		return nil, err
	}
	return plugin, nil
}

func (plugin *Plugin) newLog() error {
	if plugin.rotateLog != nil {
		plugin.rotateLog.Close()
	}
	level, err := tblog.ParseLevel(plugin.config.Plugin.Log.Level)
	if err != nil {
		tblog.Errorf("newlog | tblog parse level: %s err: %s",
			plugin.config.Plugin.Log.Level, err)
		return err
	}
	logFile := filepath.Join(plugin.config.Plugin.Log.Path,
		plugin.name, plugin.name+ExtLog)
	rotateLog, err := rotatelogs.New(logFile,
		rotatelogs.WithRotationCount(plugin.config.Plugin.Log.MaxRolls),
		rotatelogs.WithRotationSize(plugin.config.Plugin.Log.MaxSize))
	if err != nil {
		tblog.Errorf("newplog | rotate log: %s new err: %s",
			logFile, err)
		return err
	}
	log := tblog.NewTbLog().WithLevel(level).WithOutput(rotateLog)
	plugin.log = log
	plugin.rotateLog = rotateLog
	return nil
}

func (plugin *Plugin) Load() error {
	// plugin runtime
	runtime, err := plugin.vmFactory()
	if err != nil {
		tblog.Errorf("newplugin | plugin: %s, get vm err: %s",
			plugin.name, err)
		return err
	}
	plugin.mu.Lock()
	if runtime.route != nil {
		httpPool := sync.Pool{
			New: plugin.httpHandlerFactory,
		}
		httpPool.Put(runtime.route.handler)
		plugin.http = true
		plugin.httpPath = runtime.route.path
		plugin.httpMethod = runtime.route.method
		plugin.httpPool = httpPool
	}
	if runtime.consume != nil {
		kafkaPool := sync.Pool{
			New: plugin.kafkaHandlerFactory,
		}
		kafkaPool.Put(runtime.consume.handler)
		plugin.kafka = true
		plugin.kafkaTopic = runtime.consume.topic
		plugin.kafkaGroup = runtime.consume.group
		plugin.kafkaPool = kafkaPool
	}
	plugin.mu.Unlock()
	return nil
}

func (plugin *Plugin) Name() string {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.name
}

func (plugin *Plugin) Rename(name string) error {
	plugin.mu.Lock()
	defer plugin.mu.Unlock()

	plugin.name = name
	plugin.ctx = &capal.PluginContext{name}
	return plugin.newLog()
}

func (plugin *Plugin) Reload(content []byte) error {
	plugin.mu.Lock()
	plugin.content = content
	plugin.mu.Unlock()

	return plugin.Load()
}

func (plugin *Plugin) vmFactory() (*runtime, error) {
	vm := otto.New()
	plugin.mu.RLock()
	err := vm.Set(VarContext, plugin.ctx)
	plugin.mu.RUnlock()
	if err != nil {
		plugin.log.Errorf("vm factory set context err: %s", err)
		return nil, err
	}

	err = vm.Set(FuncRequire, plugin.capal.Require)
	if err != nil {
		plugin.log.Errorf("vm factory set require err: %s", err)
		return nil, err
	}

	plugin.mu.RLock()
	content := plugin.content
	plugin.mu.RUnlock()
	// watch out the concurrency condition in script
	_, err = vm.Run(content)
	if err != nil {
		plugin.log.Errorf("vm factory run content err: %s", err)
		return nil, err
	}

	pr, err := vm.Call(FuncRegister, nil)
	if err != nil {
		plugin.log.Errorf("vm factory register err: %s", err)
		return nil, err
	}
	if !pr.IsObject() {
		plugin.log.Error(tigerbalm.ErrRegisterNotObject)
		return nil, tigerbalm.ErrRegisterNotObject
	}

	registration, err := getRegistration(pr.Object())
	if err != nil {
		plugin.log.Errorf("vm factory get registration err: %s", err)
		return nil, err
	}
	return &runtime{registration, vm}, nil
}

func (plugin *Plugin) httpHandlerFactory() interface{} {
	runtime, err := plugin.vmFactory()
	if err != nil {
		plugin.log.Errorf("plugin: %s, http handler factory get vm err: %s",
			plugin.name, err)
		return nil
	}
	return runtime.route.handler
}

func (plugin *Plugin) kafkaHandlerFactory() interface{} {
	runtime, err := plugin.vmFactory()
	if err != nil {
		plugin.log.Errorf("plugin: %s, kafka handler factory get vm err: %s",
			plugin.name, err)
		return nil
	}
	return runtime.consume.handler
}

func (plugin *Plugin) Http() bool {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.http
}

func (plugin *Plugin) HttpMethod() string {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.httpMethod
}

func (plugin *Plugin) HttpPath() string {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.httpPath
}

func (plugin *Plugin) Kafka() bool {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.kafka
}

func (plugin *Plugin) KafkaTopic() string {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.kafkaTopic
}

func (plugin *Plugin) KafkaGroup() string {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.kafkaGroup
}

func (plugin *Plugin) Log() *tblog.TbLog {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.log
}

func (plugin *Plugin) HttpHandle(req *tbhttp.Request) (*tbhttp.Response, error) {
	handler := plugin.httpPool.Get()
	defer plugin.httpPool.Put(handler)

	if handler == nil {
		plugin.log.Error("plugin get nil handler from pool")
		return nil, tigerbalm.ErrNewInterpreter
	}
	this, err := otto.ToValue(nil)
	if err != nil {
		plugin.log.Errorf("plugin to value err: %s", err)
		return nil, err
	}
	ottoRsp, err := handler.(otto.Value).Call(this, req)
	if err != nil {
		plugin.log.Errorf("plugin call err: %s", err)
		return nil, err
	}

	return tbhttp.OttoValue2TbRsp(ottoRsp)
}

func (plugin *Plugin) KafkaHandle(msg *tbkafka.CGMessage) {
	handler := plugin.kafkaPool.Get()
	defer plugin.httpPool.Put(handler)
	if handler == nil {
		plugin.log.Error("plugin get nil handler from pool")
		return
	}
	this, err := otto.ToValue(nil)
	if err != nil {
		plugin.log.Errorf("plugin to value err: %s", err)
		return
	}
	_, err = handler.(otto.Value).Call(this, msg)
	if err != nil {
		plugin.log.Errorf("plugin call err: %s", err)
		return
	}
}

func (plugin *Plugin) Fini() {
	plugin.rotateLog.Close()
}
