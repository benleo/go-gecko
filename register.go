package gecko

import (
	"container/list"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 负责对Engine组件的注册管理
type RegisterEngine struct {
	plugins      list.List
	pipelines    map[string]DevicePipeline
	interceptors list.List
	drivers      list.List
	triggers     list.List
}

// 添加一个设备对象。
// 设备对象的地址必须唯一。如果设备地址重复，会抛出异常。
func (slf *RegisterEngine) AddVirtualDevice(device VirtualDevice) {
	proto := device.GetProtoName()
	if pipeline, ok := slf.pipelines[proto]; ok {
		if err := pipeline.AddDevice(device); nil != err {
			slf.withTag(log.Panic).Msgf("设备地址重复: %s", device.GetUnionAddress())
		}
	} else {
		slf.withTag(log.Panic).Msgf("未找到对应协议的Pipeline: %s", proto)
	}
}

// 添加Plugin
func (slf *RegisterEngine) AddPlugin(plugin Plugin) {
	slf.plugins.PushBack(plugin)
}

// 添加Interceptor
func (slf *RegisterEngine) AddInterceptor(interceptor Interceptor) {
	slf.interceptors.PushBack(interceptor)
}

// 添加Driver
func (slf *RegisterEngine) AddDriver(driver Driver) {
	slf.drivers.PushBack(driver)
}

// 添加Pipeline。
// 如果已存在相同协议的Pipeline，会抛出异常
func (slf *RegisterEngine) AddPipeline(pipeline DevicePipeline) {
	proto := pipeline.GetProtoName()
	if _, ok := slf.pipelines[proto]; !ok {
		slf.pipelines[proto] = pipeline
	} else {
		slf.withTag(log.Panic).Msgf("已存在相同协议的Pipeline: %s", proto)
	}
}

func (slf *RegisterEngine) showBundles() {
	slf.withTag(log.Info).Msgf("已加载 Interceptors: %d", slf.interceptors.Len())

	devices := make([]VirtualDevice, 0)
	for _, pi := range slf.pipelines {
		devices = append(devices, pi.GetDevices()...)
	}
	slf.withTag(log.Info).Msgf("已加载 Devices: %d", len(devices))

	slf.withTag(log.Info).Msgf("已加载 Drivers: %d", slf.interceptors.Len())
	slf.withTag(log.Info).Msgf("已加载 Drivers: %d", slf.drivers.Len())
	slf.withTag(log.Info).Msgf("已加载 Triggers: %d", slf.triggers.Len())
	slf.withTag(log.Info).Msgf("已加载 Plugins: %d", slf.plugins.Len())
}

func (slf *RegisterEngine) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Engine")
}
