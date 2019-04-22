package nop

import (
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type NopInterceptor struct {
	*gecko.AbcInterceptor
	gecko.Initial
}

func (ni *NopInterceptor) OnInit(config map[string]interface{}, ctx gecko.Context) {

}

func (ni *NopInterceptor) Handle(attrs gecko.Attributes, topic string, uuid string, in *gecko.MessagePacket, ctx gecko.Context) error {
	//return ni.Drop()
	return ni.Next()
}

func NewNopInterceptor() gecko.Interceptor {
	return &NopInterceptor{
		AbcInterceptor: gecko.NewAbcInterceptor(),
	}
}

func NopInterceptorFactor() (string, gecko.Factory) {
	return "NopInterceptor", func() interface{} {
		return NewNopInterceptor()
	}
}
