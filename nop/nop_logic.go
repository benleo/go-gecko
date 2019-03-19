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

func NopLogicDeviceFactory() (string, gecko.Factory) {
	return "NopLogicDevice", func() interface{} {
		return NewNopLogicDevice()
	}
}

type NopLogicDevice struct {
	*gecko.AbcLogicDevice
	gecko.Initial
}

func (s *NopLogicDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	// nop
}

// 检查是否符合逻辑设备的数据
func (s *NopLogicDevice) CheckIfMatch(json gecko.MessagePacket) bool {
	return true
}

// 转换返回给设备的数据
func (s *NopLogicDevice) Transform(json gecko.MessagePacket) (newJson gecko.MessagePacket) {
	return json
}
