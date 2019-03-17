package network

import (
	"github.com/yoojia/go-gecko"
	"net"
	"time"
)

func NewAbcNetworkInputDevice(network string) *AbcNetworkInputDevice {
	return &AbcNetworkInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
		networkType:    network,
		socket:         NewSocketServer(),
	}
}

// Socket服务器读取设备
type AbcNetworkInputDevice struct {
	*gecko.AbcInputDevice
	gecko.StructuredInitial
	gecko.LifeCycle

	networkType string
	socket      *SocketServer
}

func (d *AbcNetworkInputDevice) StructuredConfig() interface{} {
	return &NetConfig{}
}

func (d *AbcNetworkInputDevice) Init(structConfig interface{}, ctx gecko.Context) {
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

func (d *AbcNetworkInputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger
	config := d.socket.Config()
	if !config.IsValid() {
		zlog.Panicw("未设置网络通讯地址和网络类型", "address", config.Addr, "type", config.Type)
	}
	zlog.Infof("使用%s服务模式，绑定地址：%s", config.Type, config.Addr)
}

func (d *AbcNetworkInputDevice) OnStop(ctx gecko.Context) {
	d.socket.Shutdown()
}

func (d *AbcNetworkInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	return d.Socket().Serve(func(addr net.Addr, input []byte) (output []byte, err error) {
		return deliverer.Deliver(d.GetTopic(), gecko.NewFramePacket(input))
	})
}

func (d *AbcNetworkInputDevice) Socket() *SocketServer {
	return d.socket
}
