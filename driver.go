package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Driver-用户驱动，是实现设备与设备之间联动、设备事件响应业务处理的核心组件，它通常与Interceptor一起完成某种业务功能；
// 它负责监听接收特定事件Topic的设备事件，经过内部数据库、业务方法等逻辑计算后，使用OutputDeliverer来
// 控制下一级输出设备。最典型的例子是：Driver接收到门禁刷卡ID后，驱动门锁开关设备；
type Driver interface {
	Bundle
	NeedTopicFilter
	// 处理外部请求，返回响应结果。
	// 在Driver内部，可以通过 OutputDeliverer 来控制其它设备。
	Handle(session EventSession, deliverer OutputDeliverer, ctx Context) error
}

//// Driver抽象实现

type AbcDriver struct {
	Driver
	topics []*TopicExpr
}

func (ad *AbcDriver) setTopics(topics []string) {
	for _, t := range topics {
		ad.topics = append(ad.topics, newTopicExpr(t))
	}
}

func (ad *AbcDriver) GetTopicExpr() []*TopicExpr {
	return ad.topics
}

func NewAbcDriver() *AbcDriver {
	return &AbcDriver{
		topics: make([]*TopicExpr, 0),
	}
}
