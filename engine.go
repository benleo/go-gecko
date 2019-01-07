package gecko

import (
	"github.com/rs/zerolog/log"
	"time"
)

////

// 默诵
const DefaultLifeCycleTimeout = time.Duration(3)

type GeckoEngine struct {
	*RegisterEngine

	scoped  GeckoScoped
	invoker TriggerInvoker
}

// 初始化Engine
func (slf *GeckoEngine) Init(config map[string]interface{}) {

}

// 启动Engine
func (slf *GeckoEngine) Start() {
	slf.withTag(log.Info).Msgf("Engine启动...")
	defer slf.withTag(log.Info).Msgf("Engine启动...OK")
	// Plugin
	for el := slf.plugins.Front(); el != nil; el = el.Next() {
		slf.checkDefTimeout(el.Value.(Plugin).OnStart)
	}
	// Pipeline
	for _, pipeline := range slf.pipelines {
		slf.checkDefTimeout(pipeline.OnStart)
	}
	// Drivers
	for el := slf.drivers.Front(); el != nil; el = el.Next() {
		slf.checkDefTimeout(el.Value.(Driver).OnStart)
	}
	// Trigger
	for el := slf.triggers.Front(); el != nil; el = el.Next() {
		slf.scoped.CheckTimeout(DefaultLifeCycleTimeout, func() {
			el.Value.(Trigger).OnStart(slf.scoped, slf.invoker)
		})
	}
}

// 停止Engine
func (slf *GeckoEngine) Stop() {
	slf.withTag(log.Info).Msgf("Engine停止...")
	defer slf.withTag(log.Info).Msgf("Engine停止...OK")
	// Triggers
	for el := slf.triggers.Front(); el != nil; el = el.Next() {
		slf.scoped.CheckTimeout(DefaultLifeCycleTimeout, func() {
			el.Value.(Trigger).OnStop(slf.scoped, slf.invoker)
		})
	}
	// Drivers
	for el := slf.drivers.Front(); el != nil; el = el.Next() {
		slf.checkDefTimeout(el.Value.(Driver).OnStop)
	}
	// Pipeline
	for _, pipeline := range slf.pipelines {
		slf.checkDefTimeout(pipeline.OnStop)
	}
	// Plugin
	for el := slf.plugins.Front(); el != nil; el = el.Next() {
		slf.checkDefTimeout(el.Value.(Plugin).OnStop)
	}
}

func (slf *GeckoEngine) checkDefTimeout(act func(GeckoScoped)) {
	slf.scoped.CheckTimeout(DefaultLifeCycleTimeout, func() {
		act(slf.scoped)
	})
}
