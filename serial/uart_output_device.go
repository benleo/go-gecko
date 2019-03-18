package serial

import (
	"github.com/parkingwang/go-conf"
	"github.com/tarm/serial"
	"github.com/yoojia/go-gecko"
	"time"
)

func UARTOutputDeviceFactory() (string, gecko.ComponentFactory) {
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
}

func (d *UARTOutputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.bufferSize = int(config.MustInt64("bufferSize"))
	d.config = getSerialConfig(config, time.Millisecond*100)
}

func (d *UARTOutputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger
	zlog.Debugf("打开串口设备: %s", d.config.Name)
	if port, err := serial.OpenPort(d.config); nil != err {
		zlog.Fatalf("打开串口设备发生错误", err)
	} else {
		d.port = port
	}
}

func (d *UARTOutputDevice) OnStop(ctx gecko.Context) {
	if nil != d.port {
		if err := d.port.Close(); nil != err {
			gecko.ZapSugarLogger.Fatalf("关闭串口设备发生错误", err)
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
	if n, err := port.Read(buffer); nil != err {
		return nil, err
	} else {
		return gecko.FramePacket(buffer[:n]), nil
	}
}
