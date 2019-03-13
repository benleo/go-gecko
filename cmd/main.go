package main

import (
	"github.com/yoojia/go-gecko"
	"github.com/yoojia/go-gecko/network"
	"github.com/yoojia/go-gecko/nop"
	"github.com/yoojia/go-gecko/serial"
)

// Main
func main() {
	// 默认Log方式
	gecko.Bootstrap(func(pipeline *gecko.Pipeline) {
		// 通常使用这个函数来注册组件工厂函数
		pipeline.AddCodecFactory(gecko.JSONDefaultEncoderFactory())
		pipeline.AddCodecFactory(gecko.JSONDefaultDecoderFactory())

		pipeline.AddBundleFactory(network.UDPInputDeviceFactory())
		pipeline.AddBundleFactory(network.UDPOutputDeviceFactory())
		pipeline.AddBundleFactory(network.TCPInputDeviceFactory())
		pipeline.AddBundleFactory(network.TCPOutputDeviceFactory())
		pipeline.AddBundleFactory(serial.SerialPortInputDeviceFactory())
		pipeline.AddBundleFactory(serial.SerialPortOutputDeviceFactory())

		pipeline.AddBundleFactory(nop.NopDriverFactory())
		pipeline.AddBundleFactory(nop.NopInterceptorFactor())
		pipeline.AddBundleFactory(nop.NopPluginFactory())
		pipeline.AddBundleFactory(nop.NopShadowDeviceFactory())
	})
}
