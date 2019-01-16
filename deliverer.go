package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// InputDeliverer，用于InputDevice发起输入事件，获取其返回的响应数据。
type InputDeliverer func(topic string, frame PacketFrame) (PacketFrame, error)

// 只发送事件通知，忽略响应结果
func (d InputDeliverer) Broadcast(topic string, frame PacketFrame) error {
	_, err := d(topic, frame)
	return err
}

// 发送事件请求，并获取结果
func (d InputDeliverer) Execute(topic string, frame PacketFrame) (PacketFrame, error) {
	return d(topic, frame)
}

// OutputDeliverer，用于OutputDevice输出事件，它负责向指定地址的设备，发送数据
type OutputDeliverer func(unionOrGroupAddress string, isUnionAddress bool, data PacketMap) (PacketMap, error)

// 执行指定Union地址的设备
func (fun OutputDeliverer) Execute(deviceUnionAddress string, frame PacketMap) (PacketMap, error) {
	return fun(deviceUnionAddress, true, frame)
}

// 广播给Group地址的全部设备
func (fun OutputDeliverer) Broadcast(deviceGroupAddress string, frame PacketMap) {
	fun(deviceGroupAddress, false, frame)
}
