package drivers

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
	"github.com/yoojia/go-gecko/utils"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func StrategyDriverFactory() (string, gecko.Factory) {
	return "StrategyDriver", func() interface{} {
		return NewStrategyDriver()
	}
}

func NewStrategyDriver() *StrategyDriver {
	return &StrategyDriver{
		AbcDriver:  gecko.NewAbcDriver(),
		strategies: make([]DriveStrategy, 0),
	}
}

// 联动目标设备
type ConnectedDevice struct {
	DeviceUUID string
	Payload    *gecko.MessagePacket
}

// 驱动触发策略
type DriveStrategy func(event *gecko.MessagePacket) *ConnectedDevice

func (ds DriveStrategy) Do(event *gecko.MessagePacket) *ConnectedDevice {
	return ds(event)
}

// 策略驱动Driver
type StrategyDriver struct {
	*gecko.AbcDriver
	initArgs   map[string]interface{}
	strategies []DriveStrategy
}

// 添加驱动触发策略
func (d *StrategyDriver) AddDriveStrategy(strategy DriveStrategy) {
	d.strategies = append(d.strategies, strategy)
}

func (d *StrategyDriver) GetInitArgs() map[string]interface{} {
	return d.initArgs
}

func (d *StrategyDriver) OnInit(args map[string]interface{}, ctx gecko.Context) {
	d.initArgs = args

	zlog := gecko.ZapSugarLogger
	for _, value := range utils.ToMap(args["strategies"]) {
		strategy := cfg.Wrap(value.(map[string]interface{}))
		matchFields, err := strategy.GetStringMapOrDefault("matchFields", make(map[string]string, 0))
		if err != nil {
			zlog.Panicf("配置字段[matchFields]格式错误[TABLE]", err)
		}
		uuid := strategy.MustString("uuid")
		command := strategy.MustConfig("command")

		if 0 == len(matchFields) || "" == uuid || command.IsEmpty() {
			zlog.Panicf("未正确配置匹配规则： matchFields= %s, uuid= %s, command=%s",
				matchFields, uuid, command.RefMap())
		} else {
			zlog.Debugf("联动配置规则： matchFields= %s, uuid= %s, command=%s",
				matchFields, uuid, command.RefMap())
		}

		d.AddDriveStrategy(func(event *gecko.MessagePacket) *ConnectedDevice {
			// 检查是否匹配字段
			matches := true
			for key, excepted := range matchFields {
				value, ok := event.GetFieldString(key)
				if ok && excepted != value {
					matches = false
					break
				}
			}
			if !matches {
				return nil
			}

			return &ConnectedDevice{
				DeviceUUID: uuid,
				Payload:    gecko.NewMessagePacketFields(command.RefMap()),
			}
		})
	}

}

func (d *StrategyDriver) Handle(session gecko.EventSession, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	responses := make(map[string]*gecko.MessagePacket, 0)
	inbound := session.Inbound()
	for _, strategy := range d.strategies {
		target := strategy.Do(inbound)
		if nil == target {
			continue
		}
		uuid := target.DeviceUUID
		if ret, err := deliverer.Deliver(uuid, target.Payload); nil != err {
			responses[uuid] = gecko.NewMessagePacketFields(map[string]interface{}{
				"status":  "error",
				"message": err.Error(),
			})
		} else {
			responses[uuid] = gecko.NewMessagePacketFields(map[string]interface{}{
				"status": "success",
				"data":   ret,
			})
		}
	}
	session.Outbound().AddField("driverResponse", responses)
	return nil
}
