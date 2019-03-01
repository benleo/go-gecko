package gecko

import (
	"github.com/parkingwang/go-conf"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

var Version = "G1-0.4"

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
	GlobalConfig() *cfg.Config

	// Globals中是否开启了 loggingVerbose 标记位
	IsVerboseEnabled() bool

	// 返回是否在Globals中配置了快速失败标记位 failFastEnabled 字段设置
	IsFailFastEnabled() bool

	// 如果Globals设置了Verbose标记，则调用此函数
	OnIfLogV(fun func())

	// 如果启用了FailFast标记则调用此函数
	OnIfFailFast(fun func())

	// 向Context添加Key-Value数据。注意：添加的Key不可重复
	PutScoped(key interface{}, value interface{})
	// 读取Context的KeyValue数据
	GetScoped(key interface{}) interface{}

	////

	// 返回Gecko的配置
	gecko() *cfg.Config

	// 返回分布式ID生成器的WorkerId
	workerId() int64
}

///

type _GeckoContext struct {
	geckos       *cfg.Config
	globals      *cfg.Config
	interceptors *cfg.Config
	drivers      *cfg.Config
	outputs      *cfg.Config
	inputs       *cfg.Config
	plugins      *cfg.Config
	scopedKV     map[interface{}]interface{}
}

func (ci *_GeckoContext) gecko() *cfg.Config {
	return ci.geckos
}

func (ci *_GeckoContext) workerId() int64 {
	return ci.geckos.GetInt64OrDefault("workerId", 0)
}

func (ci *_GeckoContext) Version() string {
	return Version
}

func (ci *_GeckoContext) PutMagic(key interface{}, value interface{}) {
	ci.PutScoped(key, value)
}

func (ci *_GeckoContext) PutScoped(key interface{}, value interface{}) {
	if _, ok := ci.scopedKV[key]; ok {
		zap := ZapSugarLogger()
		defer zap.Sync()
		zap.Panicw("ScopedKey 不可重复，Key已存在", "key", key)
	}
	ci.scopedKV[key] = value
}

func (ci *_GeckoContext) GetMagic(key interface{}) interface{} {
	return ci.GetScoped(key)
}

func (ci *_GeckoContext) GetScoped(key interface{}) interface{} {
	return ci.scopedKV[key]
}

func (ci *_GeckoContext) CheckTimeout(msg string, timeout time.Duration, action func()) {
	t := time.AfterFunc(timeout, func() {
		zap := ZapSugarLogger()
		defer zap.Sync()
		zap.Errorw("指令执行时间太长", "action", msg, "timeout", timeout.String())
	})
	defer t.Stop()
	action()
}

func (ci *_GeckoContext) GlobalConfig() *cfg.Config {
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
