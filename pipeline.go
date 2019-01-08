package gecko

import (
	"sync"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// ProtoPipeline 是可以处理一类设备通讯协议的管理类。
type ProtoPipeline interface {
	Bundle

	// 返回当前支持的通讯协议名称
	GetProtoName() string

	// 根据设备地址查找设备对象
	FindDeviceByAddress(unionAddress string) VirtualDevice

	// 根据Group地址，返回设备列表
	FindDevicesByGroup(group string) []VirtualDevice

	// 返回当前管理的全部设备对象
	GetDevices() []VirtualDevice

	// 添加设备对象。
	// 如果设备对象的地址重复，会返回False。
	AddDevice(device VirtualDevice) bool

	// 移除设备对象
	RemoveDevice(unionAddress string)
}

// 根据指定协议名，返回指定协议的Pipeline
type ProtoPipelineSelector func(proto string) ProtoPipeline

////

// ProtoPipeline抽象实现类
type AbcProtoPipeline struct {
	addressDevices map[string]VirtualDevice
	rwLock         *sync.RWMutex
}

func (app *AbcProtoPipeline) Init() {
	app.addressDevices = make(map[string]VirtualDevice)
	app.rwLock = new(sync.RWMutex)
}

func (app *AbcProtoPipeline) FindDeviceByAddress(unionAddress string) VirtualDevice {
	app.rwLock.RLock()
	defer app.rwLock.RUnlock()
	return app.addressDevices[unionAddress]
}

func (app *AbcProtoPipeline) FindDevicesByGroup(groupAddress string) []VirtualDevice {
	app.rwLock.RLock()
	defer app.rwLock.RUnlock()
	out := make([]VirtualDevice, 0)
	for _, vd := range app.addressDevices {
		if groupAddress == vd.GetGroupAddress() {
			out = append(out, vd)
		}
	}
	return out
}

func (app *AbcProtoPipeline) GetDevices() []VirtualDevice {
	out := make([]VirtualDevice, 0, len(app.addressDevices))
	app.rwLock.RLock()
	defer app.rwLock.RUnlock()
	for _, vd := range app.addressDevices {
		out = append(out, vd)
	}
	return out
}

func (app *AbcProtoPipeline) AddDevice(device VirtualDevice) bool {
	app.rwLock.Lock()
	defer app.rwLock.Unlock()
	addr := device.GetUnionAddress()
	if _, ok := app.addressDevices[addr]; ok {
		return false
	} else {
		app.addressDevices[addr] = device
		return true
	}
}

func (app *AbcProtoPipeline) RemoveDevice(unionAddress string) {
	app.rwLock.Lock()
	defer app.rwLock.Unlock()
	delete(app.addressDevices, unionAddress)
}
