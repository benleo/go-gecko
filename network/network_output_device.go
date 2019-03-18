package network

import (
	"github.com/yoojia/go-gecko"
	"time"
)

func NewAbcNetworkOutputDevice(network string) *AbcNetworkOutputDevice {
	return &AbcNetworkOutputDevice{
		AbcOutputDevice: gecko.NewAbcOutputDevice(),
		networkType:     network,
		socket:          new(SocketClient),
	}
}

// Socket客户端输出设备
type AbcNetworkOutputDevice struct {
	*gecko.AbcOutputDevice
	gecko.StructuredInitial
	gecko.LifeCycle

	networkType string
	socket      *SocketClient
}

func (d *AbcNetworkOutputDevice) StructuredConfig() interface{} {
	return &NetConfig{
		ReadTimeout:  "3s",
		WriteTimeout: "3s",
	}
}

func (d *AbcNetworkOutputDevice) Init(structConfig interface{}, ctx gecko.Context) {
	config := structConfig.(*NetConfig)
	zlog := gecko.ZapSugarLogger
	read, err := time.ParseDuration(config.ReadTimeout)
	if nil != err {
		zlog.Panic(err)
	}
	write, err := time.ParseDuration(config.WriteTimeout)
	if nil != err {
		zlog.Panic(err)
	}
	d.socket.Init(SocketConfig{
		Type:         d.networkType,
		Addr:         config.Address,
		ReadTimeout:  read,
		WriteTimeout: write,
		BufferSize:   config.BufferSize,
	})
}

func (d *AbcNetworkOutputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger
	config := d.socket.Config()
	if !config.IsValid() {
		zlog.Panicw("未设置网络通讯地址和网络类型", "address", config.Addr, "type", config.Type)
	}
	zlog.Infof("使用%s客户端模式，远程地址： %s", config.Type, config.Addr)
	if err := d.socket.Open(); nil != err {
		zlog.Errorf("客户端连接失败： %s", config.Addr)
	}
}

func (d *AbcNetworkOutputDevice) OnStop(ctx gecko.Context) {
	if err := d.socket.Close(); nil != err {
		gecko.ZapSugarLogger.Error("客户端断开连接发生错误", err)
	}
}

func (d *AbcNetworkOutputDevice) Process(frame gecko.FramePacket, ctx gecko.Context) (gecko.FramePacket, error) {
	socket := d.Socket()
	buffer := make([]byte, socket.BufferSize())
	if _, err := socket.Send(frame); nil != err {
		return nil, err
	}
	if n, err := socket.Receive(buffer); nil != err {
		return nil, err
	} else {
		return gecko.FramePacket(buffer[:n]), nil
	}
}

func (d *AbcNetworkOutputDevice) Socket() *SocketClient {
	return d.socket
}
