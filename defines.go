package gecko

import "github.com/parkingwang/go-conf"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// 初始化接口提供一个初始化组件的Interface。
type Initial interface {
	// 通用类型的组件初始化函数。
	OnInit(config *cfg.Config, context Context)
}

// 初始化接口。
// 允许实现组件通过Struct结构体来包装config参数
type StructuredInitial interface {
	// 指定参数类型的组件初始化函数。其中config参数为 StructuredConfig 返回的结构体。
	Init(structConfig interface{}, context Context)

	// StructuredConfig 返回一个结构体指针，指定Init函数的config参数类型。
	// 其中的结构体解析参数为组件的TOML配置文件的InitArgs字段；
	// 其中结构体公开访问字段必须带“toml”tag标签
	StructuredConfig() interface{}
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
