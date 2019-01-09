package bundles

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tidwall/evio"
	"github.com/yoojia/go-gecko"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 使用Evio框架的事件式网络服务器触发器
type NetworkServerTrigger struct {
	gecko.Trigger
	// 数据序列化与反序列化
	decoder gecko.Decoder
	encoder gecko.Encoder
	topic   string
	//
	events    evio.Events
	addresses []string
}

func (nst *NetworkServerTrigger) OnInit(args map[string]interface{}, scoped gecko.GeckoScoped) {
	nst.decoder = gecko.JSONDefaultDecoder
	nst.encoder = gecko.JSONDefaultEncoder
	config := conf.MapToMap(args)
	//
	if addrs, err := config.MustStringArray("bindAddresses"); nil != err {
		nst.withTag(log.Panic).Err(err).Msg("配置字段[bindAddresses]必须是个字符串数组")
	} else {
		nst.addresses = addrs
	}
	nst.withTag(log.Info).Msg("Network服务器 Trigger初始化")
}

func (nst *NetworkServerTrigger) ONStart(scoped gecko.GeckoScoped, invoker gecko.Invoker) {
	nst.withTag(log.Info).Msgf("Network服务器启动，绑定地址： %s", nst.addresses)
	go func() {
		if err := evio.Serve(nst.events, nst.addresses...); nil != err {
			nst.withTag(log.Error).Err(err).Msg("Network服务器停止")
		}
		// TODO
	}()
}

func (nst *NetworkServerTrigger) OnStop(scoped gecko.GeckoScoped, invoker gecko.Invoker) {
	//nst.withTag(log.Info).Msgf("UDP服务器关闭，绑定地址： %s", nst.udpAddr.String())
	//if nil != nst.udpConn {
	//	nst.udpConn.Close()
	//}
	//nst.shutdownCompleted <- struct{}{}
}

func (nst *NetworkServerTrigger) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return fun().Str("tag", "NetworkServerTrigger")
}
