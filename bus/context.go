package bus

import "github.com/kataras/iris/v12"

type Context struct {
	iris.Context
	RelativePath string
}

type ContextHandler func(ctx *Context)
