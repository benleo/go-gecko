package nop

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NopDriverFactory() (string, gecko.BundleFactory) {
	return "NopDriver", func() interface{} {
		return &NopDriver{
			AbcDriver: gecko.NewAbcDriver(),
		}
	}
}

// 触发UDP设备的模拟Driver
type NopDriver struct {
	*gecko.AbcDriver
}

func (du *NopDriver) OnInit(args map[string]interface{}, ctx gecko.Context) {
	du.withTag(log.Debug).Msg("初始化...")
}

func (du *NopDriver) OnStart(ctx gecko.Context) {
	du.withTag(log.Debug).Msg("启动...")
}

func (du *NopDriver) OnStop(ctx gecko.Context) {
	du.withTag(log.Debug).Msg("停止...")
}

func (du *NopDriver) Handle(session gecko.Session, selector gecko.ProtoPipelineSelector, ctx gecko.Context) error {
	//if pipeline, ok := selector("udp"); !ok {
	//	return errors.New("无法查找到udp协议对应的Pipeline")
	//} else {
	//	// 通过Pipeline，向特定设备发送消息：
	//	//groupAddress := "目标设备的GroupAddress"
	//	//privateAddress := "目标设备的PrivateAddress"
	//	//resp, err := pipeline.ExecuteDevice(groupAddress, privateAddress, gecko.NewPackFrame([]byte{0}))
	//
	//}

	return nil
}

func (du *NopDriver) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return fun().Str("tag", "NopDriver")
}
