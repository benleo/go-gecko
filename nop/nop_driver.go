package nop

import (
	"github.com/yoojia/go-gecko/v2"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NopDriverFactory() (string, gecko.Factory) {
	return "NopDriver", func() interface{} {
		return &NopDriver{
			AbcDriver: gecko.NewAbcDriver(),
		}
	}
}

// 触发UDP设备的模拟Driver
type NopDriver struct {
	*gecko.AbcDriver
	gecko.Initial
	gecko.LifeCycle
}

func (du *NopDriver) OnInit(config map[string]interface{}, ctx gecko.Context) {
	log.Debug("初始化...")
}

func (du *NopDriver) OnStart(ctx gecko.Context) {
	log.Debug("启动...")
}

func (du *NopDriver) OnStop(ctx gecko.Context) {
	log.Debug("停止...")
}

func (du *NopDriver) Drive(attrs gecko.Attributes, topic string, uuid string, in *gecko.MessagePacket, deliverer gecko.OutputDeliverer, ctx gecko.Context) (out *gecko.MessagePacket, err error) {
	return gecko.NewMessagePacketFields(map[string]interface{}{
		"status": "success",
	}), nil
}
