package gecko

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Scoped提供一些全局性质的函数
type GeckoScoped interface {
	// 返回当前版本
	Version() string

	// 检查操作是否超时
	CheckTimeout(timeout time.Duration, action func())

	// 返回服务节点Domain
	Domain() string

	// 返回服务节点NodeId
	NodeId() string

	// 返回Global配置项
	Globals() map[string]interface{}

	// Globals中是否开启了 loggingVerbose 标记位
	IsVerboseEnabled() bool

	// 如果Globals设置了Verbose标记，则显示详细日志。
	LogIfV(fun func())

	// 返回Gecko的配置
	gecko() map[string]interface{}
}

///

type abcGeckoScoped struct {
	GeckoScoped

	geckoConf        map[string]interface{}
	globalsConf      map[string]interface{}
	pipelinesConf    map[string]interface{}
	interceptorsConf map[string]interface{}
	driversConf      map[string]interface{}
	devicesConf      map[string]interface{}
	triggersConf     map[string]interface{}
	pluginsConf      map[string]interface{}
}

func (gs *abcGeckoScoped) gecko() map[string]interface{} {
	return gs.geckoConf
}

func (gs *abcGeckoScoped) Version() string {
	return "G1-1.0.0"
}

func (gs *abcGeckoScoped) CheckTimeout(timeout time.Duration, action func()) {
	t := time.AfterFunc(timeout, func() {
		log.Debug().Str("tag", "GeckoScoped").Msgf("Action takes toooo long: %s", timeout.String())
	})
	defer t.Stop()
	action()
}

func (gs *abcGeckoScoped) Domain() string {
	return conf.MapToMap(gs.geckoConf).MustString("domain")
}

func (gs *abcGeckoScoped) NodeId() string {
	return conf.MapToMap(gs.geckoConf).MustString("nodeId")
}

func (gs *abcGeckoScoped) IsVerboseEnabled() bool {
	return conf.MapToMap(gs.geckoConf).MustBool("loggingVerbose")
}

func (gs *abcGeckoScoped) LogIfV(fun func()) {
	if gs.IsVerboseEnabled() {
		fun()
	}
}

func newAbcGeckoScoped(config map[string]interface{}) GeckoScoped {
	mapConf := conf.MapToMap(config)
	scoped := &abcGeckoScoped{
		geckoConf:        mapConf.MustMap("GECKO"),
		globalsConf:      mapConf.MustMap("GLOBALS"),
		pipelinesConf:    mapConf.MustMap("PIPELINES"),
		interceptorsConf: mapConf.MustMap("INTERCEPTORS"),
		driversConf:      mapConf.MustMap("DRIVERS"),
		devicesConf:      mapConf.MustMap("DEVICES"),
		triggersConf:     mapConf.MustMap("TRIGGERS"),
		pluginsConf:      mapConf.MustMap("PLUGINS"),
	}
	return scoped
}
