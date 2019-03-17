package gecko

import "github.com/parkingwang/go-conf"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// 初始化接口提供一个初始化组件的Interface。
type NeedInit interface {
	// 指定参数初始化组件
	OnInit(config *cfg.Config, context Context)
}

// 初始化接口。允许实现组件通过Struct结构体来包装config参数
type NeedStructInit interface {
	// 初始化。其中config参数为GetConfigStruct返回的结构体。
	OnInit(config interface{}, context Context)
	// 指定参数结构体类型。内部字段必须带toml tag
	GetConfigStruct() interface{}
}

// 生命周期接口
type LifeCycle interface {
	// 启动
	OnStart(ctx Context)
	// 停止
	OnStop(ctx Context)
}

// 组件创建工厂函数
type ComponentFactory func() interface{}

// 编码解码工厂函数
type CodecFactory func() interface{}

// Plugin
type Plugin interface {
	LifeCycle
}

// 生命周期Hook
type HookFunc func(pipeline *Pipeline)
