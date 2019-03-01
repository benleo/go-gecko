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

// 输出设备读取响应的外置处理函数
type OutputReceiver func(conn net.Conn) (gecko.FramePacket, error)

// Socket客户端输出设备
type AbcNetOutputDevice struct {
	*gecko.AbcOutputDevice
	maxBufferSize  int64
	writeTimeout   time.Duration
	readTimeout    time.Duration
	netConn        net.Conn
	networkType    string
	networkAddress string
	outputReceiver OutputReceiver
}

func (d *AbcNetOutputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcOutputDevice.OnInit(config, ctx)
	d.maxBufferSize = config.GetInt64OrDefault("bufferSize", 512)
	d.writeTimeout = config.GetDurationOrDefault("writeTimeout", time.Second*5)
	d.readTimeout = config.GetDurationOrDefault("readTimeout", time.Second*5)
	d.networkAddress = config.MustString("networkAddress")

	// 输出默认不读取响应结果
	d.SetOutputReceiver(func(conn net.Conn) (gecko.FramePacket, error) {
		return gecko.NewFramePacket([]byte{}), nil
	})
}

func (d *AbcNetOutputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger()

	if d.networkAddress == "" || d.networkType == "" {
		zlog.Panicw("未设置网络通讯地址和网络类型", "address", d.networkAddress, "type", d.networkType)
	}
	zlog.Infof("使用%s客户端模式，远程地址： %s", d.networkType, d.networkAddress)
	if "udp" == d.networkType {
		if addr, err := net.ResolveUDPAddr("udp", d.networkAddress); err != nil {
			zlog.Panicw("无法创建UDP地址", "addr", d.networkAddress, "err", err)
		} else {
			if conn, err := net.DialUDP("udp", nil, addr); nil != err {
				zlog.Panicw("无法连接UDP服务端", "addr", d.networkAddress, "err", err)
			} else {
				d.netConn = conn
			}
		}
	} else if "tcp" == d.networkType {
		// TODO 应当支持自动连接
		if conn, err := net.Dial("tcp", d.networkAddress); nil != err {
			zlog.Panicw("无法连接TCP服务端", "addr", d.networkAddress, "err", err)
		} else {
			d.netConn = conn
		}
	} else {
		zlog.Panicf("未识别的网络连接模式： %s", d.networkType)
	}
}

func (d *AbcNetOutputDevice) OnStop(ctx gecko.Context) {
	if nil != d.netConn {
		d.netConn.Close()
	}
}

func (d *AbcNetOutputDevice) Process(frame gecko.FramePacket, ctx gecko.Context) (gecko.FramePacket, error) {
	// 写
	if err := d.netConn.SetWriteDeadline(time.Now().Add(d.writeTimeout)); nil != err {
		return nil, err
	}
	if _, err := d.netConn.Write(frame); nil != err {
		return nil, err
	}
	// 读
	if err := d.netConn.SetReadDeadline(time.Now().Add(d.writeTimeout)); nil != err {
		return nil, err
	}
	return d.outputReceiver(d.netConn)
}

func (d *AbcNetOutputDevice) SetOutputReceiver(receiver OutputReceiver) {
	d.outputReceiver = receiver
}
