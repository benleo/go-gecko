package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Driver interface {
	Bundle

	// 设置当前Driver可处理的Topic列表
	SetTopics(topics []string)

	// 返回当前Driver可处理的Topic列表
	GetTopics() []string

	// 处理外部请求，返回响应结果。
	// 在Driver内部，可以通过 ProtoPipelineSelector 来获取需要的设备管理器，从而控制设备。
	Handle(ctx GeckoContext, selector ProtoPipelineSelector, scoped GeckoScoped) error
}
