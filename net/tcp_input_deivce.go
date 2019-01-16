package net

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
		AbcNetInputDevice: NewAbcNetInputDevice("tcp"),
	}
}

// TCP服务器读取设备
type TCPInputDevice struct {
	*AbcNetInputDevice
}
