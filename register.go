package gecko

import (
	"container/list"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko/util"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 负责对Engine组件的注册管理
type RegisterEngine struct {
	// 组件管理
	plugins      *list.List
	pipelines    map[string]ProtoPipeline
	interceptors *list.List
	drivers      *list.List
	triggers     *list.List
	// 组件创建工厂函数
	bundleFactories map[string]BundleFactory
}

func prepare() *RegisterEngine {
	re := new(RegisterEngine)
	re.plugins = list.New()
	re.pipelines = make(map[string]ProtoPipeline)
	re.interceptors = list.New()
	re.drivers = list.New()
	re.triggers = list.New()
	re.bundleFactories = make(map[string]BundleFactory)
	return re
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

// 添加Trigger
func (re *RegisterEngine) AddTrigger(trigger Trigger) {
	re.triggers.PushBack(trigger)
}

// 添加Pipeline。
// 如果已存在相同协议的Pipeline，会抛出异常
func (re *RegisterEngine) AddProtoPipeline(pipeline ProtoPipeline) {
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

func (re *RegisterEngine) RegisterBundleFactory(typeName string, factory BundleFactory) {
	if _, ok := re.bundleFactories[typeName]; ok {
		re.withTag(log.Warn).Msgf("组件类型[%s]工厂函数被覆盖注册： %s", typeName, util.SimpleClassName(factory))
	}
	re.bundleFactories[typeName] = factory
}

// 查找指定类型的
func (re *RegisterEngine) findFactory(typeName string) (BundleFactory, bool) {
	if f, ok := re.bundleFactories[typeName]; ok {
		return f, true
	} else {
		return nil, false
	}
}

func (re *RegisterEngine) registerBundles(configs conf.Map,
	initAct func(bundle Initialize, args map[string]interface{})) {

	for typeName, item := range configs {
		asMap, ok := item.(map[string]interface{})
		if !ok {
			re.withTag(log.Panic).Msgf("组件配置信息类型错误: %s", typeName)
		}
		config := conf.MapToMap(asMap)
		if config.MustBool("disable") {
			re.withTag(log.Panic).Msgf("组件[%s]在配置中禁用", typeName)
			continue
		}

		if factory, ok := re.findFactory(typeName); !ok {
			re.withTag(log.Panic).Msgf("组件类型[%s]没有注册对应的工厂函数")
		} else {
			bundle := factory()
			switch bundle.(type) {
			case Plugin:
				re.AddPlugin(bundle.(Plugin))
			case ProtoPipeline:
				re.AddProtoPipeline(bundle.(ProtoPipeline))
			case Interceptor:
				re.AddInterceptor(bundle.(Interceptor))
			case Driver:
				re.AddDriver(bundle.(Driver))
			case VirtualDevice:
				dev := bundle.(VirtualDevice)
				if name := config.MustString("name"); "" == name {
					dev.setDisplayName(typeName)
				} else {
					dev.setDisplayName(name)
				}
				dev.setGroupAddress(config.MustString("groupAddress"))
				dev.setPhyAddress(config.MustString("physicalAddress"))
				re.AddVirtualDevice(dev)
			case Trigger:
				re.AddTrigger(bundle.(Trigger))
			default:
				re.withTag(log.Panic).Msgf("未支持的组件类型：%s", typeName)
			}
			// 需要Topic过滤
			if tf, ok := bundle.(NeedTopicFilter); ok {
				if topics, err := config.MustStringArray("topics"); nil != err {
					re.withTag(log.Panic).Msgf("配置项中[topics]必须是字符串数组： %s", typeName)
				} else {
					tf.SetTopics(topics)
				}
			}
			// 组件初始化
			if init, ok := bundle.(Initialize); ok {
				initAct(init, map[string]interface{}(config.MustMap("InitArgs")))
			}
		}
	}
}

func (re *RegisterEngine) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Engine")
}
