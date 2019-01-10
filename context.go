package gecko

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Context 提供一些全局性质的函数
type Context interface {
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

	////

	// 返回Gecko的配置
	gecko() map[string]interface{}

	// 返回分布式ID生成器的WorkerId
	workerId() int64

	// 返回是否在Globals中配置了快速失败标记位
	failFastEnabled() bool
}

///

type contextImpl struct {
	Context

	geckoConf        map[string]interface{}
	globalsConf      map[string]interface{}
	pipelinesConf    map[string]interface{}
	interceptorsConf map[string]interface{}
	driversConf      map[string]interface{}
	devicesConf      map[string]interface{}
	triggersConf     map[string]interface{}
	pluginsConf      map[string]interface{}
}

func (ci *contextImpl) gecko() map[string]interface{} {
	return ci.geckoConf
}

func (ci *contextImpl) workerId() int64 {
	return conf.MapToMap(ci.geckoConf).GetInt64OrDefault("workerId", 0)
}

func (ci *contextImpl) failFastEnabled() bool {
	return conf.MapToMap(ci.globalsConf).GetBoolOrDefault("failFastEnabled", false)
}

func (ci *contextImpl) Version() string {
	return "G1-1.0.0"
}

func (ci *contextImpl) CheckTimeout(timeout time.Duration, action func()) {
	t := time.AfterFunc(timeout, func() {
		log.Debug().Str("tag", "Context").Msgf("Action takes toooo long: %s", timeout.String())
	})
	defer t.Stop()
	action()
}

func (ci *contextImpl) Globals() map[string]interface{} {
	return ci.globalsConf
}

func (ci *contextImpl) Domain() string {
	return conf.MapToMap(ci.gecko()).MustString("domain")
}

func (ci *contextImpl) NodeId() string {
	return conf.MapToMap(ci.gecko()).MustString("nodeId")
}

func (ci *contextImpl) IsVerboseEnabled() bool {
	return conf.MapToMap(ci.Globals()).MustBool("loggingVerbose")
}

func (ci *contextImpl) LogIfV(fun func()) {
	if ci.IsVerboseEnabled() {
		fun()
	}
}
