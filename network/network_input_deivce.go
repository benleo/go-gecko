package network

import (
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko"
	"net"
	"time"
)

func NewAbcNetworkInputDevice(network string) *AbcNetworkInputDevice {
	return &AbcNetworkInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
		networkType:    network,
		socket:         NewSocketServer(),
		receiveHandler: func(addr net.Addr, data []byte, ctx gecko.Context, deliverer gecko.InputDeliverer) (resp []byte, err error) {
			return []byte("AbstractReceiveHandler!!Implement should set receive handler"), nil
		},
	}
}

// 服务端接收数据处理函数
type OnDataReceiveHandler func(addr net.Addr, data []byte, ctx gecko.Context, deliverer gecko.InputDeliverer) (resp []byte, err error)

// Socket服务器读取设备
type AbcNetworkInputDevice struct {
	*gecko.AbcInputDevice
	networkType    string
	topic          string
	socket         *SocketServer
	receiveHandler OnDataReceiveHandler
}

func (d *AbcNetworkInputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcInputDevice.OnInit(config, ctx)
	d.topic = config.MustString("topic")
	d.socket.Init(SocketConfig{
		Type:         d.networkType,
		Addr:         config.MustString("networkAddress"),
		ReadTimeout:  config.GetDurationOrDefault("readTimeout", time.Second*3),
		WriteTimeout: config.GetDurationOrDefault("writeTimeout", time.Second*3),
		BufferSize:   uint(config.GetInt64OrDefault("bufferSize", 512)),
	})
}

func (d *AbcNetworkInputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger
	config := d.socket.Config()
	if !config.IsValid() {
		zlog.Panicw("未设置网络通讯地址和网络类型", "address", config.Addr, "type", config.Type)
	}
	zlog.Infof("使用%s服务模式，绑定地址： %s", config.Type, config.Addr)
}

func (d *AbcNetworkInputDevice) OnStop(ctx gecko.Context) {
	d.socket.Shutdown()
}

func (d *AbcNetworkInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	handler := func(addr net.Addr, data []byte) []byte {
		if data, err := d.receiveHandler(addr, data, ctx, deliverer); nil != err {
			gecko.ZapSugarLogger.Errorf("处理数据函数发生错误", err)
			return []byte{}
		} else {
			return data
		}
	}
	if err := d.socket.Serve(handler); nil != err {
		return err
	} else {
		return nil
	}
}

// 由于不需要返回响应数据到NetInputDevice，Encoder编码器可以不做业务处理
func (d *AbcNetworkInputDevice) GetEncoder() gecko.Encoder {
	return gecko.NopEncoder
}

func (d *AbcNetworkInputDevice) Topic() string {
	return d.topic
}

// 设置Serve接收到数据处理函数
func (d *AbcNetworkInputDevice) SetDataReceiveHandler(handler OnDataReceiveHandler) {
	d.receiveHandler = handler
}
