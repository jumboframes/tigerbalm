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

	// capals
	capal *capal.Capal
	ctx   *capal.PluginContext

	// runtimes
	pool sync.Pool

	// logs
	log       *tblog.TbLog
	rotateLog *rotatelogs.RotateLogs
}

func NewPlugin(name string, content []byte,
	cpl *capal.Capal, config *tigerbalm.Config) (*Plugin, error) {

	plugin := &Plugin{
		name:    name,
		content: content,
		capal:   cpl,
		ctx:     &capal.PluginContext{name},
	}

	// plugin log
	level, err := tblog.ParseLevel(config.Plugin.Log.Level)
	if err != nil {
		tblog.Errorf("newplugin | tblog parse level: %s err: %s",
			config.Plugin.Log.Level, err)
		return nil, err
	}
	logFile := filepath.Join(config.Plugin.Log.Path, name, name+ExtLog)
	rotateLog, err := rotatelogs.New(logFile,
		rotatelogs.WithRotationCount(config.Plugin.Log.MaxRolls),
		rotatelogs.WithRotationSize(config.Plugin.Log.MaxSize))
	if err != nil {
		tblog.Errorf("newplugin | rotate log: %s new err: %s",
			logFile, err)
		return nil, err
	}
	log := tblog.NewTbLog().WithLevel(level).WithOutput(rotateLog)

	plugin.log = log
	plugin.rotateLog = rotateLog
	return plugin, nil
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

func (plugin *Plugin) Reload(content []byte) error {
	plugin.mu.Lock()
	plugin.content = content
	plugin.mu.Unlock()

	return plugin.Load()
}

func (plugin *Plugin) vmFactory() (*runtime, error) {
	vm := otto.New()
	err := vm.Set(VarContext, plugin.ctx)
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
	_, err = vm.Run(plugin.content)
	plugin.mu.RUnlock()
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
	return plugin.path
}

func (plugin *Plugin) Method() string {
	return plugin.method
}

func (plugin *Plugin) Log() *tblog.TbLog {
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
