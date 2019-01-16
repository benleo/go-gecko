package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"github.com/yoojia/go-gecko/abc"
	"github.com/yoojia/go-gecko/nop"
	"os"
)

// Main
func main() {
	// 默认Log方式
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	gecko.Bootstrap(func(pipeline *gecko.Pipeline) {
		// 通常使用这个函数来注册组件工厂函数
		pipeline.RegisterCodecFactory(gecko.JSONDefaultEncoderFactory())
		pipeline.RegisterCodecFactory(gecko.JSONDefaultDecoderFactory())

		pipeline.RegisterBundleFactory(abc.UDPInputDeviceFactory())
		pipeline.RegisterBundleFactory(abc.UDPOutputDeviceFactory())

		pipeline.RegisterBundleFactory(nop.NopDriverFactory())
		pipeline.RegisterBundleFactory(nop.NopInterceptorFactor())
		pipeline.RegisterBundleFactory(nop.NopPluginFactory())
	})
}
