package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// Device是一个硬件设备的表示符号。
// 它可以接收由其它组件派发到此设备的事件，做出操作后，返回一个响应事件。
type VirtualDevice interface {
	NeedInitialize
	Bundle

	// 属组地址
	setGroupAddress(addr string)
	GetGroupAddress() string
	// 设置设备物理地址
	setPhyAddress(addr string)
	GetPhyAddress() string

	// 获取设备地址，由 /{GroupAddress}/{PhysicalAddress} 组成。
	GetUnionAddress() string

	// 设备名称
	setDisplayName(name string)
	GetDisplayName() string

	// 返回当前设备支持的通讯协议名称
	GetProtoName() string

	// 设备对象接收控制事件；经设备驱动处理后，返回处理结果事件；
	Process(frame *EventFrame, scoped GeckoScoped) (*EventFrame, error)
}

//

// 虚拟设备对象抽象实现
type AbcVirtualDevice struct {
	VirtualDevice
	displayName  string
	groupAddress string
	phyAddress   string
}

func (avd *AbcVirtualDevice) setDisplayName(name string) {
	avd.displayName = name
}

func (avd *AbcVirtualDevice) GetDisplayName() string {
	return avd.displayName
}

func (avd *AbcVirtualDevice) setGroupAddress(addr string) {
	avd.groupAddress = addr
}

func (avd *AbcVirtualDevice) GetGroupAddress() string {
	return avd.groupAddress
}

func (avd *AbcVirtualDevice) setPhyAddress(addr string) {
	avd.phyAddress = addr
}

func (avd *AbcVirtualDevice) GetPhyAddress() string {
	return avd.phyAddress
}

func (avd *AbcVirtualDevice) GetUnionAddress() string {
	return "/" + avd.groupAddress + "/" + avd.phyAddress
}
