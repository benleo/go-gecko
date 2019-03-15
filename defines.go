package gecko

import "github.com/parkingwang/go-conf"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// 初始化接口提供一个初始化组件的Interface。
type Initialize interface {
	// 指定参数初始化组件
	OnInit(config *cfg.Config, context Context)
}

// Bundle 是一个具有生命周期管理的接口
type Bundle interface {
	Initialize
	// 启动
	OnStart(ctx Context)
	// 停止
	OnStop(ctx Context)
}

// 组件创建工厂函数
type BundleFactory func() interface{}

// 编码解码工厂函数
type CodecFactory func() interface{}

// Plugin 用于隔离其它类型与 Bundle 的类型。
type Plugin interface {
	Bundle
}

// 生命周期Hook
type HookFunc func(pipeline *Pipeline)
