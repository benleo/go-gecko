package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// DevicePipeline 是可以处理一类设备通讯协议的管理类。
type DevicePipeline interface {
	Bundle

	// 返回当前支持的通讯协议名称
	GetProtoName() string

	// 根据设备地址查找设备对象
	FindDeviceByAddress(unionAddress string) VirtualDevice

	// 根据Group地址，返回设备列表
	FindDevicesByGroup(group string) []VirtualDevice

	// 返回当前管理的全部设备对象
	GetDevices() []VirtualDevice

	// 添加设备对象。如果设备对象的地址重复，会返回错误。
	AddDevice(device VirtualDevice) error

	// 移除设备对象
	RemoveDevice(device VirtualDevice)
}

// 根据指定协议名，返回指定协议的Pipeline
type ProtoPipelineSelector func(proto string) DevicePipeline
