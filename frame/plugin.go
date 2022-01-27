package frame

import (
	"path/filepath"
	"sync"

	"github.com/jumboframes/tigerbalm"
	"github.com/jumboframes/tigerbalm/frame/capal"
	"github.com/jumboframes/tigerbalm/frame/capal/tbhttp"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"

	"github.com/robertkrimen/otto"
)

type Plugin struct {
	mu sync.RWMutex
	// metas
	path    string
	method  string
	name    string
	content []byte

	// conf
	config *tigerbalm.Config

	// capals
	capal *capal.Capal
	ctx   *capal.PluginContext

	// runtimes
	pool sync.Pool

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
		tblog.Errorf("newplugin | get vm err: %s", err)
		return err
	}

	pool := sync.Pool{
		New: plugin.handlerFactory,
	}
	pool.Put(runtime.handler)

	plugin.mu.Lock()
	plugin.path = runtime.path
	plugin.method = runtime.method
	plugin.pool = pool
	plugin.mu.Unlock()
	return nil
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

	route, err := getRoute(pr.Object())
	if err != nil {
		plugin.log.Errorf("vm factory get route err: %s", err)
		return nil, err
	}
	return &runtime{route, vm}, nil
}

func (plugin *Plugin) handlerFactory() interface{} {
	runtime, err := plugin.vmFactory()
	if err != nil {
		plugin.log.Errorf("handler factory get vm err: %s", err)
		return nil
	}
	return runtime.handler
}

func (plugin *Plugin) Path() string {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.path
}

func (plugin *Plugin) Method() string {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.method
}

func (plugin *Plugin) Log() *tblog.TbLog {
	plugin.mu.RLock()
	defer plugin.mu.RUnlock()
	return plugin.log
}

func (plugin *Plugin) Handle(req *tbhttp.Request) (*tbhttp.Response, error) {
	// get runtime
	handler := plugin.pool.Get()
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
	plugin.pool.Put(handler)

	return tbhttp.OttoValue2TbRsp(ottoRsp)
}

func (plugin *Plugin) Fini() {
	plugin.rotateLog.Close()
}
