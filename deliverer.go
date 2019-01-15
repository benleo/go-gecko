package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Deliverer用于发起事件并获取其返回的响应数据。
type Deliverer func(topic string, frame PacketFrame) (PacketFrame, error)

// 只通知，不接收响应结果
func (d Deliverer) DeliverPublish(topic string, frame PacketFrame) error {
	_, err := d(topic, frame)
	return err
}

// 发送事件，并获取结果
func (d Deliverer) DeliverInvoke(topic string, event PacketFrame) (PacketFrame, error) {
	return d(topic, event)
}
