package network

import (
	"github.com/yoojia/go-gecko"
)

func UDPInputDeviceFactory() (string, gecko.ComponentFactory) {
	return "UDPInputDevice", func() interface{} {
		return NewUDPInputDevice()
	}
}

func NewUDPInputDevice() *UDPInputDevice {
	return &UDPInputDevice{
		AbcNetworkInputDevice: NewAbcNetworkInputDevice("udp"),
	}
}

// UDP服务器读取设备
type UDPInputDevice struct {
	*AbcNetworkInputDevice
}
