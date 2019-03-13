package gecko

import (
	"container/list"
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko/utils"
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

	// 获取Input设备列表，返回一个复制列表
	GetInputDevices() *list.List

	// 获取Output设备列表，返回一个复制列表
	GetOutputDevices() *list.List

	// 获取Interceptor设备列表，返回一个复制列表
	GetInterceptors() *list.List

	// 获取Driver列表，返回一个复制列表
	GetDrivers() *list.List

	// 获取插件列表，返回一个复制列表
	GetPlugins() *list.List

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
	cfgGeckos       *cfg.Config
	cfgGlobals      *cfg.Config
	cfgInterceptors *cfg.Config
	cfgDrivers      *cfg.Config
	cfgOutputs      *cfg.Config
	cfgInputs       *cfg.Config
	cfgInputsLogic  *cfg.Config
	cfgPlugins      *cfg.Config
	scopedKV        map[interface{}]interface{}
	plugins         *list.List
	interceptors    *list.List
	drivers         *list.List
	outputs         *list.List
	inputs          *list.List
}

func (c *_GeckoContext) gecko() *cfg.Config {
	return c.cfgGeckos
}

func (c *_GeckoContext) workerId() int64 {
	return c.cfgGeckos.GetInt64OrDefault("workerId", 0)
}

func (c *_GeckoContext) Version() string {
	return Version
}

// 获取Input设备列表
func (c *_GeckoContext) GetInputDevices() *list.List {
	return copyList(c.inputs)
}

// 获取Output设备列表
func (c *_GeckoContext) GetOutputDevices() *list.List {
	return copyList(c.outputs)
}

// 获取Interceptor设备列表
func (c *_GeckoContext) GetInterceptors() *list.List {
	return copyList(c.interceptors)
}

// 获取Driver列表
func (c *_GeckoContext) GetDrivers() *list.List {
	return copyList(c.drivers)
}

// 获取插件列表
func (c *_GeckoContext) GetPlugins() *list.List {
	return copyList(c.plugins)
}

func (c *_GeckoContext) PutMagic(key interface{}, value interface{}) {
	c.PutScoped(key, value)
}

func (c *_GeckoContext) PutScoped(key interface{}, value interface{}) {
	if _, ok := c.scopedKV[key]; ok {
		ZapSugarLogger.Panicw("ScopedKey 不可重复，Key已存在", "key", key)
	}
	c.scopedKV[key] = value
}

func (c *_GeckoContext) GetMagic(key interface{}) interface{} {
	return c.GetScoped(key)
}

func (c *_GeckoContext) GetScoped(key interface{}) interface{} {
	return c.scopedKV[key]
}

func (c *_GeckoContext) CheckTimeout(msg string, timeout time.Duration, action func()) {
	t := time.AfterFunc(timeout, func() {
		ZapSugarLogger.Errorw("指令执行时间太长", "action", msg, "timeout", timeout.String())
	})
	defer t.Stop()
	action()
}

func (c *_GeckoContext) GlobalConfig() *cfg.Config {
	return c.cfgGlobals
}

func (c *_GeckoContext) Domain() string {
	return c.cfgGeckos.MustString("domain")
}

func (c *_GeckoContext) NodeId() string {
	return c.cfgGeckos.MustString("nodeId")
}

func (c *_GeckoContext) IsVerboseEnabled() bool {
	return c.cfgGlobals.MustBool("loggingVerbose")
}

func (c *_GeckoContext) IsFailFastEnabled() bool {
	return c.cfgGlobals.GetBoolOrDefault("failFastEnable", false)
}

func (c *_GeckoContext) OnIfLogV(fun func()) {
	if c.IsVerboseEnabled() {
		fun()
	}
}

func (c *_GeckoContext) OnIfFailFast(fun func()) {
	if c.IsFailFastEnabled() {
		fun()
	}
}

func copyList(src *list.List) *list.List {
	out := new(list.List)
	utils.ForEach(src, func(it interface{}) {
		out.PushBack(it)
	})
	return out
}
