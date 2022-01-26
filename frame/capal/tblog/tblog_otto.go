package tblog

import (
	"github.com/robertkrimen/otto"
)

type TbLogOtto struct {
	tblog *TbLog
}

func NewTbLogOtto(tblog *TbLog) *TbLogOtto {
	return &TbLogOtto{tblog}
}

func (log *TbLogOtto) Trace(call otto.FunctionCall) otto.Value {
	vs := make([]interface{}, len(call.ArgumentList))
	for index, arg := range call.ArgumentList {
		vs[index] = arg
	}
	log.tblog.Trace(vs...)
	return otto.NullValue()
}

func (log *TbLogOtto) Tracef(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) < 1 {
		return otto.NullValue()
	}
	format, err := call.ArgumentList[0].ToString()
	if err != nil {
		return otto.NullValue()
	}
	vs := make([]interface{}, len(call.ArgumentList)-1)
	for index, arg := range call.ArgumentList[1:] {
		vs[index] = arg
	}
	log.tblog.Tracef(format, vs...)
	return otto.NullValue()
}

func (log *TbLogOtto) Debug(call otto.FunctionCall) otto.Value {
	vs := make([]interface{}, len(call.ArgumentList))
	for index, arg := range call.ArgumentList {
		vs[index] = arg
	}
	log.tblog.Debug(vs...)
	return otto.NullValue()
}

func (log *TbLogOtto) Debugf(call otto.FunctionCall) otto.Value {
	if len(call.ArgumentList) < 1 {
		return otto.NullValue()
	}
	format, err := call.ArgumentList[0].ToString()
	if err != nil {
		return otto.NullValue()
	}
	vs := make([]interface{}, len(call.ArgumentList)-1)
	for index, arg := range call.ArgumentList[1:] {
		vs[index] = arg
	}
	log.tblog.Debugf(format, vs...)
	return otto.NullValue()
}

func (log *TbLogOtto) Info(call otto.FunctionCall) otto.Value {
	vs := make([]interface{}, len(call.ArgumentList))
	for index, arg := range call.ArgumentList {
		vs[index] = arg
	}
	log.tblog.Info(vs...)
	return otto.NullValue()
}

func (log *TbLogOtto) Warn(call otto.FunctionCall) otto.Value {
	vs := make([]interface{}, len(call.ArgumentList))
	for index, arg := range call.ArgumentList {
		vs[index] = arg
	}
	log.tblog.Warn(vs...)
	return otto.NullValue()
}

func (log *TbLogOtto) Error(call otto.FunctionCall) otto.Value {
	vs := make([]interface{}, len(call.ArgumentList))
	for index, arg := range call.ArgumentList {
		vs[index] = arg
	}
	log.tblog.Error(vs...)
	return otto.NullValue()
}

func (log *TbLogOtto) Fatal(call otto.FunctionCall) otto.Value {
	vs := make([]interface{}, len(call.ArgumentList))
	for index, arg := range call.ArgumentList {
		vs[index] = arg
	}
	log.tblog.Fatal(vs...)
	return otto.NullValue()
}
