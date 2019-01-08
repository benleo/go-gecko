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
	pipelines    map[string]ProtoPipeline
	interceptors list.List
	drivers      list.List
	triggers     list.List
}

// 添加一个设备对象。
// 设备对象的地址必须唯一。如果设备地址重复，会抛出异常。
func (re *RegisterEngine) AddVirtualDevice(device VirtualDevice) {
	proto := device.GetProtoName()
	if pipeline, ok := re.pipelines[proto]; ok {
		if !pipeline.AddDevice(device) {
			re.withTag(log.Panic).Msgf("设备地址重复: %s", device.GetUnionAddress())
		}
	} else {
		re.withTag(log.Panic).Msgf("未找到对应协议的Pipeline: %s", proto)
	}
}

// 添加Plugin
func (re *RegisterEngine) AddPlugin(plugin Plugin) {
	re.plugins.PushBack(plugin)
}

// 添加Interceptor
func (re *RegisterEngine) AddInterceptor(interceptor Interceptor) {
	re.interceptors.PushBack(interceptor)
}

// 添加Driver
func (re *RegisterEngine) AddDriver(driver Driver) {
	re.drivers.PushBack(driver)
}

// 添加Pipeline。
// 如果已存在相同协议的Pipeline，会抛出异常
func (re *RegisterEngine) AddPipeline(pipeline ProtoPipeline) {
	proto := pipeline.GetProtoName()
	if _, ok := re.pipelines[proto]; !ok {
		re.pipelines[proto] = pipeline
	} else {
		re.withTag(log.Panic).Msgf("已存在相同协议的Pipeline: %s", proto)
	}
}

func (re *RegisterEngine) showBundles() {
	re.withTag(log.Info).Msgf("已加载 Interceptors: %d", re.interceptors.Len())

	devices := make([]VirtualDevice, 0)
	for _, pi := range re.pipelines {
		devices = append(devices, pi.GetDevices()...)
	}
	re.withTag(log.Info).Msgf("已加载 Devices: %d", len(devices))

	re.withTag(log.Info).Msgf("已加载 Drivers: %d", re.interceptors.Len())
	re.withTag(log.Info).Msgf("已加载 Drivers: %d", re.drivers.Len())
	re.withTag(log.Info).Msgf("已加载 Triggers: %d", re.triggers.Len())
	re.withTag(log.Info).Msgf("已加载 Plugins: %d", re.plugins.Len())
}

func (re *RegisterEngine) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Engine")
}
