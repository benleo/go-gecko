package drivers

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func StrategyDriverFactory() (string, gecko.BundleFactory) {
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
	Payload    gecko.JSONPacket
}

// 驱动触发策略
type DriveStrategy func(event map[string]interface{}) *ConnectedDevice

// 策略驱动Driver
type StrategyDriver struct {
	*gecko.AbcDriver
	initArgs   *cfg.Config
	strategies []DriveStrategy
}

// 添加驱动触发策略
func (d *StrategyDriver) AddDriveStrategy(strategy DriveStrategy) {
	d.strategies = append(d.strategies, strategy)
}

func (d *StrategyDriver) GetInitArgs() *cfg.Config {
	return d.initArgs
}

func (d *StrategyDriver) OnInit(args *cfg.Config, ctx gecko.Context) {
	d.initArgs = args

	strategies := args.MustConfig("strategies")
	zlog := gecko.ZapSugarLogger

	strategies.ForEach(func(name string, value interface{}) {
		rule := cfg.Wrap(value.(map[string]interface{}))
		matchFields, err := rule.GetStringMapOrDefault("matchEventFields", make(map[string]string, 0))
		if err != nil {
			zlog.Panicf("matchEventFields字段格式错误[TABLE]", err)
		}
		targetUUID := rule.MustString("targetUUID")
		targetCommand := rule.MustConfig("targetCommand")

		if 0 == len(matchFields) || "" == targetUUID || targetCommand.IsEmpty() {
			zlog.Panicf("未正确配置匹配规则： matchFields= %s, targetUUID= %s, targetCommand=%s",
				matchFields, targetUUID, targetCommand.RefMap())
		} else {
			zlog.Debugf("联动配置规则： matchFields= %s, targetUUID= %s, targetCommand=%s",
				matchFields, targetUUID, targetCommand.RefMap())
		}

		d.AddDriveStrategy(func(event map[string]interface{}) *ConnectedDevice {
			// 检查是否匹配字段
			matches := true
			for k, v := range matchFields {
				if cfg.Value2String(v) == event[k] {
					matches = false
					break
				}
			}
			if !matches {
				return nil
			}

			return &ConnectedDevice{
				DeviceUUID: targetUUID,
				Payload:    targetCommand.RefMap(),
			}
		})
	})

}

func (d *StrategyDriver) Handle(session gecko.EventSession, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	responses := make(map[string]gecko.JSONPacket, 0)
	message := session.Inbound()
	for _, strategy := range d.strategies {
		target := strategy(message.Data)
		if nil == target {
			continue
		}
		uuid := target.DeviceUUID
		if ret, err := deliverer.Execute(uuid, target.Payload); nil != err {
			responses[uuid] = gecko.JSONPacket{
				"status":  "error",
				"message": err.Error(),
			}
		} else {
			responses[uuid] = gecko.JSONPacket{
				"status": "success",
				"data":   ret,
			}
		}
	}
	session.Outbound().AddDataField("driverResponse", responses)
	return nil
}
