package main

import (
	"github.com/yoojia/go-gecko"
	"github.com/yoojia/go-gecko/net"
	"github.com/yoojia/go-gecko/nop"
)

// Main
func main() {
	// 默认Log方式
	gecko.Bootstrap(func(pipeline *gecko.Pipeline) {
		// 通常使用这个函数来注册组件工厂函数
		pipeline.AddCodecFactory(gecko.JSONDefaultEncoderFactory())
		pipeline.AddCodecFactory(gecko.JSONDefaultDecoderFactory())

		pipeline.AddBundleFactory(net.UDPInputDeviceFactory())
		pipeline.AddBundleFactory(net.UDPOutputDeviceFactory())
		pipeline.AddBundleFactory(net.TCPInputDeviceFactory())
		pipeline.AddBundleFactory(net.TCPOutputDeviceFactory())

		pipeline.AddBundleFactory(nop.NopDriverFactory())
		pipeline.AddBundleFactory(nop.NopInterceptorFactor())
		pipeline.AddBundleFactory(nop.NopPluginFactory())
	})
}
