package serial

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/tarm/serial"
	"github.com/yoojia/go-gecko"
)

func SerialPortInputDeviceFactory() (string, gecko.ComponentFactory) {
	return "SerialPortInputDevice", func() interface{} {
		return NewSerialInputDevice()
	}
}

func NewSerialInputDevice() *SerialPortInputDevice {
	return &SerialPortInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
	}
}

// SerialPort客户端输入设备
type SerialPortInputDevice struct {
	*gecko.AbcInputDevice
	config       *serial.Config
	port         *serial.Port
	bufferSize   int
	closeContext context.Context
	closeFunc    context.CancelFunc
}

func (d *SerialPortInputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.bufferSize = int(config.MustInt64("bufferSize"))
	// 如果设置Read超时，port.Read方法会启用NonBlocking读模式。
	// 此处设置为0，使用阻塞读模式。
	d.config = getSerialConfig(config, 0)
}

func (d *SerialPortInputDevice) OnStart(ctx gecko.Context) {
	zlog := gecko.ZapSugarLogger
	zlog.Debugf("打开串口设备: %s", d.config.Name)
	if port, err := serial.OpenPort(d.config); nil != err {
		zlog.Fatalf("打开串口设备发生错误", err)
	} else {
		d.port = port
	}
}

func (d *SerialPortInputDevice) OnStop(ctx gecko.Context) {
	if nil != d.port {
		if err := d.port.Close(); nil != err {
			gecko.ZapSugarLogger.Fatalf("关闭串口设备发生错误", err)
		}
	}
}

func (d *SerialPortInputDevice) SerialPort() *serial.Port {
	return d.port
}

func (d *SerialPortInputDevice) BufferSize() int {
	return d.bufferSize
}

func (d *SerialPortInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	port := d.SerialPort()
	buffer := make([]byte, d.BufferSize())
	for {
		if n, err := port.Read(buffer); nil != err {
			// FIXME 需要处理Port被Close后的Error状态
			return err
		} else {
			output, err := deliverer.Deliver(d.GetTopic(), buffer[:n])
			if nil != err {
				return err
			}
			if _, err := port.Write(output); nil != err {
				return err
			}
		}
	}
}
