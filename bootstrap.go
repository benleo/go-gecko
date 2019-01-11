package gecko

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Bootstrap提供一个启动入口
func Bootstrap(prepare func(engine *Engine)) {
	config, err := conf.LoadConfig("conf.d")
	if nil != err {
		_bootstrapTag(log.Panic).Err(err).Msg("加载配置文件出错")
	}
	if len(config) <= 0 {
		_bootstrapTag(log.Panic).Msgf("没有任何配置信息")
	}
	engine := SharedEngine()
	prepare(engine)
	// Run Engine
	engine.Init(config)
	engine.Start()
	defer engine.Stop()
	engine.AwaitTermination()
}

func _bootstrapTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Bootstrap")
}
