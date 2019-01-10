package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com

// Device是一个硬件设备的表示符号。
// 它可以接收由其它组件派发到此设备的事件，做出操作后，返回一个响应事件。
type VirtualDevice interface {
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
	Process(frame *PacketFrame, ctx Context) (*PacketFrame, error)
}

//

// 虚拟设备对象抽象实现
type AbcVirtualDevice struct {
	VirtualDevice
	displayName  string
	groupAddress string
	phyAddress   string
}

func (av *AbcVirtualDevice) setDisplayName(name string) {
	av.displayName = name
}

func (av *AbcVirtualDevice) GetDisplayName() string {
	return av.displayName
}

func (av *AbcVirtualDevice) setGroupAddress(addr string) {
	av.groupAddress = addr
}

func (av *AbcVirtualDevice) GetGroupAddress() string {
	return av.groupAddress
}

func (av *AbcVirtualDevice) setPhyAddress(addr string) {
	av.phyAddress = addr
}

func (av *AbcVirtualDevice) GetPhyAddress() string {
	return av.phyAddress
}

func (av *AbcVirtualDevice) GetUnionAddress() string {
	return "/" + av.groupAddress + "/" + av.phyAddress
}
