package abc

import (
	"github.com/yoojia/go-gecko"
)

func UDPOutputDeviceFactory() (string, gecko.BundleFactory) {
	return "UDPOutputDevice", func() interface{} {
		return NewUDPOutputDevice()
	}
}

func NewUDPOutputDevice() *UDPOutputDevice {
	return &UDPOutputDevice{
		NetOutputDevice: NewNetOutputDevice("udp"),
	}
}

// UDP客户端输出设备
type UDPOutputDevice struct {
	*NetOutputDevice
}
