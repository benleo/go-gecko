package serial

import (
	"github.com/tarm/serial"
	"github.com/yoojia/go-gecko/v2"
	"github.com/yoojia/go-value"
	"time"
)

func UARTOutputDeviceFactory() (string, gecko.Factory) {
	return "UARTOutputDevice", func() interface{} {
		return NewUARTOutputDevice()
	}
}

func NewUARTOutputDevice() *UARTOutputDevice {
	return &UARTOutputDevice{
		AbcOutputDevice: gecko.NewAbcOutputDevice(),
	}
}

// UART客户端输出设备
type UARTOutputDevice struct {
	*gecko.AbcOutputDevice
	gecko.Initial
	config     *serial.Config
	port       *serial.Port
	bufferSize int
	broadcast  bool
}

func (d *UARTOutputDevice) OnInit(config map[string]interface{}, ctx gecko.Context) {
	d.bufferSize = int(value.Of(config["bufferSize"]).MustInt64())
	d.broadcast = value.Of(config["broadcast"]).MustBool()
	d.config = getSerialConfig(config, time.Millisecond*100)
}

func (d *UARTOutputDevice) OnStart(ctx gecko.Context) {
	log.Debugf("打开串口设备: %s", d.config.Name)
	if port, err := serial.OpenPort(d.config); nil != err {
		log.Fatalf("打开串口设备发生错误", err)
	} else {
		d.port = port
	}
}

func (d *UARTOutputDevice) OnStop(ctx gecko.Context) {
	if nil != d.port {
		if err := d.port.Close(); nil != err {
			log.Fatalf("关闭串口设备发生错误", err)
		}
	}
}

func (d *UARTOutputDevice) Port() *serial.Port {
	return d.port
}

func (d *UARTOutputDevice) BufferSize() int {
	return d.bufferSize
}

func (d *UARTOutputDevice) Process(frame gecko.FramePacket, ctx gecko.Context) (gecko.FramePacket, error) {
	port := d.Port()
	buffer := make([]byte, d.BufferSize())
	if _, err := port.Write(frame); nil != err {
		return nil, err
	}
	if d.broadcast {
		return []byte(`{"status": "success", "broadcast": "true"}`), nil
	}
	if n, err := port.Read(buffer); nil != err {
		return nil, err
	} else {
		return gecko.FramePacket(buffer[:n]), nil
	}
}

func (d *UARTOutputDevice) VendorName() string {
	return "GoGecko/UART/Output"
}

func (d *UARTOutputDevice) Description() string {
	return `使用UART串口通信协议的输出虚拟设备，作为从设备发送数据`
}
