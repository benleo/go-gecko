package gecko

import (
	"github.com/parkingwang/go-conf"
)

// Bootstrap提供一个启动入口
func Bootstrap(prepare func(pipeline *Pipeline)) {
	zap := Zap()
	defer zap.Sync()
	config, err := cfg.LoadConfig("conf.d")
	if nil != err {
		zap.Panicw("加载配置文件出错", "err", err)
	}
	if config.IsEmpty() {
		zap.Panic("没有任何配置信息")
	}
	pipeline := SharedPipeline()
	prepare(pipeline)
	// Run Pipeline
	pipeline.Init(config)
	pipeline.Start()
	defer pipeline.Stop()
	pipeline.AwaitTermination()
}
