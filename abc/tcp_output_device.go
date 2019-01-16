package abc

import (
	"github.com/yoojia/go-gecko"
)

func TCPOutputDeviceFactory() (string, gecko.BundleFactory) {
	return "TCPOutputDevice", func() interface{} {
		return NewTCPOutputDevice()
	}
}

func NewTCPOutputDevice() *TCPOutputDevice {
	return &TCPOutputDevice{
		NetOutputDevice: NewNetOutputDevice(),
	}
}

// TCP客户端输出设备
type TCPOutputDevice struct {
	*NetOutputDevice
}

func (*TCPOutputDevice) Network() string {
	return "tcp"
}
