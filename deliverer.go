package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// InputDeliverer，用于为InputDevice发起输入事件的接口，并获取系统处理结果数据。
// InputDeliverer负责传递的数据，是InputDeliverer的原始数据。
// 当数据从Deliverer输入系统内部时，会由InputDevice的Decoder解析成PacketMap格式；
// 当内部系统返回处理结果数据时，会由InputDevice的Encoder编码成原始数据格式；
type InputDeliverer func(topic string, frame FramePacket) (FramePacket, error)

// 扩展函数1：发送事件请求，并获取处理结果；
func (d InputDeliverer) Execute(topic string, frame FramePacket) (FramePacket, error) {
	return d(topic, frame)
}

// 扩展函数2：发送事件通知，忽略处理结果；
func (d InputDeliverer) Broadcast(topic string, frame FramePacket) error {
	_, err := d(topic, frame)
	return err
}

// OutputDeliverer，用于向OutputDevice输出控制事件的接口，并获取设备处理结果数据；
// OutputDeliverer所传递的数据格式，是系统内部的统一数据PacketMap；
// 当内部系统向OutputDevice输出数据时，由OutputDevice的Encoder编码成设备原始格式；
// 当OutputDevice返回处理结果数据时，由其Decoder解码成内部系统格式；
type OutputDeliverer func(address string, broadcast bool, data JSONPacket) (JSONPacket, error)

// 扩展函数1：指定设备地址的设备，获取处理结果；
func (fun OutputDeliverer) Execute(address string, frame JSONPacket) (JSONPacket, error) {
	return fun(address, false, frame)
}

// 扩展函数2：通过Group地址，广播给相同Group的设备列表，忽略设备处理结果；
func (fun OutputDeliverer) Broadcast(group string, frame JSONPacket) error {
	_, err := fun(group, true, frame)
	return err
}
