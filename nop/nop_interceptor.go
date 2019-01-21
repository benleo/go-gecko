package nop

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type NopInterceptor struct {
	*gecko.AbcInterceptor
}

func (ni *NopInterceptor) OnInit(config *cfg.Config, ctx gecko.Context) {

}

func (ni *NopInterceptor) Handle(session gecko.Session, ctx gecko.Context) error {
	return nil
}

func NewNopInterceptor() gecko.Interceptor {
	return &NopInterceptor{
		AbcInterceptor: gecko.NewAbcInterceptor(),
	}
}

func NopInterceptorFactor() (string, gecko.BundleFactory) {
	return "NopInterceptor", func() interface{} {
		return NewNopInterceptor()
	}
}
