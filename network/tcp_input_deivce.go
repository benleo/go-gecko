package network

import (
	"github.com/yoojia/go-gecko"
)

func TCPInputDeviceFactory() (string, gecko.ComponentFactory) {
	return "TCPInputDevice", func() interface{} {
		return NewTCPInputDevice()
	}
}

func NewTCPInputDevice() *TCPInputDevice {
	return &TCPInputDevice{
		AbcNetworkInputDevice: NewAbcNetworkInputDevice("tcp"),
	}
}

// TCP服务器读取设备
type TCPInputDevice struct {
	*AbcNetworkInputDevice
}
