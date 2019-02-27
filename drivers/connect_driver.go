package node

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
	"parkingwang.com/irain-edge/hardwares/dongkong"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func ConnectDriverFactory() (string, gecko.BundleFactory) {
	return "ConnectDriver", func() interface{} {
		return &ConnectDriver{
			AbcDriver: gecko.NewAbcDriver(),
		}
	}
}

// 设备直接连接联动Driver
type ConnectDriver struct {
	*gecko.AbcDriver
	// 开关组地址信息
	targetGroupAddress   string
	targetPrivateAddress string
	targetSubAddress     string
	// 事件源信息
	sourceGroupAddress   string
	sourcePrivateAddress string
	sourceSubAddress     string
	// 事件源读取字段的Key
	sourceGroupAddressKey   string
	sourcePrivateAddressKey string
	sourceSubAddressKey     string
}

func (sd *ConnectDriver) OnInit(config *cfg.Config, ctx gecko.Context) {
	gecko.ZapDebug("初始化...")

	zap := gecko.Zap()
	defer zap.Sync()

	// 目标设备的地址信息
	sd.targetGroupAddress = config.MustString("targetGroupAddress")
	sd.targetPrivateAddress = config.MustString("targetPrivateAddress")
	sd.targetSubAddress = config.MustString("targetSubAddress")
	if sd.targetGroupAddress == "" || sd.targetPrivateAddress == "" || sd.targetSubAddress == "" {
		zap.Panicw("未配置联动目标设备的地址信息")
	}
	zap.Debugw("联动目标设备", "groupAddress", sd.targetGroupAddress,
		"privateAddress", sd.targetPrivateAddress,
		"subAddress", sd.targetSubAddress,
	)

	// 触发源设备的地址信息
	sd.sourceGroupAddress = config.MustString("sourceGroupAddress")
	sd.sourcePrivateAddress = config.MustString("sourcePrivateAddress")
	sd.sourceSubAddress = config.MustString("sourceSubAddress")
	if sd.sourceGroupAddress == "" || sd.sourcePrivateAddress == "" || sd.sourceSubAddress == "" {
		zap.Panicw("未配置联动目标设备的地址信息")
	}
	zap.Debugw("联动触发源设备", "groupAddress", sd.sourceGroupAddress,
		"privateAddress", sd.sourcePrivateAddress,
		"subAddress", sd.sourceSubAddress,
	)

	// 读取事件源数据字段的Key
	sd.sourceGroupAddressKey = config.GetStringOrDefault("sourceGroupAddressKey", dongk.DK_KEY_DEV_SN)
	sd.sourcePrivateAddressKey = config.GetStringOrDefault("sourcePrivateAddressKey", dongk.DK_KEY_OP_DOOR_ID)
	sd.sourceSubAddressKey = config.GetStringOrDefault("sourceSubAddressKey", dongk.DK_KEY_OP_IO_DIR)

}

func (sd *ConnectDriver) OnStart(ctx gecko.Context) {
	gecko.ZapDebug("启动...")
}

func (sd *ConnectDriver) OnStop(ctx gecko.Context) {
	gecko.ZapDebug("停止...")
}

func (sd *ConnectDriver) Handle(session gecko.Session, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	data := cfg.WrapConfig(session.Inbound().Data)
	// 读取当前事件的数据，判断是否满足触发联动的条件
	sourceGroupAddr := data.MustString(sd.sourceGroupAddressKey)
	sourcePrivateAddr := data.MustString(sd.sourcePrivateAddressKey)
	sourceSubAddr := data.MustString(sd.sourceSubAddressKey)
	if sd.sourceGroupAddress == sourceGroupAddr &&
		sd.sourcePrivateAddress == sourcePrivateAddr &&
		sd.sourceSubAddress == sourceSubAddr {
		// 满足条件即触发目标设备
		zap := gecko.Zap()
		defer zap.Sync()

		zap.Debugw("联动设备", "Group", sd.targetGroupAddress, "Private", sd.targetPrivateAddress)

		deliverer.Execute(gecko.MakeUnionAddress(sd.targetGroupAddress, sd.targetPrivateAddress), gecko.PacketMap{
			"addr":  sd.targetSubAddress,
			"state": "triggered",
		})
	}
	return nil
}
