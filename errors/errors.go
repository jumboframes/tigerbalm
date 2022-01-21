package errors

import "errors"

var (
	ErrIllegalRegister     = errors.New("illegal register")
	ErrIllegalFunction     = errors.New("illegal function")
	ErrRegisterNotFunction = errors.New("register not function")
	ErrNewInterpreter      = errors.New("new interpreter error")
)
