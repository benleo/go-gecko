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
	namedOutputs  map[string]OutputDevice
	namedInputs   map[string]InputDevice
	namedDecoders map[string]Decoder
	namedEncoders map[string]Encoder
	plugins       *list.List
	interceptors  *list.List
	drivers       *list.List
	outputs       *list.List
	inputs        *list.List
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
	re.namedOutputs = make(map[string]OutputDevice)
	re.namedInputs = make(map[string]InputDevice)
	re.namedDecoders = make(map[string]Decoder)
	re.namedEncoders = make(map[string]Encoder)
	re.plugins = list.New()
	re.interceptors = list.New()
	re.drivers = list.New()
	re.inputs = list.New()
	re.outputs = list.New()
	re.startBeforeHooks = list.New()
	re.startAfterHooks = list.New()
	re.stopBeforeHooks = list.New()
	re.stopAfterHooks = list.New()
	re.bundleFactories = make(map[string]BundleFactory)
	return re
}

// 添加Encoder
func (re *Registration) AddEncoder(name string, encoder Encoder) {
	if _, ok := re.namedEncoders[name]; ok {
		re.withTag(log.Panic).Msgf("Encoder类型重复: %s", name)
	} else {
		re.namedEncoders[name] = encoder
	}
}

// 添加Decoder
func (re *Registration) AddDecoder(name string, decoder Decoder) {
	if _, ok := re.namedDecoders[name]; ok {
		re.withTag(log.Panic).Msgf("Decoder类型重复: %s", name)
	} else {
		re.namedDecoders[name] = decoder
	}
}

// 添加OutputDevice
func (re *Registration) AddOutputDevice(device OutputDevice) {
	addr := device.GetUnionAddress()
	if _, ok := re.namedOutputs[addr]; ok {
		re.withTag(log.Panic).Msgf("OutputDevice设备地址重复: %s", addr)
	} else {
		re.namedOutputs[addr] = device
		re.outputs.PushBack(device)
	}
}

// 添加InputDevice
func (re *Registration) AddInputDevice(device InputDevice) {
	addr := device.GetUnionAddress()
	if _, ok := re.namedInputs[addr]; ok {
		re.withTag(log.Panic).Msgf("InputDevice设备地址重复: %s", addr)
	} else {
		re.namedInputs[addr] = device
		re.inputs.PushBack(device)
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

func (re *Registration) showBundles() {
	re.withTag(log.Info).Msgf("已加载 Interceptors: %d", re.interceptors.Len())
	x.ForEach(re.interceptors, func(it interface{}) {
		re.withTag(log.Info).Msg("  - Interceptor: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载 InputDevices: %d", re.inputs.Len())
	x.ForEach(re.inputs, func(it interface{}) {
		re.withTag(log.Info).Msg("  - InputDevice: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载OutputDevices: %d", re.outputs.Len())
	x.ForEach(re.outputs, func(it interface{}) {
		re.withTag(log.Info).Msg("  - OutputDevice: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载 Drivers: %d", re.drivers.Len())
	x.ForEach(re.drivers, func(it interface{}) {
		re.withTag(log.Info).Msg("  - Driver: " + x.SimpleClassName(it))
	})

	re.withTag(log.Info).Msgf("已加载 Plugins: %d", re.plugins.Len())
	x.ForEach(re.plugins, func(it interface{}) {
		re.withTag(log.Info).Msg("  - Plugin: " + x.SimpleClassName(it))
	})
}

// 注册组件工厂函数
func (re *Registration) RegisterBundleFactory(typeName string, factory BundleFactory) {
	if _, ok := re.bundleFactories[typeName]; ok {
		re.withTag(log.Warn).Msgf("组件类型[%s]，旧的工厂函数将被覆盖为： %s", typeName, x.SimpleClassName(factory))
	}
	re.withTag(log.Info).Msgf("正在注册组件工厂函数： %s", typeName)
	re.bundleFactories[typeName] = factory
}

// 注册编码解码工厂函数
func (re *Registration) RegisterCodecFactory(typeName string, factory CodecFactory) {
	codec := factory()
	switch codec.(type) {
	case Decoder:
		re.AddDecoder(typeName, codec.(Decoder))

	case Encoder:
		re.AddEncoder(typeName, codec.(Encoder))

	default:
		re.withTag(log.Panic).Msgf("未知的编/解码类型[%s]，工厂函数： %s", typeName, x.SimpleClassName(factory))
	}
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
func (re *Registration) registerBundlesIfHit(configs *conf.ImmutableMap,
	initAct func(bundle Initialize, args map[string]interface{})) bool {
	if configs.IsEmpty() {
		return false
	}
	configs.ForEach(func(bundleType string, item interface{}) {
		asMap, ok := item.(map[string]interface{})
		if !ok {
			re.withTag(log.Panic).Msgf("组件配置信息类型错误: %s", bundleType)
		}
		config := conf.MapToMap(asMap)
		if config.MustBool("disable") {
			re.withTag(log.Warn).Msgf("组件[%s]在配置中禁用", bundleType)
			return
		}

		// 配置选项中，指定 type 字段为类型名称
		if typeName := config.MustString("type"); "" != typeName {
			bundleType = typeName
		}

		factory, ok := re.findFactory(bundleType)
		if !ok {
			re.withTag(log.Panic).Msgf("组件类型[%s]，没有注册对应的工厂函数", bundleType)
		}
		// 根据类型注册
		bundle := factory()
		switch bundle.(type) {

		case Driver:
			re.AddDriver(bundle.(Driver))

		case Interceptor:
			it := bundle.(Interceptor)
			it.setPriority(int(config.MustInt64("priority")))
			re.AddInterceptor(it)

		case VirtualDevice:
			device := bundle.(VirtualDevice)
			if name := config.MustString("displayName"); "" == name {
				re.withTag(log.Panic).Msgf("VirtualDevice[%s]配置项[displayName]是必填参数", bundleType)
			} else {
				device.setDisplayName(name)
			}

			if group := config.MustString("groupAddress"); "" == group {
				re.withTag(log.Panic).Msgf("VirtualDevice[%s]配置项[groupAddress]是必填参数", bundleType)
			} else {
				device.setGroupAddress(group)
			}

			if private := config.MustString("privateAddress"); "" == private {
				re.withTag(log.Panic).Msgf("VirtualDevice[%s]配置项[privateAddress]是必填参数", bundleType)
			} else {
				device.setPrivateAddress(private)
			}

			if name := config.MustString("encoder"); "" == name {
				if nil == device.GetEncoder() {
					re.withTag(log.Panic).Msgf("未设置默认Encoder时，Device[%s]配置项[encoder]是必填参数", bundleType)
				}
			} else {
				if encoder, ok := re.namedEncoders[name]; ok {
					device.setEncoder(encoder)
				} else {
					re.withTag(log.Panic).Msgf("Encoder[%s]未注册", name)
				}
			}

			if name := config.MustString("decoder"); "" == name {
				if nil == device.GetDecoder() {
					re.withTag(log.Panic).Msgf("未设置默认Decoder时，Device[%s]配置项[decoder]是必填参数", bundleType)
				}
			} else {
				if decoder, ok := re.namedDecoders[name]; ok {
					device.setDecoder(decoder)
				} else {
					re.withTag(log.Panic).Msgf("Decoder[%s]未注册", name)
				}
			}

			if inputDevice, ok := device.(InputDevice); ok {
				re.AddInputDevice(inputDevice)
			} else if outputDevice, ok := device.(OutputDevice); ok {
				re.AddOutputDevice(outputDevice)
			} else {
				re.withTag(log.Panic).Msgf("未知VirtualDevice类型： %s", x.SimpleClassName(device))
			}

		default:
			if plg, ok := bundle.(Plugin); ok {
				re.AddPlugin(plg)
			} else {
				re.withTag(log.Panic).Msgf("未支持的组件类型：%s. 你是否没有实现某个函数接口？", bundleType)
			}
		}

		// 需要Topic过滤
		if tf, ok := bundle.(NeedTopicFilter); ok {
			if topics, err := config.MustStringArray("topics"); nil != err || 0 == len(topics) {
				re.withTag(log.Panic).Err(err).Msgf("配置项中[topics]必须是字符串数组： %s", bundleType)
			} else {
				tf.setTopics(topics)
			}
		}

		// 组件初始化。由外部函数处理，减少不必要的依赖
		if init, ok := bundle.(Initialize); ok {
			initAct(init, map[string]interface{}(config.MustMap("InitArgs")))
		}
	})
	return true
}

func (re *Registration) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Registration")
}
