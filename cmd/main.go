package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"github.com/yoojia/go-gecko/nop"
	"os"
)

// Main
func main() {
	// 默认Log方式
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	gecko.Bootstrap(func(engine *gecko.Pipeline) {
		// 通常使用这个函数来注册组件工厂函数
		engine.RegisterBundleFactory(nop.NopDriverFactory())
		engine.RegisterBundleFactory(nop.NopInterceptorFactor())
		engine.RegisterBundleFactory(nop.NopPluginFactory())
	})
}
