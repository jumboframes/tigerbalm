/*
 * Apache License 2.0
 *
 * Copyright (c) 2022, Austin Zhai (singchia1202@gmail.com)
 * All rights reserved.
 */

package tigerbalm

import (
	"errors"
)

var (
	ErrRegisterNotFunction = errors.New("register not function")
	ErrRegisterNotObject   = errors.New("register not object")
	ErrNewInterpreter      = errors.New("new interpreter error")
	ErrNoSuchSlot          = errors.New("no such slot")
)
