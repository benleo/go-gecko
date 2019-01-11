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
		re.withTag(log.Info).Msg("  - Interceptor: " + x.SimpleClassName(it))
	})

	devices := make([]VirtualDevice, 0)
	re.withTag(log.Info).Msgf("已加载 Pipelines: %d", len(re.pipelines))
	for proto, pi := range re.pipelines {
		re.withTag(log.Info).Msgf("  -Pipeline[%s]: %s", proto, x.SimpleClassName(pi))
		devices = append(devices, pi.GetDevices()...)
	}
	re.withTag(log.Info).Msgf("已加载 Devices: %d", len(devices))
	for _, it := range devices {
		re.withTag(log.Info).Msg("  - Device: " + x.SimpleClassName(it))
	}

	re.withTag(log.Info).Msgf("已加载 Drivers: %d", re.drivers.Len())
	x.ForEach(re.drivers, func(it interface{}) {
		re.withTag(log.Info).Msg("  - Driver: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载 Triggers: %d", re.triggers.Len())
	x.ForEach(re.triggers, func(it interface{}) {
		re.withTag(log.Info).Msg("  - Trigger: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载 Plugins: %d", re.plugins.Len())
	x.ForEach(re.plugins, func(it interface{}) {
		re.withTag(log.Info).Msg("  - Plugin: " + x.SimpleClassName(it))
	})
}

func (re *Registration) RegisterBundleFactory(typeName string, factory BundleFactory) {
	if _, ok := re.bundleFactories[typeName]; ok {
		re.withTag(log.Warn).Msgf("组件类型[%s]工厂函数被覆盖注册： %s", typeName, x.SimpleClassName(factory))
	}
	re.withTag(log.Info).Msgf("正在注册组件工厂函数： %s", typeName)
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

// 注册组件，如果注册失败，返回False
func (re *Registration) registerBundlesIfHit(configs conf.Map, initAct func(bundle Initialize, args map[string]interface{})) bool {
	if 0 == len(configs) {
		return false
	}
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
		}
		// 根据类型注册
		bundle := factory()
		switch bundle.(type) {

		case ProtoPipeline:
			re.AddProtoPipeline(bundle.(ProtoPipeline))

		case Interceptor:
			it := bundle.(Interceptor)
			it.setPriority(int(config.MustInt64("priority")))
			re.AddInterceptor(it)

		case Driver:
			re.AddDriver(bundle.(Driver))

		case VirtualDevice:
			dev := bundle.(VirtualDevice)
			if name := config.MustString("name"); "" == name {
				dev.setDisplayName(typeName)
			} else {
				dev.setDisplayName(name)
			}
			group := config.MustString("groupAddress")
			if "" == group {
				re.withTag(log.Panic).Msgf("配置项[groupAddress]是必填参数")
			}
			phy := config.MustString("physicalAddress")
			if "" == phy {
				re.withTag(log.Panic).Msgf("配置项[physicalAddress]是必填参数")
			}
			dev.setGroupAddress(group)
			dev.setPhyAddress(phy)
			re.AddVirtualDevice(dev)

		case Trigger:
			tr := bundle.(Trigger)
			tp := config.MustString("topic")
			if "" == tp {
				re.withTag(log.Panic).Msgf("配置项[topic]是必填参数")
			}
			tr.setTopic(tp)
			re.AddTrigger(tr)

		default:
			if plg, ok := bundle.(Plugin); ok {
				re.AddPlugin(plg)
			} else {
				re.withTag(log.Panic).Msgf("未支持的组件类型：%s", typeName)
			}
		}

		// 需要Topic过滤
		if tf, ok := bundle.(NeedTopicFilter); ok {
			if topics, err := config.MustStringArray("topics"); nil != err || 0 == len(topics) {
				re.withTag(log.Panic).Err(err).Msgf("配置项中[topics]必须是字符串数组： %s", typeName)
			} else {
				tf.setTopics(topics)
			}
		}

		// 组件初始化。由外部函数处理，减少不必要的依赖
		if init, ok := bundle.(Initialize); ok {
			initAct(init, map[string]interface{}(config.MustMap("InitArgs")))
		}
	}
	return true
}

func (re *Registration) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Registration")
}
