package bundles

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func UdpDemoDriverFactory() (string, gecko.BundleFactory) {
	return "UdpDemoDriver", func() interface{} {
		return &UdpDemoDriver{
			AbcDriver: new(gecko.AbcDriver),
		}
	}
}

// 触发UDP设备的模拟Driver
type UdpDemoDriver struct {
	*gecko.AbcDriver
}

func (du *UdpDemoDriver) OnInit(args map[string]interface{}, ctx gecko.Context) {
	du.withTag(log.Debug).Msg("初始化...")
}

func (du *UdpDemoDriver) OnStart(ctx gecko.Context) {
	du.withTag(log.Debug).Msg("启动...")
}

func (du *UdpDemoDriver) OnStop(ctx gecko.Context) {
	du.withTag(log.Debug).Msg("停止...")
}

func (du *UdpDemoDriver) Handle(session gecko.Session, selector gecko.ProtoPipelineSelector, ctx gecko.Context) error {
	if pl, ok := selector("udp"); ok {
		for _, dev := range pl.FindDevicesByGroup("127.0.0.1") {
			if frame, err := dev.Process(session.NewPacketFrame([]byte("HAHAHAHA")), ctx); nil != err {
				du.withTag(log.Error).Err(err).Msg("驱动设备失败")
			} else {
				du.withTag(log.Debug).Bytes("resp", frame.Data())
			}
		}
		session.Outbound().AddDataField("status", "SUCCESS")
		return nil
	} else {
		return errors.New("未找到支持UDP的Pipeline")
	}
}

func (du *UdpDemoDriver) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return fun().Str("tag", "UdpDemoDriver")
}
