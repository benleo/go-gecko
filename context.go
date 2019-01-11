package gecko

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

var GeckoVersion = "G1-0.0.1"

// Context 提供一些全局性质的函数
type Context interface {
	// 返回当前版本
	Version() string

	// 检查操作是否超时
	CheckTimeout(msg string, timeout time.Duration, action func())

	// 返回服务节点Domain
	Domain() string

	// 返回服务节点NodeId
	NodeId() string

	// 返回Global配置项
	GlobalConfig() map[string]interface{}

	// Globals中是否开启了 loggingVerbose 标记位
	IsVerboseEnabled() bool

	// 如果Globals设置了Verbose标记，则显示详细日志。
	LogIfV(fun func())

	// 向Context添加Key-Value数据
	// 添加的Key不可重复
	PutMagic(key interface{}, value interface{})

	// 读取Context的KeyValue数据
	GetMagic(key interface{}) interface{}

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
	confGecko        map[string]interface{}
	confGlobals      map[string]interface{}
	confPipelines    map[string]interface{}
	confInterceptors map[string]interface{}
	confDrivers      map[string]interface{}
	confDevices      map[string]interface{}
	confTriggers     map[string]interface{}
	confPlugins      map[string]interface{}
	magicKV          map[interface{}]interface{}
}

func (ci *contextImpl) gecko() map[string]interface{} {
	return ci.confGecko
}

func (ci *contextImpl) workerId() int64 {
	return conf.MapToMap(ci.confGecko).GetInt64OrDefault("workerId", 0)
}

func (ci *contextImpl) failFastEnabled() bool {
	return conf.MapToMap(ci.confGlobals).GetBoolOrDefault("failFastEnabled", false)
}

func (ci *contextImpl) Version() string {
	return GeckoVersion
}

func (ci *contextImpl) PutMagic(key interface{}, value interface{}) {
	if _, ok := ci.magicKV[key]; ok {
		ci.withTag(log.Panic).Msgf("MagicKey不可重复，Key已存在： %v", key)
	}
	ci.magicKV[key] = value
}

func (ci *contextImpl) GetMagic(key interface{}) interface{} {
	return ci.magicKV[key]
}

func (ci *contextImpl) CheckTimeout(msg string, timeout time.Duration, action func()) {
	t := time.AfterFunc(timeout, func() {
		ci.withTag(log.Debug).Msgf("Action [%s] takes too long: %s", msg, timeout.String())
	})
	defer t.Stop()
	action()
}

func (ci *contextImpl) GlobalConfig() map[string]interface{} {
	return ci.confGlobals
}

func (ci *contextImpl) Domain() string {
	return conf.MapToMap(ci.gecko()).MustString("domain")
}

func (ci *contextImpl) NodeId() string {
	return conf.MapToMap(ci.gecko()).MustString("nodeId")
}

func (ci *contextImpl) IsVerboseEnabled() bool {
	return conf.MapToMap(ci.GlobalConfig()).MustBool("loggingVerbose")
}

func (ci *contextImpl) LogIfV(fun func()) {
	if ci.IsVerboseEnabled() {
		fun()
	}
}

func (ci *contextImpl) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return log.Debug().Str("tag", "Context")
}
