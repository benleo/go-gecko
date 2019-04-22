package nop

import (
	"github.com/yoojia/go-gecko/v2"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NewNopPlugin() gecko.Plugin {
	return new(NopPlugin)
}

func NopPluginFactory() (string, gecko.Factory) {
	return "NopPlugin", func() interface{} {
		return NewNopPlugin()
	}
}

type NopPlugin struct {
	gecko.Plugin
}

func (no *NopPlugin) OnStart(ctx gecko.Context) {

}

func (no *NopPlugin) OnStop(ctx gecko.Context) {

}
