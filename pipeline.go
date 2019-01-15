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
	FindHardwareByUnionAddress(unionAddress string) VirtualHardware

	// 根据Group地址，返回设备列表
	FindHardwareByGroupAddress(group string) []VirtualHardware

	// 返回当前管理的全部设备对象
	GetManagedHardware() []VirtualHardware

	// 添加设备对象。
	// 如果设备对象的地址重复，会返回False。
	AddHardware(hw VirtualHardware) bool

	// 移除设备对象
	RemoveHardware(unionAddress string)
}

// 根据指定协议名，返回指定协议的Pipeline
type ProtoPipelineSelector func(proto string) (ProtoPipeline, bool)

////

// ProtoPipeline抽象实现类
type AbcProtoPipeline struct {
	ProtoPipeline
	addressedHardware map[string]VirtualHardware
	rwLock            *sync.RWMutex
}

func NewAbcProtoPipeline() *AbcProtoPipeline {
	return &AbcProtoPipeline{
		addressedHardware: make(map[string]VirtualHardware),
		rwLock:            new(sync.RWMutex),
	}
}

func (ap *AbcProtoPipeline) OnStart(ctx Context) {
	for addr, dev := range ap.addressedHardware {
		ctx.CheckTimeout("Device.Start@"+addr, time.Second*3, func() {
			dev.OnStart(ctx)
		})
	}
}

func (ap *AbcProtoPipeline) OnStop(ctx Context) {
	for addr, dev := range ap.addressedHardware {
		ctx.CheckTimeout("Device.Stop@"+addr, time.Second*3, func() {
			dev.OnStop(ctx)
		})
	}
}

func (ap *AbcProtoPipeline) FindHardwareByUnionAddress(unionAddress string) VirtualHardware {
	ap.rwLock.RLock()
	defer ap.rwLock.RUnlock()
	return ap.addressedHardware[unionAddress]
}

func (ap *AbcProtoPipeline) FindHardwareByGroupAddress(groupAddress string) []VirtualHardware {
	ap.rwLock.RLock()
	defer ap.rwLock.RUnlock()
	out := make([]VirtualHardware, 0)
	for _, vd := range ap.addressedHardware {
		if groupAddress == vd.GetGroupAddress() {
			out = append(out, vd)
		}
	}
	return out
}

func (ap *AbcProtoPipeline) GetManagedHardware() []VirtualHardware {
	out := make([]VirtualHardware, 0, len(ap.addressedHardware))
	ap.rwLock.RLock()
	defer ap.rwLock.RUnlock()
	for _, vd := range ap.addressedHardware {
		out = append(out, vd)
	}
	return out
}

func (ap *AbcProtoPipeline) AddHardware(hw VirtualHardware) bool {
	ap.rwLock.Lock()
	defer ap.rwLock.Unlock()
	addr := hw.GetUnionAddress()
	if _, ok := ap.addressedHardware[addr]; ok {
		return false
	} else {
		ap.addressedHardware[addr] = hw
		return true
	}
}

func (ap *AbcProtoPipeline) RemoveHardware(unionAddress string) {
	ap.rwLock.Lock()
	defer ap.rwLock.Unlock()
	delete(ap.addressedHardware, unionAddress)
}
