package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// Bundle 是一个具有生命周期管理的接口
type Bundle interface {
	Initialize
	// 启动
	OnStart(scoped GeckoScoped)
	// 停止
	OnStop(scoped GeckoScoped)
}

// Plugin 用于隔离其它类型与 Bundle 的类型。
type Plugin interface {
	Bundle
}
