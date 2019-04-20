package serial

import (
	"context"
	"github.com/tarm/serial"
	"github.com/yoojia/go-gecko"
	"github.com/yoojia/go-value"
)

func UARTInputDeviceFactory() (string, gecko.Factory) {
	return "UARTInputDevice", func() interface{} {
		return NewUARTInputDevice()
	}
}

func NewUARTInputDevice() *UARTInputDevice {
	return &UARTInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
	}
}

// UART客户端输入设备
type UARTInputDevice struct {
	*gecko.AbcInputDevice
	gecko.Initial
	config       *serial.Config
	port         *serial.Port
	bufferSize   int
	closeContext context.Context
	closeFunc    context.CancelFunc
}

func (d *UARTInputDevice) OnInit(config map[string]interface{}, ctx gecko.Context) {
	d.bufferSize = int(value.Of(config["bufferSize"]).MustInt64())
	// 如果设置Read超时，port.Read方法会启用NonBlocking读模式。
	// 此处设置为0，使用阻塞读模式。
	d.config = getSerialConfig(config, 0)
}

func (d *UARTInputDevice) OnStart(ctx gecko.Context) {
	log.Debugf("打开串口设备: %s", d.config.Name)
	if port, err := serial.OpenPort(d.config); nil != err {
		log.Fatalf("打开串口设备发生错误", err)
	} else {
		d.port = port
	}
}

func (d *UARTInputDevice) OnStop(ctx gecko.Context) {
	if nil != d.port {
		if err := d.port.Close(); nil != err {
			log.Fatalf("关闭串口设备发生错误", err)
		}
	}
}

func (d *UARTInputDevice) Port() *serial.Port {
	return d.port
}

func (d *UARTInputDevice) BufferSize() int {
	return d.bufferSize
}

func (d *UARTInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	port := d.Port()
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

func (d *UARTInputDevice) VendorName() string {
	return "GoGecko/UART/Input"
}

func (d *UARTInputDevice) Description() string {
	return `使用UART串口通信协议的输入虚拟设备，作为主设备接收数据`
}
