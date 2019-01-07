package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// Bundle 是一个融合了初始化，生命周期管理的接口
type Bundle interface {
	// 启动
	OnStart(scoped GeckoScoped)

	// 停止
	OnStop(scoped GeckoScoped)
}

// VirtualPlugin 用于隔离其它类型与 Bundle 的类型。
type Plugin interface {
	Bundle
}
