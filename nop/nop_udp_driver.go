package nop

import (
	"errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func NopUdpDriverFactory() (string, gecko.BundleFactory) {
	return "NopUdpDriver", func() interface{} {
		return &NopUdpDriver{
			AbcDriver: gecko.NewAbcDriver(),
		}
	}
}

// 触发UDP设备的模拟Driver
type NopUdpDriver struct {
	*gecko.AbcDriver
}

func (du *NopUdpDriver) OnInit(args map[string]interface{}, ctx gecko.Context) {
	du.withTag(log.Debug).Msg("初始化...")
}

func (du *NopUdpDriver) OnStart(ctx gecko.Context) {
	du.withTag(log.Debug).Msg("启动...")
}

func (du *NopUdpDriver) OnStop(ctx gecko.Context) {
	du.withTag(log.Debug).Msg("停止...")
}

func (du *NopUdpDriver) Handle(session gecko.Session, selector gecko.ProtoPipelineSelector, ctx gecko.Context) error {
	if pl, ok := selector("udp"); ok {
		for _, dev := range pl.FindHardwareByGroupAddress("127.0.0.1") {
			if iDev, ok := dev.(gecko.InteractiveDevice); ok {
				if frame, err := iDev.Process(session.NewPacketFrame([]byte("HAHAHAHA")), ctx); nil != err {
					du.withTag(log.Error).Err(err).Msg("驱动设备失败")
				} else {
					du.withTag(log.Debug).Bytes("resp", frame.Data())
				}
			}

		}
		session.Outbound().AddDataField("status", "SUCCESS")
		return nil
	} else {
		return errors.New("未找到支持UDP的Pipeline")
	}
}

func (du *NopUdpDriver) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return fun().Str("tag", "NopUdpDriver")
}
