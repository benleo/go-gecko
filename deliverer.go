package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// InputDeliverer，用于Input设备发起输入事件数据，并获取系统处理结果数据；
// @param topic 输入事件的Topic；
// @param frame 输入事件数据包；
type InputDeliverer func(topic string, frame FramePacket) (FramePacket, error)

func (fn InputDeliverer) Deliver(topic string, frame FramePacket) (FramePacket, error) {
	return fn(topic, frame)
}

////

// OutputDeliverer，用于向Output设备发送指令请求，并返回Output设备的处理结果。
// @param uuid 设备UUID地址
// @param data 指令数据包；
type OutputDeliverer func(uuid string, data JSONPacket) (JSONPacket, error)

// @see OutputDeliverer
func (fn OutputDeliverer) Deliver(uuid string, data JSONPacket) (JSONPacket, error) {
	return fn(uuid, data)
}
