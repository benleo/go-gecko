package gecko

import (
	"container/list"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko/x"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 负责对Engine组件的注册管理
type Registration struct {
	// 组件管理
	pipelines    map[string]ProtoPipeline
	plugins      *list.List
	interceptors *list.List
	drivers      *list.List
	triggers     *list.List
	// Hooks
	startBeforeHooks *list.List
	startAfterHooks  *list.List
	stopBeforeHooks  *list.List
	stopAfterHooks   *list.List
	// 组件创建工厂函数
	bundleFactories map[string]BundleFactory
}

func prepare() *Registration {
	re := new(Registration)
	re.plugins = list.New()
	re.pipelines = make(map[string]ProtoPipeline)
	re.interceptors = list.New()
	re.drivers = list.New()
	re.triggers = list.New()
	re.startBeforeHooks = list.New()
	re.startAfterHooks = list.New()
	re.stopBeforeHooks = list.New()
	re.stopAfterHooks = list.New()
	re.bundleFactories = make(map[string]BundleFactory)
	return re
}

// 添加一个设备对象。
// 设备对象的地址必须唯一。如果设备地址重复，会抛出异常。
func (re *Registration) AddVirtualDevice(device VirtualDevice) {
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
func (re *Registration) AddPlugin(plugin Plugin) {
	re.plugins.PushBack(plugin)
}

// 添加Interceptor
func (re *Registration) AddInterceptor(interceptor Interceptor) {
	re.interceptors.PushBack(interceptor)
}

// 添加Driver
func (re *Registration) AddDriver(driver Driver) {
	re.drivers.PushBack(driver)
}

// 添加Trigger
func (re *Registration) AddTrigger(trigger Trigger) {
	re.triggers.PushBack(trigger)
}

func (re *Registration) AddStartBeforeHook(hook HookFunc) {
	re.startBeforeHooks.PushBack(hook)
}

func (re *Registration) AddStartAfterHook(hook HookFunc) {
	re.startAfterHooks.PushBack(hook)
}

func (re *Registration) AddStopBeforeHook(hook HookFunc) {
	re.stopBeforeHooks.PushBack(hook)
}

func (re *Registration) AddStopAfterHook(hook HookFunc) {
	re.startAfterHooks.PushBack(hook)
}

// 添加Pipeline。
// 如果已存在相同协议的Pipeline，会抛出异常
func (re *Registration) AddProtoPipeline(pipeline ProtoPipeline) {
	proto := pipeline.GetProtoName()
	if _, ok := re.pipelines[proto]; !ok {
		re.pipelines[proto] = pipeline
	} else {
		re.withTag(log.Panic).Msgf("已存在相同协议的Pipeline: %s", proto)
	}
}

func (re *Registration) showBundles() {
	re.withTag(log.Info).Msgf("已加载 Interceptors: %d", re.interceptors.Len())
	x.ForEach(re.interceptors, func(it interface{}) {
		re.withTag(log.Info).Msgf("  -Interceptor: " + x.SimpleClassName(it))
	})

	devices := make([]VirtualDevice, 0)
	for _, pi := range re.pipelines {
		devices = append(devices, pi.GetDevices()...)
	}
	re.withTag(log.Info).Msgf("已加载 Devices: %d", len(devices))
	for _, it := range devices {
		re.withTag(log.Info).Msgf("  -Device: " + x.SimpleClassName(it))
	}

	re.withTag(log.Info).Msgf("已加载 Drivers: %d", re.drivers.Len())
	x.ForEach(re.drivers, func(it interface{}) {
		re.withTag(log.Info).Msgf("  -Driver: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载 Triggers: %d", re.triggers.Len())
	x.ForEach(re.triggers, func(it interface{}) {
		re.withTag(log.Info).Msgf("  -Trigger: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载 Plugins: %d", re.plugins.Len())
	x.ForEach(re.plugins, func(it interface{}) {
		re.withTag(log.Info).Msgf("  -Plugin: " + x.SimpleClassName(it))
	})
}

func (re *Registration) RegisterBundleFactory(typeName string, factory BundleFactory) {
	if _, ok := re.bundleFactories[typeName]; ok {
		re.withTag(log.Warn).Msgf("组件类型[%s]工厂函数被覆盖注册： %s", typeName, x.SimpleClassName(factory))
	}
	re.bundleFactories[typeName] = factory
}

// 查找指定类型的
func (re *Registration) findFactory(typeName string) (BundleFactory, bool) {
	if f, ok := re.bundleFactories[typeName]; ok {
		return f, true
	} else {
		return nil, false
	}
}

func (re *Registration) registerBundles(configs conf.Map,
	initAct func(bundle Initialize, args map[string]interface{})) {

	for typeName, item := range configs {
		asMap, ok := item.(map[string]interface{})
		if !ok {
			re.withTag(log.Panic).Msgf("组件配置信息类型错误: %s", typeName)
		}
		config := conf.MapToMap(asMap)
		if config.MustBool("disable") {
			re.withTag(log.Warn).Msgf("组件[%s]在配置中禁用", typeName)
			continue
		}

		// 配置选项中，指定 type 字段为类型名称
		if defineTypeName := config.MustString("type"); "" != defineTypeName {
			typeName = defineTypeName
		}

		factory, ok := re.findFactory(typeName)
		if !ok {
			re.withTag(log.Panic).Msgf("组件类型[%s]没有注册对应的工厂函数", typeName)
			return
		}
		// 根据类型注册
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
			return
		}

		// 需要Topic过滤
		if tf, ok := bundle.(NeedTopicFilter); ok {
			if topics, err := config.MustStringArray("topics"); nil != err {
				re.withTag(log.Panic).Msgf("配置项中[topics]必须是字符串数组： %s", typeName)
			} else {
				tf.SetTopics(topics)
			}
		}

		// 组件初始化。由外部函数处理，减少不必要的依赖
		if init, ok := bundle.(Initialize); ok {
			initAct(init, map[string]interface{}(config.MustMap("InitArgs")))
		}
	}
}

func (re *Registration) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Engine")
}
