package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
// Device是一个硬件设备的表示符号。
// 它可以接收由其它组件派发到此设备的事件，做出操作后，返回一个响应事件。

type VirtualDevice interface {
	Bundle

	// 取设备名称
	SetDisplayName(name string)

	// 返回设备属组地址
	SetGroupAddress(addr string)

	// 设置设备物理地址
	SetPhyAddress(addr string)

	// 获取设备地址，由 /{GroupAddress}/{PhysicalAddress} 组成。
	GetUnionAddress() string

	// 返回当前设备支持的通讯协议名称
	GetProtoName() string

	// 设备对象接收控制事件；经设备驱动处理后，返回处理结果事件；
	Process(frame *EventFrame, scoped GeckoScoped) (*EventFrame, error)
}
