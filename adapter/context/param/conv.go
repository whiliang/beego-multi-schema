package param

import (
	"reflect"

	beecontext "github.com/whiliang/beego-multi-schema/adapter/context"
	"github.com/whiliang/beego-multi-schema/server/web/context"
	"github.com/whiliang/beego-multi-schema/server/web/context/param"
)

// ConvertParams converts http method params to values that will be passed to the method controller as arguments
func ConvertParams(methodParams []*MethodParam, methodType reflect.Type, ctx *beecontext.Context) (result []reflect.Value) {
	nps := make([]*param.MethodParam, 0, len(methodParams))
	for _, mp := range methodParams {
		nps = append(nps, (*param.MethodParam)(mp))
	}
	return param.ConvertParams(nps, methodType, (*context.Context)(ctx))
}
