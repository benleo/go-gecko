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
		networkType:     network,
	}
}

// Socket客户端输出设备
type AbcNetOutputDevice struct {
	*gecko.AbcOutputDevice
	maxBufferSize  int64
	writeTimeout   time.Duration
	netConn        net.Conn
	networkType    string
	networkAddress string
}

func (d *AbcNetOutputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcOutputDevice.OnInit(config, ctx)
	d.maxBufferSize = config.GetInt64OrDefault("bufferSize", 512)
	d.writeTimeout = config.GetDurationOrDefault("writeTimeout", time.Second*10)
	d.networkAddress = config.MustString("networkAddress")
}

func (d *AbcNetOutputDevice) OnStart(ctx gecko.Context) {
	zap := gecko.Zap()
	defer zap.Sync()
	if d.networkAddress == "" || d.networkType == "" {
		zap.Panicw("未设置网络通讯地址和网络类型", "address", d.networkAddress, "type", d.networkType)
	}
	zap.Infof("使用%s客户端模式，远程地址： %s", d.networkType, d.networkAddress)
	if "udp" == d.networkType {
		if addr, err := net.ResolveUDPAddr("udp", d.networkAddress); err != nil {
			zap.Panicw("无法创建UDP地址", "addr", d.networkAddress, "err", err)
		} else {
			if conn, err := net.DialUDP("udp", nil, addr); nil != err {
				zap.Panicw("无法连接UDP服务端", "addr", d.networkAddress, "err", err)
			} else {
				d.netConn = conn
			}
		}
	} else if "tcp" == d.networkType {
		// TODO 应当支持自动连接
		if conn, err := net.Dial("tcp", d.networkAddress); nil != err {
			zap.Panicw("无法连接TCP服务端", "addr", d.networkAddress, "err", err)
		} else {
			d.netConn = conn
		}
	} else {
		zap.Panicf("未识别的网络连接模式： %s", d.networkType)
	}
}

func (d *AbcNetOutputDevice) OnStop(ctx gecko.Context) {
	if nil != d.netConn {
		d.netConn.Close()
	}
}

func (d *AbcNetOutputDevice) Process(frame gecko.PacketFrame, ctx gecko.Context) (gecko.PacketFrame, error) {
	if err := d.netConn.SetWriteDeadline(time.Now().Add(d.writeTimeout)); nil != err {
		return nil, err
	}
	if _, err := d.netConn.Write(frame); nil != err {
		return nil, err
	} else {
		return gecko.PacketFrame([]byte{}), nil
	}
}
