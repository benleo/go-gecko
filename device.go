package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 硬件抽象，提供通讯地址和命名接口，以及支持的通讯协议
type VirtualDevice interface {
	Bundle
	// 属组地址
	setGroupAddress(addr string)
	GetGroupAddress() string
	// 设置设备私有地址
	setPrivateAddress(addr string)
	GetPrivateAddress() string
	// 获取设备地址，由 /{GroupAddress}/{PrivateAddress} 组成。
	GetUnionAddress() string
	// 设备名称
	setDisplayName(name string)
	GetDisplayName() string
	// 返回当前设备支持的通讯协议名称
	GetProtoName() string
}

//// InteractiveDevice - 可交互设备

// InteractiveDevice 是可交互的硬件的设备。
// 它可以接收派发到此设备的事件，做出操作后，返回一个响应事件。
type InteractiveDevice interface {
	VirtualDevice
	// 设备对象接收控制事件；经设备驱动处理后，返回处理结果事件；
	Process(frame *PacketFrame, ctx Context) (*PacketFrame, error)
}

//// Input设备

// Input设备是表示向系统输入数据的设备；它只输出数据，不接受其它控制信号；
type InputDevice interface {
	VirtualDevice

	// 监听设备的输入数据。如果设备发生错误，返回错误信息。
	Subscribe(ctx Context, onReceived OnReceivedListener) error
}

// Input设备监听回调函数
type OnReceivedListener func(frame *PacketFrame, err error)

//// Output设备

// Output设备表示系统向其输出数据的设备；它只负责接收数据，并向特定目标设备输出，不对处理结果做出反馈；
type OutputDevice interface {
	VirtualDevice

	// 发送输出数据。如果设备发生错误，返回错误信息；
	Publish(ctx Context, frame *PacketFrame) error
}
