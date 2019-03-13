package nop

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NewNopShadowDevice() gecko.ShadowDevice {
	return &NopShadowDevice{
		AbcShadowDevice: gecko.NewAbcShadowDevice(),
	}
}

func NopShadowDeviceFactory() (string, gecko.BundleFactory) {
	return "NopShadowDevice", func() interface{} {
		return NewNopShadowDevice()
	}
}

type NopShadowDevice struct {
	*gecko.AbcShadowDevice
}

func (s *NopShadowDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	// nop
}

// 检查是否符合影子设备的数据
func (s *NopShadowDevice) IsShadow(json gecko.JSONPacket) bool {
	return true
}

// 转换输入的数据
func (s *NopShadowDevice) TransformInput(topic string, json gecko.JSONPacket) (newTopic string, newJson gecko.JSONPacket) {
	return topic + "#NOP", json
}

// 转换返回给设备的数据
func (s *NopShadowDevice) TransformOutput(json gecko.JSONPacket) (newJson gecko.JSONPacket) {
	return json
}
