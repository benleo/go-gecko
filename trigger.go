package gecko

//
// Author: yoojiachen@gmail.com
//

// Trigger与Driver类似, 也同样可以驱动设备. 不同点是,Trigger只作为触发操作, 不会返回处理结果数据到Input输出.
type Trigger interface {
	NeedTopicFilter
	NeedName
	// 处理外部请求，返回响应结果。
	// 在Trigger内部，可以通过 OutputDeliverer 来控制其它设备。
	Handle(attrs Attributes, topic string, inputUUid string, inbound *MessagePacket, deliverer OutputDeliverer, ctx Context) error
}

//// Trigger抽象实现

type AbcTrigger struct {
	Trigger
	name   string
	topics []*TopicExpr
}

func (ad *AbcTrigger) setName(name string) {
	ad.name = name
}

// 获取Trigger名字
func (ad *AbcTrigger) GetName() string {
	return ad.name
}

func (ad *AbcTrigger) setTopics(topics []string) {
	for _, t := range topics {
		ad.topics = append(ad.topics, newTopicExpr(t))
	}
}

// 获取Trigger可处理的Topic列表
func (ad *AbcTrigger) GetTopicExpr() []*TopicExpr {
	return ad.topics
}

func NewAbcTrigger() *AbcTrigger {
	return &AbcTrigger{
		topics: make([]*TopicExpr, 0),
	}
}
