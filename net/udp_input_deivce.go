package net

import (
	"github.com/yoojia/go-gecko"
)

func UDPInputDeviceFactory() (string, gecko.BundleFactory) {
	return "UDPInputDevice", func() interface{} {
		return NewUDPInputDevice()
	}
}

func NewUDPInputDevice() *UDPInputDevice {
	return &UDPInputDevice{
		AbcNetInputDevice: NewAbcNetInputDevice("udp"),
	}
}

// UDP服务器读取设备
type UDPInputDevice struct {
	*AbcNetInputDevice
}
