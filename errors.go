package tigerbalm

import (
	"errors"
)

var (
	ErrRegisterNotFunction = errors.New("register not function")
	ErrRegisterNotObject   = errors.New("register not object")
	ErrNewInterpreter      = errors.New("new interpreter error")
)
