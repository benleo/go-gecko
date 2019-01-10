package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Driver interface {
	Bundle
	NeedTopicFilter
	// 处理外部请求，返回响应结果。
	// 在Driver内部，可以通过 ProtoPipelineSelector 来获取需要的设备管理器，从而控制设备。
	Handle(session Session, selector ProtoPipelineSelector, ctx Context) error
}

////

// Driver抽象实现
type AbcDriver struct {
	Driver
	topics []*TopicExpr
}

func (ad *AbcDriver) setTopics(topics []string) {
	if nil == ad.topics {
		ad.topics = make([]*TopicExpr, 0, len(topics))
	}
	for _, t := range topics {
		ad.topics = append(ad.topics, newTopicExpr(t))
	}
}

func (ad *AbcDriver) GetTopicExpr() []*TopicExpr {
	return ad.topics
}
