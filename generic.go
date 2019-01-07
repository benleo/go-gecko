package gecko

import (
	"container/list"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type GenericEngine struct {
	plugins      list.List
	pipelines    map[string]DevicePipeline
	interceptors list.List
	drivers      list.List
	triggers     list.List
}

// 添加一个设备对象。
// 设备对象的地址必须唯一。如果设备地址重复，会抛出异常。
func (slf *GenericEngine) AddVirtualDevice(device VirtualDevice) {
	pn := device.GetProtoName()
	if pipeline, ok := slf.pipelines[pn]; ok {
		if err := pipeline.AddDevice(device); nil != err {
			slf.withTag(log.Panic).Msgf("设备地址重复: %s", device.GetUnionAddress())
		}
	} else {
		slf.withTag(log.Panic).Msgf("未找到对应协议的Pipeline: %s", pn)
	}
}

// 添加Plugin
func (slf *GenericEngine) AddPlugin(plugin Plugin) {
	slf.plugins.PushBack(plugin)
}

// 添加Interceptor
func (slf *GenericEngine) AddInterceptor(interceptor Interceptor) {
	slf.interceptors.PushBack(interceptor)
}

// 添加Driver
func (slf *GenericEngine) AddDriver(driver Driver) {
	slf.drivers.PushBack(driver)
}

// 添加Pipeline。
// 如果已存在相同协议的Pipeline，会抛出异常
func (slf *GenericEngine) AddPipeline(pipeline DevicePipeline) {
	pn := pipeline.GetProtoName()
	if _, ok := slf.pipelines[pn]; !ok {
		slf.pipelines[pn] = pipeline
	} else {
		slf.withTag(log.Panic).Msgf("已存在相同协议的Pipeline: %s", pn)
	}
}

func (slf *GenericEngine) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "GenericEngine")
}
