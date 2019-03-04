package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// InputDeliverer，用于Input设备发起输入事件数据，并获取系统处理结果数据；
//// @param topic 输入事件的Topic；
//// @param frame 输入事件数据包；
type InputDeliverer func(topic string, frame FramePacket) (FramePacket, error)

func (fn InputDeliverer) Deliver(topic string, frame FramePacket) (FramePacket, error) {
	return fn(topic, frame)
}

////

// OutputDeliverer，用于向Output设备发送指令请求，并返回Output设备的处理结果。
// @param address 设备UUID地址，或者Group地址；与 broadcast 参数一起声明；
// @param broadcast 是否为广播模式；如果为True，则address作为设备Group来查找设备列表；
// @param data 指令数据包；
type OutputDeliverer func(address string, broadcast bool, data JSONPacket) (JSONPacket, error)

// @see OutputDeliverer
func (fn OutputDeliverer) Deliver(address string, broadcast bool, data JSONPacket) (JSONPacket, error) {
	return fn(address, broadcast, data)
}

// Execute 用于向指定UUID地址的Output设备发送指令请求，并返回Output设备的处理结果；
func (fn OutputDeliverer) Execute(address string, frame JSONPacket) (JSONPacket, error) {
	return fn(address, false, frame)
}

// Broadcast 用于向指定设备Group地址组发送广播指令请求，
// 并返回接收广播消息的所有设备的处理结果Map（以设备UUID为Key）；
func (fn OutputDeliverer) Broadcast(address string, frame JSONPacket) (JSONPacket, error) {
	return fn(address, true, frame)
}
