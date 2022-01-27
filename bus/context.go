package bus

import "github.com/kataras/iris/v12"

type ContextHttp struct {
	iris.Context
	RelativePath string
}
