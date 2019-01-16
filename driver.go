package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 用户驱动
type Driver interface {
	Bundle
	NeedTopicFilter
	// 处理外部请求，返回响应结果。
	// 在Driver内部，可以通过 OutputExecutor 来获取需要的设备管理器，从而控制设备。
	Handle(session Session, executor OutputExecutor, ctx Context) error
}

////

// 输出设备执行器，它负责向指定地址的设备，发送数据
type OutputExecutor func(unionOrGroupAddress string, isUnionAddress bool, frame PacketFrame) (PacketFrame, error)

// 执行指定Union地址的设备
func (fun OutputExecutor) Execute(deviceUnionAddress string, frame PacketFrame) (PacketFrame, error) {
	return fun(deviceUnionAddress, true, frame)
}

// 广播给Group地址的全部设备
func (fun OutputExecutor) Broadcast(deviceGroupAddress string, frame PacketFrame) {
	fun(deviceGroupAddress, false, frame)
}

////

// Driver抽象实现
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
