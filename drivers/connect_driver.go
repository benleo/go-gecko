package node

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
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

func NewConnectDriver() *ConnectDriver {
	return &ConnectDriver{
		AbcDriver: gecko.NewAbcDriver(),
	}
}

// 触发设备数据包生产接口
type TriggerPacketProducer func(session gecko.Session, trigger gecko.DeviceAddress) gecko.PacketMap

// 设备直接连接联动Driver
type ConnectDriver struct {
	*gecko.AbcDriver

	targetAddress         gecko.DeviceAddress // 目标设备地址
	triggerAddress        gecko.DeviceAddress // 触发源设备地址
	eventAddressKeys      addressKey          // 当前事件读取触发源地址的字段Key
	triggerPacketProducer TriggerPacketProducer
}

func (cd *ConnectDriver) OnInit(config *cfg.Config, ctx gecko.Context) {
	gecko.ZapDebug("初始化...")

	zap := gecko.Zap()
	defer zap.Sync()

	// 目标设备的地址信息
	cd.targetAddress = gecko.DeviceAddress{
		Group:    config.MustString("targetGroupAddress"),
		Private:  config.MustString("targetPrivateAddress"),
		Internal: config.MustString("targetInternalAddress"),
	}

	if !cd.targetAddress.IsValid() {
		zap.Panicw("未配置联动目标设备的地址信息")
	}
	zap.Debugf("联动目标设备: %s", cd.targetAddress.String())

	// 触发源设备的地址信息
	cd.triggerAddress = gecko.DeviceAddress{
		Group:    config.MustString("triggerGroupAddress"),
		Private:  config.MustString("triggerPrivateAddress"),
		Internal: config.MustString("triggerInternalAddress"),
	}
	if !cd.triggerAddress.IsValid() {
		zap.Panic("未配置联动触发源设备的地址")
	}
	zap.Debugf("联动触发源设备: %s", cd.triggerAddress.String())

	// 读取当前事件的地址数据字段的Key
	cd.eventAddressKeys = addressKey{
		GroupKey:    config.GetStringOrDefault("eventGroupAddressKey", "groupAddress"),
		PrivateKey:  config.GetStringOrDefault("eventPrivateAddressKey", "privateAddress"),
		InternalKey: config.GetStringOrDefault("eventInternalAddressKey", "internalAddress"),
	}

	// 默认事件生产接口
	cd.SetTriggerPacketProducer(func(session gecko.Session, trigger gecko.DeviceAddress) gecko.PacketMap {
		return gecko.PacketMap{
			"state": "triggered",
			"addr":  trigger.Internal,
		}
	})
}

func (cd *ConnectDriver) OnStart(ctx gecko.Context) {
	gecko.ZapDebug("启动...")
}

func (cd *ConnectDriver) OnStop(ctx gecko.Context) {
	gecko.ZapDebug("停止...")
}

func (cd *ConnectDriver) Handle(session gecko.Session, deliverer gecko.OutputDeliverer, ctx gecko.Context) error {
	data := cfg.WrapConfig(session.Inbound().Data)
	// 读取当前事件的数据，判断是否满足触发联动的条件
	sessionAddress := gecko.DeviceAddress{
		Group:    data.MustString(cd.eventAddressKeys.GroupKey),
		Private:  data.MustString(cd.eventAddressKeys.PrivateKey),
		Internal: data.MustString(cd.eventAddressKeys.InternalKey),
	}

	if cd.triggerAddress.Equals(sessionAddress) {
		// 满足条件即触发目标设备
		zap := gecko.Zap()
		defer zap.Sync()

		zap.Debugw("联动设备", cd.targetAddress.String())

		pack := cd.triggerPacketProducer(session, cd.triggerAddress)
		if ret, err := deliverer.Execute(cd.targetAddress.GetUnionAddress(), pack); nil != err {
			zap.Error("联动设备发生错误", err)
		} else {
			zap.Debug("联动设备返回结果", ret)
		}
	}
	return nil
}

// 设置触发目标数据包生产接口
func (cd *ConnectDriver) SetTriggerPacketProducer(producer TriggerPacketProducer) {
	cd.triggerPacketProducer = producer
}

////

type addressKey struct {
	GroupKey    string
	PrivateKey  string
	InternalKey string
}
