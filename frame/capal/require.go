package capal

import (
	"github.com/jumboframes/tigerbalm/frame/capal/tbhttp"
	"github.com/jumboframes/tigerbalm/frame/capal/tblog"
	"github.com/robertkrimen/otto"
)

const (
	ModuleHttp = "http"
	ModuleLog  = "log"
)

type Capal struct {
	httpFactory func(ctx *PluginContext) *tbhttp.TbHttp
	logFactory  func(ctx *PluginContext) *tblog.TbLog
}

func NewCapal(httpFactory func(ctx *PluginContext) *tbhttp.TbHttp,
	logFactory func(ctx *PluginContext) *tblog.TbLog) *Capal {
	return &Capal{
		httpFactory: httpFactory,
		logFactory:  logFactory,
	}
}

func (capal *Capal) Require(call otto.FunctionCall) otto.Value {
	ctx, err := getPluginContext(call)
	if err != nil {
		tblog.Errorf("require | get plugin context err: %s", err)
		return otto.NullValue()
	}

	log := capal.logFactory(ctx)
	if log == nil {
		tblog.Errorf("require | get nil from log factory")
	}
	argc := len(call.ArgumentList)
	if argc != 1 {
		log.Errorf("require args != 1, callee: %s, line: %d",
			call.Otto.Context().Callee, call.Otto.Context().Line)
		return otto.NullValue()
	}
	module := call.ArgumentList[0].String()
	switch module {
	case ModuleHttp:
		http := capal.httpFactory(ctx)
		value, err := otto.New().ToValue(http)
		if err != nil {
			log.Errorf("require http err: %s, callee: %s, line: %d",
				err, call.Otto.Context().Callee, call.Otto.Context().Line)
			return otto.NullValue()
		}
		return value

	case ModuleLog:
		logOtto := tblog.NewTbLogOtto(log)
		value, err := otto.New().ToValue(logOtto)
		if err != nil {
			log.Errorf("require log err: %s, callee: %s, line: %d",
				err, call.Otto.Context().Callee, call.Otto.Context().Line)
			return otto.NullValue()
		}
		return value
	}
	log.Error("require unsupported module")
	return otto.NullValue()
}

func getPluginContext(call otto.FunctionCall) (*PluginContext, error) {
	context, err := call.Otto.Get("context")
	if err != nil {
		return nil, err
	}
	name, err := context.Object().Get("Name")
	if err != nil {
		return nil, err
	}
	return &PluginContext{name.String()}, nil
}
