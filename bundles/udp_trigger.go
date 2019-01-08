package bundles

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"net"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// UDP服务端Trigger
type UdpServerTrigger struct {
	gecko.Trigger
	udpAddr *net.UDPAddr
	udpConn net.Conn
	//
	sendBuff chan *UdpPack
	recvBuff chan *UdpPack
}

func (ust *UdpServerTrigger) OnInit(args map[string]interface{}, scoped gecko.GeckoScoped) {
	config := conf.MapToMap(args)
	address := config.MustString("listenAddress")
	if addr, err := net.ResolveUDPAddr("udp", address); nil != err {
		ust.withTag(log.Panic).Err(err).Msgf("UDP服务端绑定地址错误: %s", address)
	} else {
		ust.udpAddr = addr
	}
	ust.withTag(log.Info).Msg("UDP服务器Trigger初始化")
}

func (ust *UdpServerTrigger) OnStart(scoped gecko.GeckoScoped, invoker gecko.TriggerInvoker) {
	ust.withTag(log.Info).Msgf("UDP服务器启动，绑定地址： %s", ust.udpAddr.String())
	if conn, err := net.ListenUDP("udp", ust.udpAddr); nil != err {
		ust.withTag(log.Panic).Err(err).Msg("UDP服务器启动绑定失败")
	} else {
		ust.udpConn = conn
	}
	// 启动服务器读写协程
	go func() {
		for {
			select {
			case pack := <-ust.sendBuff:
			case pack := <-ust.recvBuff:

			}
		}
	}()
}

func (ust *UdpServerTrigger) OnStop(scoped gecko.GeckoScoped, invoker gecko.TriggerInvoker) {
	ust.withTag(log.Info).Msgf("UDP服务器关闭，绑定地址： %s", ust.udpAddr.String())
	if nil != ust.udpConn {
		ust.udpConn.Close()
	}
}

func (ust *UdpServerTrigger) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "UdpServerTrigger")
}
