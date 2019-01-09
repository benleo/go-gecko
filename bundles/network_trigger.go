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
	// Events
	nst.events.Data = func(conn evio.Conn, in []byte) (out []byte, action evio.Action) {
		// 使用Invoker调度内部系统处理，完成后返回给客户端
		if json, deErr := nst.decoder(in); nil == deErr {
			income := gecko.NewTriggerEvent(nst.topic, json)
			// 处理并等待结果
			resp := make(chan map[string]interface{})
			invoker(income, func(data map[string]interface{}) {
				resp <- data
			})
			// Decode and send back client
			if bytes, enErr := nst.encoder(<-resp); nil == deErr {
				out = bytes
			} else {
				nst.withTag(log.Error).Err(enErr).Msg("服务器无法序列化的数据")
			}
		} else {
			nst.withTag(log.Error).Err(deErr).Msg("服务器接收到无法解析的数据：" + conn.RemoteAddr().String())
		}
		return
	}
	go func() {
		// 绑定服务
		if err := evio.Serve(nst.events, nst.addresses...); nil != err {
			nst.withTag(log.Error).Err(err).Msg("Network服务器停止")
		}
		// FIXME 如何停止evio服务？
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
