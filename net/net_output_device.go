package net

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
	"net"
	"time"
)

func NewAbcNetOutputDevice(network string) *AbcNetOutputDevice {
	return &AbcNetOutputDevice{
		AbcOutputDevice: gecko.NewAbcOutputDevice(),
		network:         network,
	}
}

// Socket客户端输出设备
type AbcNetOutputDevice struct {
	*gecko.AbcOutputDevice
	maxBufferSize int64
	writeTimeout  time.Duration
	netConn       net.Conn
	network       string
}

func (no *AbcNetOutputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	no.maxBufferSize = config.GetInt64OrDefault("bufferSize", 512)
	no.writeTimeout = config.GetDurationOrDefault("writeTimeout", time.Second*10)
}

func (no *AbcNetOutputDevice) OnStart(ctx gecko.Context) {
	address := no.GetAddress().GetUnionAddress()
	zap := gecko.Zap()
	defer zap.Sync()
	zap.Infof("使用%s客户端模式，远程地址： %s", no.network, address)
	if "udp" == no.network {
		if addr, err := net.ResolveUDPAddr("udp", address); err != nil {
			zap.Panicw("无法创建UDP地址", "addr", address, "err", err)
		} else {
			if conn, err := net.DialUDP("udp", nil, addr); nil != err {
				zap.Panicw("无法连接UDP服务端", "addr", address, "err", err)
			} else {
				no.netConn = conn
			}
		}
	} else if "tcp" == no.network {
		// TODO 应当支持自动连接
		if conn, err := net.Dial("tcp", address); nil != err {
			zap.Panicw("无法连接TCP服务端", "addr", address, "err", err)
		} else {
			no.netConn = conn
		}
	} else {
		zap.Panicf("未识别的网络连接模式： %s", no.network)
	}
}

func (no *AbcNetOutputDevice) OnStop(ctx gecko.Context) {
	if nil != no.netConn {
		no.netConn.Close()
	}
}

func (no *AbcNetOutputDevice) Process(frame gecko.PacketFrame, ctx gecko.Context) (gecko.PacketFrame, error) {
	if err := no.netConn.SetWriteDeadline(time.Now().Add(no.writeTimeout)); nil != err {
		return nil, err
	}
	if _, err := no.netConn.Write(frame); nil != err {
		return nil, err
	} else {
		return gecko.PacketFrame([]byte{}), nil
	}
}
