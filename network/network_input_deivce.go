package network

import (
	"github.com/yoojia/go-gecko/v2"
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
	return &NetConfig{
		ReadTimeout:  "3s",
		WriteTimeout: "3s",
	}
}

func (d *AbcNetworkInputDevice) Init(structConfig interface{}, ctx gecko.Context) {
	config := structConfig.(*NetConfig)
	read, err := time.ParseDuration(config.ReadTimeout)
	if nil != err {
		log.Panic(err)
	}
	write, err := time.ParseDuration(config.WriteTimeout)
	if nil != err {
		log.Panic(err)
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
	config := d.socket.Config()
	if !config.IsValid() {
		log.Panicw("未设置网络通讯地址和网络类型", "address", config.Addr, "type", config.Type)
	}
	log.Infof("使用%s服务模式，绑定地址：%s", config.Type, config.Addr)
}

func (d *AbcNetworkInputDevice) OnStop(ctx gecko.Context) {
	d.socket.Shutdown()
}

func (d *AbcNetworkInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	return d.Socket().Serve(func(addr net.Addr, input []byte) (output []byte, err error) {
		return deliverer.Deliver(d.GetTopic(), gecko.FramePacket(input))
	})
}

func (d *AbcNetworkInputDevice) Socket() *SocketServer {
	return d.socket
}

func (d *AbcNetworkInputDevice) VendorName() string {
	return "GoGecko/Input/" + d.networkType
}

func (d *AbcNetworkInputDevice) Description() string {
	return `使用TCP/UDP通信协议的输入虚拟设备，作为服务端接收数据`
}
