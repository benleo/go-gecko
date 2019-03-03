package serial

import (
	"github.com/parkingwang/go-conf"
	"github.com/tarm/serial"
	"github.com/yoojia/go-gecko"
	"time"
)

func SerialPortOutputDeviceFactory() (string, gecko.BundleFactory) {
	return "SerialPortOutputDevice", func() interface{} {
		return NewSerialOutputDevice()
	}
}

func NewSerialOutputDevice() *SerialPortOutputDevice {
	return &SerialPortOutputDevice{
		AbcOutputDevice: gecko.NewAbcOutputDevice(),
	}
}

// SerialPort客户端输出设备
type SerialPortOutputDevice struct {
	*gecko.AbcOutputDevice
	config     *serial.Config
	port       *serial.Port
	bufferSize int
}

func (d *SerialPortOutputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcOutputDevice.OnInit(config, ctx)
	d.bufferSize = int(config.MustInt64("bufferSize"))
	d.config = getSerialConfig(config, time.Millisecond*100)
}

func (d *SerialPortOutputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger
	zlog.Debugf("打开串口设备: %s", d.config.Name)
	if port, err := serial.OpenPort(d.config); nil != err {
		zlog.Fatalf("打开串口设备发生错误", err)
	} else {
		d.port = port
	}
}

func (d *SerialPortOutputDevice) OnStop(ctx gecko.Context) {
	if nil != d.port {
		if err := d.port.Close(); nil != err {
			gecko.ZapSugarLogger.Fatalf("关闭串口设备发生错误", err)
		}
	}
}

func (d *SerialPortOutputDevice) SerialPort() *serial.Port {
	return d.port
}

func (d *SerialPortOutputDevice) BufferSize() int {
	return d.bufferSize
}

func (d *SerialPortOutputDevice) Process(frame gecko.FramePacket, ctx gecko.Context) (gecko.FramePacket, error) {
	port := d.SerialPort()
	buffer := make([]byte, d.BufferSize())
	if _, err := port.Write(frame); nil != err {
		return nil, err
	}
	if n, err := port.Read(buffer); nil != err {
		return nil, err
	} else {
		return gecko.NewFramePacket(buffer[:n]), nil
	}
}
