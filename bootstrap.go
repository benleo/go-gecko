package gecko

import (
	"github.com/yoojia/go-gecko/utils"
)

// Bootstrap提供一个启动入口
func Bootstrap(conf string, prepare func(pipeline *Pipeline)) {
	config, err := utils.LoadConfig(conf)
	if nil != err {
		log.Panicw("加载配置文件出错", "err", err)
	}
	if 0 == len(config) {
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
