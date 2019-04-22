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

		pipeline.AddFactory(network.UDPInputDeviceFactory())
		pipeline.AddFactory(network.UDPOutputDeviceFactory())
		pipeline.AddFactory(network.TCPInputDeviceFactory())
		pipeline.AddFactory(network.TCPOutputDeviceFactory())
		pipeline.AddFactory(serial.UARTInputDeviceFactory())
		pipeline.AddFactory(serial.UARTOutputDeviceFactory())

		pipeline.AddFactory(nop.NopTriggerFactory())
		pipeline.AddFactory(nop.NopInputDeviceFactory())
		pipeline.AddFactory(nop.NopDriverFactory())
		pipeline.AddFactory(nop.NopInterceptorFactor())
		pipeline.AddFactory(nop.NopPluginFactory())
		pipeline.AddFactory(nop.NopLogicDeviceFactory())
	})
}
