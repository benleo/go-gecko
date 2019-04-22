package main

import (
	"flag"
	"github.com/yoojia/go-gecko/v2"
	"github.com/yoojia/go-gecko/v2/lua"
	"github.com/yoojia/go-gecko/v2/network"
	"github.com/yoojia/go-gecko/v2/nop"
	"github.com/yoojia/go-gecko/v2/serial"
)

// Main
func main() {
	confPtr := flag.String("c", "conf.d", "a file or dir path")
	// 默认Log方式
	gecko.Bootstrap(*confPtr, func(pipeline *gecko.Pipeline) {
		// 通常使用这个函数来注册组件工厂函数
		pipeline.AddCodecFactory(gecko.JSONDefaultEncoderFactory())
		pipeline.AddCodecFactory(gecko.JSONDefaultDecoderFactory())

		pipeline.AddFactory(lua.ScriptDriverFactory())

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
