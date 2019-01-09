package gecko

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko/bundles"
	"os"
	"os/signal"
	"syscall"
)

func Bootstrap(prepare func(engine *GeckoEngine)) {
	config := LoadConfig("conf.d")
	if len(config) <= 0 {
		_bootstrapTag(log.Panic).Msgf("没有任何配置信息")
	}
	engine := SharedEngine()
	// 内置组件工厂函数
	engine.RegisterBundleFactory(bundles.NetworkServerTriggerFactory())
	prepare(engine)
	// Run Engine
	engine.Init(config)
	engine.Start()
	defer engine.Stop()
	// 等待终止信号
	sysSignal := make(chan os.Signal, 1)
	signal.Notify(sysSignal, syscall.SIGINT, syscall.SIGTERM)
	<-sysSignal
	_bootstrapTag(log.Warn).Msgf("接收到系统停止信号")
}

func _bootstrapTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Bootstrap")
}
