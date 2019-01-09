package bundles

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"net"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// UDP服务端Trigger
type UdpServerTrigger struct {
	gecko.Trigger
	udpAddr *net.UDPAddr
	udpConn *net.UDPConn
	// 数据序列化与反序列化
	decoder gecko.Decoder
	encoder gecko.Encoder
	topic   string
	// UDP发送/接收设置
	sendTimeout time.Duration
	recvTimeout time.Duration
	bufferSize  int64
	//
	shutdownCompleted chan struct{}
}

func (ust *UdpServerTrigger) OnInit(args map[string]interface{}, scoped gecko.GeckoScoped) {
	ust.shutdownCompleted = make(chan struct{}, 1)
	ust.decoder = gecko.JSONDefaultDecoder
	ust.encoder = gecko.JSONDefaultEncoder
	config := conf.MapToMap(args)

	address := config.MustString("listenAddress")
	if addr, err := net.ResolveUDPAddr("udp", address); nil != err {
		ust.withTag(log.Panic).Err(err).Msgf("UDP服务端绑定地址错误: %s", address)
	} else {
		ust.udpAddr = addr
	}
	ust.bufferSize = config.GetInt64OrDefault("bufferSizeKB", 1) * 1024
	ust.sendTimeout = config.GetDurationOrDefault("sendTimeout", time.Second*3)
	ust.recvTimeout = config.GetDurationOrDefault("recvTimeout", time.Second*3)
	ust.withTag(log.Info).Msg("UDP服务器Trigger初始化")
}

func (ust *UdpServerTrigger) OnStart(scoped gecko.GeckoScoped, invoker gecko.Invoker) {
	ust.withTag(log.Info).Msgf("UDP服务器启动，绑定地址： %s", ust.udpAddr.String())
	if conn, err := net.ListenUDP("udp", ust.udpAddr); nil != err {
		ust.withTag(log.Panic).Err(err).Msg("UDP服务器启动绑定失败")
	} else {
		ust.udpConn = conn
	}
	go func(shouldBreak <-chan struct{}) {
		process := func() {
			// 读取客户端数据
			buffer := make([]byte, ust.bufferSize)
			ust.udpConn.SetReadDeadline(time.Now().Add(ust.recvTimeout))
			if n, addr, reErr := ust.udpConn.ReadFromUDP(buffer); nil != reErr {
				ust.withTag(log.Error).Err(reErr).Msg("UDP服务器读取数据失败")
			} else {
				// 使用Invoker调度内部系统处理，完成后返回给客户端
				if json, deErr := ust.decoder(buffer[:n]); nil == deErr {
					income := gecko.NewTriggerEvent(ust.topic, json)
					invoker(income, func(data map[string]interface{}) {
						ust.udpConn.SetWriteDeadline(time.Now().Add(ust.sendTimeout))
						if bytes, enErr := ust.encoder(data); nil == deErr {
							if _, err := ust.udpConn.WriteToUDP(bytes, addr); nil != err {
								ust.withTag(log.Error).Err(err).Msg("服务器返回客户端数据失败：" + addr.String())
							}
						} else {
							ust.withTag(log.Error).Err(enErr).Msg("服务器无法序列化的数据")
						}
					})
				} else {
					ust.withTag(log.Error).Err(deErr).Msg("服务器接收到无法解析的数据：" + addr.String())
				}
			}
		}
		// loop
		for {
			select {
			case <-shouldBreak:
				return

			default:
				process()
			}
		}
	}(ust.shutdownCompleted)
}

func (ust *UdpServerTrigger) OnStop(scoped gecko.GeckoScoped, invoker gecko.Invoker) {
	ust.withTag(log.Info).Msgf("UDP服务器关闭，绑定地址： %s", ust.udpAddr.String())
	if nil != ust.udpConn {
		ust.udpConn.Close()
	}
	ust.shutdownCompleted <- struct{}{}
}

func (ust *UdpServerTrigger) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return fun().Str("tag", "UdpServerTrigger")
}
