package nop

import (
	"github.com/yoojia/go-gecko/v2"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NopTriggerFactory() (string, gecko.Factory) {
	return "NopTrigger", func() interface{} {
		return &NopTrigger{
			AbcTrigger: gecko.NewAbcTrigger(),
		}
	}
}

// 触发UDP设备的模拟Driver
type NopTrigger struct {
	*gecko.AbcTrigger
	gecko.Initial
	gecko.LifeCycle
}

func (du *NopTrigger) OnInit(config map[string]interface{}, ctx gecko.Context) {
	log.Debug("初始化...")
}

func (du *NopTrigger) OnStart(ctx gecko.Context) {
	log.Debug("启动...")
}

func (du *NopTrigger) OnStop(ctx gecko.Context) {
	log.Debug("停止...")
}

func (du *NopTrigger) Touch(attrs gecko.Attributes, topic string, uuid string, in *gecko.MessagePacket, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	return nil
}
