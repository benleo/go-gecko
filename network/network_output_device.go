package network

import (
	"github.com/yoojia/go-gecko/v2"
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
	config      *NetConfig
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
	d.config = structConfig.(*NetConfig)

	read, err := time.ParseDuration(d.config.ReadTimeout)
	if nil != err {
		log.Panic(err)
	}
	write, err := time.ParseDuration(d.config.WriteTimeout)
	if nil != err {
		log.Panic(err)
	}
	d.socket.Init(SocketConfig{
		Type:         d.networkType,
		Addr:         d.config.Address,
		ReadTimeout:  read,
		WriteTimeout: write,
		BufferSize:   d.config.BufferSize,
	})
}

func (d *AbcNetworkOutputDevice) OnStart(ctx gecko.Context) {
	config := d.socket.Config()
	if !config.IsValid() {
		log.Panicw("未设置网络通讯地址和网络类型", "address", config.Addr, "type", config.Type)
	}
	log.Infof("使用%s客户端模式，远程地址： %s", config.Type, config.Addr)
	if err := d.socket.Open(); nil != err {
		log.Errorf("客户端连接失败： %s", config.Addr)
	}
}

func (d *AbcNetworkOutputDevice) OnStop(ctx gecko.Context) {
	if err := d.socket.Close(); nil != err {
		log.Error("客户端断开连接发生错误", err)
	}
}

func (d *AbcNetworkOutputDevice) Process(frame gecko.FramePacket, ctx gecko.Context) (gecko.FramePacket, error) {
	socket := d.Socket()
	buffer := make([]byte, socket.BufferSize())
	if _, err := socket.Send(frame); nil != err {
		return nil, err
	}
	if d.config.Broadcast {
		return []byte(`{"status": "success", "broadcast": "true"}`), nil
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

func (d *AbcNetworkOutputDevice) VendorName() string {
	return "GoGecko/Output/" + d.networkType
}

func (d *AbcNetworkOutputDevice) Description() string {
	return `使用TCP/UDP通信协议的输出虚拟设备，作为客户端上报数据`
}
