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
		pipeline.RegisterCodecFactory(gecko.JSONDefaultEncoderFactory())
		pipeline.RegisterCodecFactory(gecko.JSONDefaultDecoderFactory())

		pipeline.RegisterBundleFactory(net.UDPInputDeviceFactory())
		pipeline.RegisterBundleFactory(net.UDPOutputDeviceFactory())
		pipeline.RegisterBundleFactory(net.TCPInputDeviceFactory())
		pipeline.RegisterBundleFactory(net.TCPOutputDeviceFactory())

		pipeline.RegisterBundleFactory(nop.NopDriverFactory())
		pipeline.RegisterBundleFactory(nop.NopInterceptorFactor())
		pipeline.RegisterBundleFactory(nop.NopPluginFactory())
	})
}
