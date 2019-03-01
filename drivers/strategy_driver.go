package drivers

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

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
type DriveStrategy func(event map[string]interface{}) ConnectedDevice

// 策略驱动Driver
type StrategyDriver struct {
	*gecko.AbcDriver
	initArgs   *cfg.Config
	strategies []DriveStrategy
}

// 添加驱动触发策略
func (d *StrategyDriver) AddDriveStrategy(strategy DriveStrategy)  {
	d.strategies = append(d.strategies, strategy)
}

func (d *StrategyDriver) GetInitArgs() *cfg.Config {
	return d.initArgs
}

func (d *StrategyDriver) OnInit(args *cfg.Config, ctx gecko.Context) {
	d.initArgs = args
}

func (d *StrategyDriver) Handle(session gecko.EventSession, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	zap := gecko.Zap()
	defer zap.Sync()

	responses := make(map[string]gecko.JSONPacket, 0)
	message := session.Inbound()
	for _, strategy := range d.strategies {
		target := strategy(message.Data)
		if ret, err := deliverer.Execute(target.DeviceUUID, target.Payload); nil != err {
			responses[target.DeviceUUID] = gecko.JSONPacket{
				"status":  "error",
				"message": err.Error(),
			}
		} else {
			responses[target.DeviceUUID] = gecko.JSONPacket{
				"status": "success",
				"data":   ret,
			}
		}
	}
	session.Outbound().AddDataField("driverResponse", responses)
	return nil
}
