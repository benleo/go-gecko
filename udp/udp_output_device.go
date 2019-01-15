package udp

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"net"
	"time"
)

// UDP客户端输出设备
type AbcUdpOutputDevice struct {
	*gecko.AbcOutputDevice
	maxBufferSize int64
	writeTimeout  time.Duration
	udpConn       *net.UDPConn
}

func (ur *AbcUdpOutputDevice) OnInit(args map[string]interface{}, ctx gecko.Context) {
	config := conf.WrapImmutableMap(args)
	ur.maxBufferSize = config.GetInt64OrDefault("bufferSizeKB", 1) * 1024
	ur.writeTimeout = config.GetDurationOrDefault("writeTimeout", time.Second*10)
}

func (ur *AbcUdpOutputDevice) OnStart(ctx gecko.Context) {
	address := ur.GetUnionAddress()
	ur.withTag(log.Info).Msgf("使用UDP客户端模式，远程地址： %s", address)
	if addr, err := net.ResolveUDPAddr("udp", address); err != nil {
		ur.withTag(log.Panic).Err(err).Msgf("无法创建UDP地址： %s", address)
	} else {
		if conn, err := net.DialUDP("udp", nil, addr); nil != err {
			ur.withTag(log.Panic).Err(err).Msgf("无法连接UDP服务端： %s", address)
		} else {
			ur.udpConn = conn
		}
	}
}

func (ur *AbcUdpOutputDevice) OnStop(ctx gecko.Context) {
	if nil != ur.udpConn {
		ur.udpConn.Close()
	}
}

func (ur *AbcUdpOutputDevice) Process(frame gecko.PacketFrame, ctx gecko.Context) (gecko.PacketFrame, error) {
	if err := ur.udpConn.SetWriteDeadline(time.Now().Add(ur.writeTimeout)); nil != err {
		return nil, err
	}
	if _, err := ur.udpConn.Write(frame); nil != err {
		return nil, err
	}
	return gecko.PacketFrame([]byte{}), nil
}

func (ur *AbcUdpOutputDevice) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "AbcUdpOutputDevice")
}
