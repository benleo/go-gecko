package gecko

import (
	"sync"
	"time"
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
	FindDeviceByUnionAddress(unionAddress string) VirtualDevice

	// 根据Group地址，返回设备列表
	FindDeviceByGroupAddress(group string) []VirtualDevice

	// 返回当前管理的全部设备对象
	GetManagedDevices() []VirtualDevice

	// 添加设备对象。
	// 如果设备对象的地址重复，会返回False。
	AddDevice(hw VirtualDevice) bool

	// 移除设备对象
	RemoveDevice(unionAddress string)
}

// 根据指定协议名，返回指定协议的Pipeline
type ProtoPipelineSelector func(proto string) (ProtoPipeline, bool)

////

// ProtoPipeline抽象实现类
type AbcProtoPipeline struct {
	ProtoPipeline
	addressedDevice map[string]VirtualDevice
	rwLock            *sync.RWMutex
}

func NewAbcProtoPipeline() *AbcProtoPipeline {
	return &AbcProtoPipeline{
		addressedDevice: make(map[string]VirtualDevice),
		rwLock:            new(sync.RWMutex),
	}
}

func (ap *AbcProtoPipeline) OnStart(ctx Context) {
	for addr, dev := range ap.addressedDevice {
		ctx.CheckTimeout("Device.Start@"+addr, time.Second*3, func() {
			dev.OnStart(ctx)
		})
	}
}

func (ap *AbcProtoPipeline) OnStop(ctx Context) {
	for addr, dev := range ap.addressedDevice {
		ctx.CheckTimeout("Device.Stop@"+addr, time.Second*3, func() {
			dev.OnStop(ctx)
		})
	}
}

func (ap *AbcProtoPipeline) FindDeviceByUnionAddress(unionAddress string) VirtualDevice {
	ap.rwLock.RLock()
	defer ap.rwLock.RUnlock()
	return ap.addressedDevice[unionAddress]
}

func (ap *AbcProtoPipeline) FindDeviceByGroupAddress(groupAddress string) []VirtualDevice {
	ap.rwLock.RLock()
	defer ap.rwLock.RUnlock()
	out := make([]VirtualDevice, 0)
	for _, vd := range ap.addressedDevice {
		if groupAddress == vd.GetGroupAddress() {
			out = append(out, vd)
		}
	}
	return out
}

func (ap *AbcProtoPipeline) GetManagedDevices() []VirtualDevice {
	out := make([]VirtualDevice, 0, len(ap.addressedDevice))
	ap.rwLock.RLock()
	defer ap.rwLock.RUnlock()
	for _, vd := range ap.addressedDevice {
		out = append(out, vd)
	}
	return out
}

func (ap *AbcProtoPipeline) AddDevice(hw VirtualDevice) bool {
	ap.rwLock.Lock()
	defer ap.rwLock.Unlock()
	addr := hw.GetUnionAddress()
	if _, ok := ap.addressedDevice[addr]; ok {
		return false
	} else {
		ap.addressedDevice[addr] = hw
		return true
	}
}

func (ap *AbcProtoPipeline) RemoveDevice(unionAddress string) {
	ap.rwLock.Lock()
	defer ap.rwLock.Unlock()
	delete(ap.addressedDevice, unionAddress)
}
