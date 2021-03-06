package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// VirtualDevice是对硬件的抽象；
type VirtualDevice interface {
	// 内部函数
	setUuid(uuid string)
	setDecoder(decoder Decoder)
	setEncoder(encoder Encoder)
	// 公开可访问函数
	GetUuid() string
	GetDecoder() Decoder
	GetEncoder() Encoder
	NeedName
}

// 初始化接口提供一个初始化组件的Interface。
type Initial interface {
	// 通用类型的组件初始化函数。
	OnInit(config map[string]interface{}, context Context)
}

// 组件命名
type NeedName interface {
	setName(name string)
	GetName() string
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
type Factory func() interface{}

// 编码解码工厂函数
type CodecFactory func() interface{}

// Plugin
type Plugin interface {
	LifeCycle
}

// 生命周期Hook
type HookFunc func(pipeline *Pipeline)
