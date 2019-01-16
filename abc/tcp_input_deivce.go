package abc

import (
	"github.com/yoojia/go-gecko"
)

func TCPInputDeviceFactory() (string, gecko.BundleFactory) {
	return "TCPInputDevice", func() interface{} {
		return NewTCPInputDevice()
	}
}

func NewTCPInputDevice() *TCPInputDevice {
	return &TCPInputDevice{
		NetInputDevice: NewNetInputDevice("tcp"),
	}
}

// TCP服务器读取设备
type TCPInputDevice struct {
	*NetInputDevice
}
