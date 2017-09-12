package skeleton

import (
	"reflect"
)

var (
	RouteHandler = NewServiceRegister()
)

type ServiceRegister struct {
	routes map[string]*serviceInfo
}

func NewServiceRegister() *ServiceRegister {
	return &ServiceRegister{
		routes: make(map[string]*serviceInfo),
	}
}

type serviceInfo struct {
	pattern      string
	invokeTarget interface{}
	invokeMethod reflect.Value
}

type BaseRouter interface {
	InitContext(ctx *Context)
}

func Router(pattern string, si interface{}, method string) {
	reflectVal := reflect.ValueOf(si)
	st := reflect.Indirect(reflectVal).Type()
	methodVal := reflectVal.MethodByName(method)
	if !methodVal.IsValid() {
		panic("'" + method + "' method doesn't exist in the logic " + st.Name())
	}

	route := &serviceInfo{}
	route.pattern = pattern
	route.invokeTarget = si
	route.invokeMethod = methodVal

	RouteHandler.routes[pattern] = route
}
