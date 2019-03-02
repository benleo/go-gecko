package network

import (
	"github.com/parkingwang/go-conf"
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
	networkType string
	socket      *SocketClient
}

func (d *AbcNetworkOutputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcOutputDevice.OnInit(config, ctx)
	d.socket.Init(SocketConfig{
		Type:         d.networkType,
		Addr:         config.MustString("networkAddress"),
		ReadTimeout:  config.GetDurationOrDefault("readTimeout", time.Second*5),
		WriteTimeout: config.GetDurationOrDefault("writeTimeout", time.Second*5),
		BufferSize:   uint(config.GetInt64OrDefault("bufferSize", 512)),
	})
}

func (d *AbcNetworkOutputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger
	config := d.socket.Config()
	if config.IsValid() {
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

func (d *AbcNetworkOutputDevice) Socket() *SocketClient {
	return d.socket
}
