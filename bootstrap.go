package gecko

import (
	"github.com/parkingwang/go-conf"
)

// Bootstrap提供一个启动入口
func Bootstrap(prepare func(pipeline *Pipeline)) {
	log := ZapSugarLogger
	config, err := cfg.LoadConfig("conf.d")
	if nil != err {
		log.Panicw("加载配置文件出错", "err", err)
	}
	if config.IsEmpty() {
		log.Panic("没有任何配置信息")
	}
	pipeline := SharedPipeline()
	prepare(pipeline)
	// Run Pipeline
	pipeline.Init(config)
	pipeline.Start()
	defer pipeline.Stop()
	pipeline.AwaitTermination()
}
