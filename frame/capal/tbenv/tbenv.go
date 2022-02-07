package tbenv

import (
	"github.com/jumboframes/tigerbalm"
	"github.com/robertkrimen/otto"
)

type TbEnv struct{}

func (tbenv *TbEnv) Get(call otto.FunctionCall) otto.Value {
	argc := len(call.ArgumentList)
	if argc != 1 {
		return otto.NullValue()
	}
	name, err := call.ArgumentList[0].ToString()
	if err != nil {
		return otto.NullValue()
	}
	for _, env := range tigerbalm.Conf.Env {
		if name == env.Name {
			value, err := otto.ToValue(env.Value)
			if err != nil {
				return otto.NullValue()
			}
			return value
		}
	}
	return otto.NullValue()
}
