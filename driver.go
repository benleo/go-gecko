package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Driver interface {
	Bundle
	NeedTopicFilter
	// 处理外部请求，返回响应结果。
	// 在Driver内部，可以通过 ProtoPipelineSelector 来获取需要的设备管理器，从而控制设备。
	Handle(ctx GeckoContext, selector ProtoPipelineSelector, scoped GeckoScoped) error
}
