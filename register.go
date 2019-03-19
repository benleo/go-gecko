package gecko

import (
	"container/list"
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko/utils"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 负责对Engine组件的注册管理
type Register struct {
	// 组件管理
	uuidOutputs   map[string]OutputDevice
	uuidInputs    map[string]InputDevice
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
	factories map[string]Factory
}

func prepare() *Register {
	re := new(Register)
	re.uuidOutputs = make(map[string]OutputDevice)
	re.uuidInputs = make(map[string]InputDevice)
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
	re.factories = make(map[string]Factory)
	return re
}

// 添加Encoder
func (re *Register) AddEncoder(name string, encoder Encoder) {
	if _, ok := re.namedEncoders[name]; ok {
		ZapSugarLogger.Panicw("Encoder类型重复", "type", name)
	} else {
		re.namedEncoders[name] = encoder
	}
}

// 添加Decoder
func (re *Register) AddDecoder(name string, decoder Decoder) {
	if _, ok := re.namedDecoders[name]; ok {
		ZapSugarLogger.Panicw("Decoder类型重复", "type", name)
	} else {
		re.namedDecoders[name] = decoder
	}
}

// 添加OutputDevice
func (re *Register) AddOutputDevice(device OutputDevice) {
	uuid := re.ensureUniqueUUID(device.GetUuid())
	re.uuidOutputs[uuid] = device
	re.outputs.PushBack(device)
}

// 添加InputDevice
func (re *Register) AddInputDevice(device InputDevice) {
	uuid := re.ensureUniqueUUID(device.GetUuid())
	re.uuidInputs[uuid] = device
	re.inputs.PushBack(device)
}

// 添加Plugin
func (re *Register) AddPlugin(plugin Plugin) {
	re.plugins.PushBack(plugin)
}

// 添加Interceptor
func (re *Register) AddInterceptor(interceptor Interceptor) {
	re.interceptors.PushBack(interceptor)
}

// 添加Driver
func (re *Register) AddDriver(driver Driver) {
	re.drivers.PushBack(driver)
}

func (re *Register) AddStartBeforeHook(hook HookFunc) {
	re.startBeforeHooks.PushBack(hook)
}

func (re *Register) AddStartAfterHook(hook HookFunc) {
	re.startAfterHooks.PushBack(hook)
}

func (re *Register) AddStopBeforeHook(hook HookFunc) {
	re.stopBeforeHooks.PushBack(hook)
}

func (re *Register) AddStopAfterHook(hook HookFunc) {
	re.startAfterHooks.PushBack(hook)
}

func (re *Register) showBundles() {
	zlog := ZapSugarLogger
	zlog.Infof("已加载 Interceptors: %d", re.interceptors.Len())
	utils.ForEach(re.interceptors, func(it interface{}) {
		zlog.Info("  - Interceptor: " + utils.GetClassName(it))
	})

	zlog.Infof("已加载 InputDevices: %d", re.inputs.Len())
	utils.ForEach(re.inputs, func(it interface{}) {
		zlog.Info("  - InputDevice: " + utils.GetClassName(it))
		for _, shadow := range it.(InputDevice).GetLogicList() {
			zlog.Info("    - Logic: " + utils.GetClassName(shadow))
		}
	})

	zlog.Infof("已加载OutputDevices: %d", re.outputs.Len())
	utils.ForEach(re.outputs, func(it interface{}) {
		zlog.Info("  - OutputDevice: " + utils.GetClassName(it))
	})

	zlog.Infof("已加载 Drivers: %d", re.drivers.Len())
	utils.ForEach(re.drivers, func(it interface{}) {
		zlog.Info("  - Driver: " + utils.GetClassName(it))
	})

	zlog.Infof("已加载 Plugins: %d", re.plugins.Len())
	utils.ForEach(re.plugins, func(it interface{}) {
		zlog.Info("  - Plugin: " + utils.GetClassName(it))
	})
}

// 注册组件工厂函数
func (re *Register) AddFactory(typeName string, factory Factory) {
	zlog := ZapSugarLogger
	if _, ok := re.factories[typeName]; ok {
		zlog.Warnf("组件类型[%s]，旧的工厂函数将被覆盖为： %s", typeName, utils.GetClassName(factory))
	}
	zlog.Infof("正在注册组件工厂函数： %s", typeName)
	re.factories[typeName] = factory
}

// 注册编码解码工厂函数
func (re *Register) AddCodecFactory(typeName string, factory CodecFactory) {
	codec := factory()
	switch codec.(type) {
	case Decoder:
		re.AddDecoder(typeName, codec.(Decoder))

	case Encoder:
		re.AddEncoder(typeName, codec.(Encoder))

	default:
		ZapSugarLogger.Panicf("未知的编/解码类型[%s]，工厂函数： %s", typeName, utils.GetClassName(factory))
	}
}

// 查找指定类型的
func (re *Register) findFactory(typeName string) (Factory, bool) {
	if f, ok := re.factories[typeName]; ok {
		return f, true
	} else {
		return nil, false
	}
}

func (re *Register) ensureUniqueUUID(uuid string) string {
	zlog := ZapSugarLogger
	if _, ok := re.uuidInputs[uuid]; ok {
		zlog.Panicf("设备UUID重复[Input]：%s", uuid)
	} else if _, ok := re.uuidOutputs[uuid]; ok {
		zlog.Panicf("设备UUID重复[Output]：%s", uuid)
	}
	return uuid
}

func (re *Register) factory(cType string, configItem interface{}) (obj interface{}, bType string, config *cfg.Config, ok bool) {
	zlog := ZapSugarLogger
	asMap, ok := configItem.(map[string]interface{})
	if !ok {
		zlog.Panicf("组件配置信息类型错误: %s", cType)
	}
	wrap := cfg.Wrap(asMap)
	if wrap.MustBool("disable") {
		zlog.Infof("组件[%s]在配置中禁用", cType)
		return nil, cType, nil, false
	}

	// 配置选项中，指定 type 字段为类型名称
	if typeName := wrap.MustString("type"); "" != typeName {
		cType = typeName
	}

	factory, ok := re.findFactory(cType)
	if !ok {
		zlog.Panicf("组件类型[%s]，没有注册对应的工厂函数", cType)
	}
	return factory(), cType, wrap, true
}

func (re *Register) register(configs *cfg.Config,
	initFn func(component Initial, args *cfg.Config),
	structInitFn func(component StructuredInitial, args *cfg.Config)) {

	// 组件初始化。由外部函数处理，减少不必要的依赖注入
	configs.ForEach(func(rawType string, item interface{}) {
		component, config := re.register0(rawType, item)
		// 初始化0
		args := config.MustConfig("InitArgs")
		if init, ok := component.(Initial); ok {
			initFn(init, args)
		} else
		// 初始化1
		if init, ok := component.(StructuredInitial); ok {
			structInitFn(init, args)
		}
	})
}

func (re *Register) register0(rawType string, item interface{}) (interface{}, *cfg.Config) {
	zlog := ZapSugarLogger
	component, componentType, config, ok := re.factory(rawType, item)
	if !ok {
		return nil, config
	}

	switch component.(type) {

	case Driver:
		re.AddDriver(component.(Driver))

	case Interceptor:
		it := component.(Interceptor)
		it.setPriority(int(config.MustInt64("priority")))
		re.AddInterceptor(it)

	case VirtualDevice:
		device := component.(VirtualDevice)
		device.setName(required(config.MustString("name"),
			"VirtualDevice[%s]配置项[name]是必填参数", componentType))

		device.setUuid(required(config.MustString("uuid"),
			"VirtualDevice[%s]配置项[uuid]是必填参数", componentType))

		encoderName := required(config.MustString("encoder"),
			"未设置默认Encoder时，Device[%s]配置项[encoder]是必填参数", componentType)
		if encoder, ok := re.namedEncoders[encoderName]; ok {
			device.setEncoder(encoder)
		} else {
			zlog.Panicf("Encoder[%s]未注册", encoderName)
		}

		decoderName := required(config.MustString("decoder"),
			"未设置默认Decoder时，Device[%s]配置项[decoder]是必填参数", componentType)
		if decoder, ok := re.namedDecoders[decoderName]; ok {
			device.setDecoder(decoder)
		} else {
			zlog.Panicf("Decoder[%s]未注册", decoderName)
		}

		if inputDevice, ok := device.(InputDevice); ok {
			inputDevice.setTopic(required(config.MustString("topic"),
				"VirtualDevice[%s]配置项[topic]是必填参数", componentType))
			re.AddInputDevice(inputDevice)
		} else if outputDevice, ok := device.(OutputDevice); ok {
			re.AddOutputDevice(outputDevice)
		} else {
			zlog.Panicf("未知VirtualDevice类型： %s", utils.GetClassName(device))
		}

	case LogicDevice:
		logic := component.(LogicDevice)
		logic.setUuid(required(config.MustString("uuid"),
			"LogicDevice[%s]配置项[uuid]是必填参数", componentType))

		logic.setName(required(config.MustString("name"),
			"LogicDevice[%s]配置项[name]是必填参数", componentType))

		logic.setTopic(required(config.MustString("topic"),
			"LogicDevice[%s]配置项[topic]是必填参数", componentType))

		masterUuid := required(config.MustString("masterUuid"),
			"LogicDevice[%s]配置项[masterUuid]是必填参数", componentType)
		logic.setMasterUuid(masterUuid)

		// Add to input
		if input, ok := re.uuidInputs[masterUuid]; ok {
			if err := input.addLogic(logic); nil != err {
				zlog.Panic("LogicDevice挂载到MasterInputDevice发生错误", err)
			}
		} else {
			zlog.Panicf("LogicDevice[%s]配置项[masterUuid]是没找到对应设备", componentType)
		}

	default:
		if plg, ok := component.(Plugin); ok {
			re.AddPlugin(plg)
		} else {
			zlog.Panicf("未支持的组件类型：%s. 你是否没有实现某个函数接口？", componentType)
		}
	}

	// Interceptor / Driver 需要Topic过滤
	if tf, ok := component.(NeedTopicFilter); ok {
		if topics, err := config.MustStringArray("topics"); nil != err || 0 == len(topics) {
			zlog.Panicw("配置项中[topics]必须是字符串数组", "type", componentType, "error", err)
		} else {
			tf.setTopics(topics)
		}
	}

	return component, config
}

func required(value, template string, args ...interface{}) string {
	if "" == value {
		ZapSugarLogger.Panicf(template, args...)
	}
	return value
}
