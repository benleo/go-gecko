package serial

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/tarm/serial"
	"github.com/yoojia/go-gecko"
)

func NewAbcSerialInputDevice() *AbcSerialInputDevice {
	return &AbcSerialInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
	}
}

// SerialPort客户端输入设备
type AbcSerialInputDevice struct {
	*gecko.AbcInputDevice
	config       *serial.Config
	port         *serial.Port
	bufferSize   int
	closeContext context.Context
	closeFunc    context.CancelFunc
}

func (d *AbcSerialInputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcInputDevice.OnInit(config, ctx)
	d.bufferSize = int(config.MustInt64("bufferSize"))
	d.config = &serial.Config{
		Name:     config.MustString("address"),
		Baud:     int(config.MustInt64("baud")),
		Size:     byte(config.MustInt64("size")),
		Parity:   serial.Parity(config.MustInt64("parity")),
		StopBits: serial.StopBits(config.MustInt64("stopBits")),
		// 如果设置Read超时，port.Read方法会启用NonBlocking读模式。
		// 此处设置为0，使用阻塞读模式。
		ReadTimeout: 0,
	}
}

func (d *AbcSerialInputDevice) OnStart(ctx gecko.Context) {
	if port, err := serial.OpenPort(d.config); nil != err {
		gecko.ZapSugarLogger.Fatalf("打开串口设备发生错误", err)
	} else {
		d.port = port
	}
}

func (d *AbcSerialInputDevice) OnStop(ctx gecko.Context) {
	if nil != d.port {
		if err := d.port.Close(); nil != err {
			gecko.ZapSugarLogger.Fatalf("关闭串口设备发生错误", err)
		}
	}
}

func (d *AbcSerialInputDevice) SerialPort() *serial.Port {
	return d.port
}

func (d *AbcSerialInputDevice) BufferSize() int {
	return d.bufferSize
}

func (d *AbcSerialInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	port := d.SerialPort()
	buffer := make([]byte, d.BufferSize())
	for {
		if n, err := port.Read(buffer); nil != err {
			// FIXME 需要处理Port被Close后的Error状态
			return err
		} else {
			output, err := deliverer.Execute(d.GetTopic(), buffer[:n])
			if nil != err {
				return err
			}
			if _, err := port.Write(output); nil != err {
				return err
			}
		}
	}
}
