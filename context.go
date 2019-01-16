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

var GeckoVersion = "G1-0.2.0"

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
	GlobalConfig() *conf.ImmutableMap

	// Globals中是否开启了 loggingVerbose 标记位
	IsVerboseEnabled() bool

	// 返回是否在Globals中配置了快速失败标记位 failFastEnabled 字段设置
	IsFailFastEnabled() bool

	// 如果Globals设置了Verbose标记，则调用此函数
	OnIfLogV(fun func())

	// 如果启用了FailFast标记则调用此函数
	OnIfFailFast(fun func())

	// 向Context添加Key-Value数据。注意：添加的Key不可重复
	PutMagic(key interface{}, value interface{})

	// 读取Context的KeyValue数据
	GetMagic(key interface{}) interface{}

	////

	// 返回Gecko的配置
	gecko() *conf.ImmutableMap

	// 返回分布式ID生成器的WorkerId
	workerId() int64
}

///

type _GeckoContext struct {
	geckos       *conf.ImmutableMap
	globals      *conf.ImmutableMap
	interceptors *conf.ImmutableMap
	drivers      *conf.ImmutableMap
	outputs      *conf.ImmutableMap
	inputs       *conf.ImmutableMap
	plugins      *conf.ImmutableMap
	magicKV      map[interface{}]interface{}
}

func (ci *_GeckoContext) gecko() *conf.ImmutableMap {
	return ci.geckos
}

func (ci *_GeckoContext) workerId() int64 {
	return ci.geckos.GetInt64OrDefault("workerId", 0)
}

func (ci *_GeckoContext) Version() string {
	return GeckoVersion
}

func (ci *_GeckoContext) PutMagic(key interface{}, value interface{}) {
	if _, ok := ci.magicKV[key]; ok {
		ci.withTag(log.Panic).Msgf("MagicKey不可重复，Key已存在： %v", key)
	}
	ci.magicKV[key] = value
}

func (ci *_GeckoContext) GetMagic(key interface{}) interface{} {
	return ci.magicKV[key]
}

func (ci *_GeckoContext) CheckTimeout(msg string, timeout time.Duration, action func()) {
	t := time.AfterFunc(timeout, func() {
		ci.withTag(log.Debug).Msgf("Action [%s] takes too long: %s", msg, timeout.String())
	})
	defer t.Stop()
	action()
}

func (ci *_GeckoContext) GlobalConfig() *conf.ImmutableMap {
	return ci.globals
}

func (ci *_GeckoContext) Domain() string {
	return ci.geckos.MustString("domain")
}

func (ci *_GeckoContext) NodeId() string {
	return ci.geckos.MustString("nodeId")
}

func (ci *_GeckoContext) IsVerboseEnabled() bool {
	return ci.globals.MustBool("loggingVerbose")
}

func (ci *_GeckoContext) IsFailFastEnabled() bool {
	return ci.globals.GetBoolOrDefault("failFastEnable", false)
}

func (ci *_GeckoContext) OnIfLogV(fun func()) {
	if ci.IsVerboseEnabled() {
		fun()
	}
}

func (ci *_GeckoContext) OnIfFailFast(fun func()) {
	if ci.IsFailFastEnabled() {
		fun()
	}
}

func (ci *_GeckoContext) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return log.Debug().Str("tag", "Context")
}
