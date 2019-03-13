package nop

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NewNopLogicDevice() gecko.LogicDevice {
	return &NopLogicDevice{
		AbcLogicDevice: gecko.NewAbcLogicDevice(),
	}
}

func NopLogicDeviceFactory() (string, gecko.BundleFactory) {
	return "NopLogicDevice", func() interface{} {
		return NewNopLogicDevice()
	}
}

type NopLogicDevice struct {
	*gecko.AbcLogicDevice
}

func (s *NopLogicDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	// nop
}

// 检查是否符合逻辑设备的数据
func (s *NopLogicDevice) IsLogicMatch(json gecko.JSONPacket) bool {
	return true
}

// 转换输入的数据
func (s *NopLogicDevice) TransformInput(topic string, json gecko.JSONPacket) (newTopic string, newJson gecko.JSONPacket) {
	return topic + "#NOP", json
}

// 转换返回给设备的数据
func (s *NopLogicDevice) TransformOutput(json gecko.JSONPacket) (newJson gecko.JSONPacket) {
	return json
}
