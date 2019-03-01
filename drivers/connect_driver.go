package drivers

import (
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NewConnectDriver(handler TriggerHandler) *ConnectDriver {
	return &ConnectDriver{
		AbcDriver:      gecko.NewAbcDriver(),
		triggerHandler: handler,
	}
}

// 联动目标设备
type ConnectedDevice struct {
	DeviceUUID string
	Payload    gecko.JSONPacket
}

// 触发处理函数
type TriggerHandler func(event map[string]interface{}) []ConnectedDevice

// 设备直接连接联动Driver
type ConnectDriver struct {
	*gecko.AbcDriver
	triggerHandler TriggerHandler
}

func (d *ConnectDriver) Handle(session gecko.EventSession, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	targets := d.triggerHandler(session.Inbound().Data)
	zap := gecko.Zap()
	defer zap.Sync()
	responses := make(map[string]gecko.JSONPacket, 0)
	for _, target := range targets {
		uuid := target.DeviceUUID
		if ret, err := deliverer.Execute(uuid, target.Payload); nil != err {
			zap.Errorf("目标设备联动操作发生错误：", err)
			responses[uuid] = gecko.JSONPacket{
				"error": err.Error(),
			}
		} else {
			zap.Debugf("目标设备联动操作返回响应：", ret)
			responses[uuid] = ret
		}
	}
	session.Outbound().AddDataField("driverResponse", responses)
	return nil
}