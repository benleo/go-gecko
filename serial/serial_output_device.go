package serial

import (
	"github.com/parkingwang/go-conf"
	"github.com/tarm/serial"
	"github.com/yoojia/go-gecko"
	"time"
)

func NewAbcSerialOutputDevice() *AbcSerialOutputDevice {
	return &AbcSerialOutputDevice{
		AbcOutputDevice: gecko.NewAbcOutputDevice(),
	}
}

// SerialPort客户端输出设备
type AbcSerialOutputDevice struct {
	*gecko.AbcOutputDevice
	config     *serial.Config
	port       *serial.Port
	bufferSize int
}

func (d *AbcSerialOutputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcOutputDevice.OnInit(config, ctx)
	d.bufferSize = int(config.MustInt64("bufferSize"))
	d.config = &serial.Config{
		Name:     config.MustString("address"),
		Baud:     int(config.MustInt64("baud")),
		Size:     byte(config.MustInt64("size")),
		Parity:   serial.Parity(config.MustInt64("parity")),
		StopBits: serial.StopBits(config.MustInt64("stopBits")),
		// 设置ReadTime，在Port读数据时，会超时返回
		ReadTimeout: config.GetDurationOrDefault("readTimeout", time.Millisecond*5),
	}
}

func (d *AbcSerialOutputDevice) OnStart(ctx gecko.Context) {
	if port, err := serial.OpenPort(d.config); nil != err {
		gecko.ZapSugarLogger.Fatalf("打开串口设备发生错误", err)
	} else {
		d.port = port
	}
}

func (d *AbcSerialOutputDevice) OnStop(ctx gecko.Context) {
	if nil != d.port {
		if err := d.port.Close(); nil != err {
			gecko.ZapSugarLogger.Fatalf("关闭串口设备发生错误", err)
		}
	}
}

func (d *AbcSerialOutputDevice) SerialPort() *serial.Port {
	return d.port
}

func (d *AbcSerialOutputDevice) BufferSize() int {
	return d.bufferSize
}

func (d *AbcSerialOutputDevice) Process(frame gecko.FramePacket, ctx gecko.Context) (gecko.FramePacket, error) {
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
