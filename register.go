package gecko

import (
	"container/list"
	"github.com/yoojia/go-gecko/utils"
	"github.com/yoojia/go-value"
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

func newRegister() *Register {
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
		log.Panicw("Encoder类型重复", "type", name)
	} else {
		re.namedEncoders[name] = encoder
	}
}

// 添加Decoder
func (re *Register) AddDecoder(name string, decoder Decoder) {
	if _, ok := re.namedDecoders[name]; ok {
		log.Panicw("Decoder类型重复", "type", name)
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

func (re *Register) showComponents() {
	log.Infof("已加载 Interceptors: %d", re.interceptors.Len())
	utils.ForEach(re.interceptors, func(it interface{}) {
		log.Infof("  -> Interceptor: [%s::%s]", utils.GetClassName(it), it.(NeedName).GetName())
	})

	log.Infof("已加载 InputDevices: %d", re.inputs.Len())
	utils.ForEach(re.inputs, func(it interface{}) {
		typeName := utils.GetClassName(it)
		log.Infof("  -> InputDevice: [%s::%s]", typeName, it.(NeedName).GetName())
		for _, shadow := range it.(InputDevice).GetLogicList() {
			log.Info("    --> Logic: " + utils.GetClassName(shadow))
		}
	})

	log.Infof("已加载OutputDevices: %d", re.outputs.Len())
	utils.ForEach(re.outputs, func(it interface{}) {
		typeName := utils.GetClassName(it)
		log.Infof("  -> OutputDevice: [%s::%s]", typeName, it.(NeedName).GetName())
	})

	log.Infof("已加载 Drivers: %d", re.drivers.Len())
	utils.ForEach(re.drivers, func(it interface{}) {
		log.Infof("  -> Driver: [%s::%s]", utils.GetClassName(it), it.(NeedName).GetName())
	})

	log.Infof("已加载 Plugins: %d", re.plugins.Len())
	utils.ForEach(re.plugins, func(it interface{}) {
		log.Info("  -> Plugin: " + utils.GetClassName(it))
	})
}

// 注册组件工厂函数
func (re *Register) AddFactory(typeName string, factory Factory) {
	if _, ok := re.factories[typeName]; ok {
		log.Warnf("组件类型[%s]，旧的工厂函数将被覆盖为： %s", typeName, utils.GetClassName(factory))
	}
	log.Infof("正在注册组件工厂函数： %s", typeName)
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
		log.Panicf("未知的编/解码类型[%s]，工厂函数： %s", typeName, utils.GetClassName(factory))
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
	if _, ok := re.uuidInputs[uuid]; ok {
		log.Panicf("设备UUID重复[Input]：%s", uuid)
	} else if _, ok := re.uuidOutputs[uuid]; ok {
		log.Panicf("设备UUID重复[Output]：%s", uuid)
	}
	return uuid
}

func (re *Register) factory(rawType string, rawCfg interface{}) (obj interface{}, typeName string, config map[string]interface{}, ok bool) {
	config, ok = rawCfg.(map[string]interface{})
	if !ok {
		log.Panicf("组件配置信息类型错误: %s", rawType)
		return nil, rawType, nil, false
	}

	if value.Of(config["disable"]).MustBool() {
		log.Infof("组件[%s]在配置中禁用", rawType)
		return nil, rawType, nil, false
	}

	// 配置选项中，指定 type 字段为类型名称
	if typeName := value.Of(config["type"]).String(); "" != typeName {
		rawType = typeName
	}

	factory, found := re.findFactory(rawType)
	if !found {
		log.Panicf("组件类型[%s]，没有注册对应的工厂函数", rawType)
		return nil, rawType, nil, false
	} else {
		return factory(), rawType, config, true
	}
}

func (re *Register) register(
	configs map[string]interface{},
	initFn func(initial Initial, args map[string]interface{}),
	structInitFn func(initial StructuredInitial, args map[string]interface{})) {
	// 组件初始化。由外部函数处理，减少不必要的依赖注入
	for keyAsTypeName, item := range configs {
		component, config := re.register0(keyAsTypeName, item)
		if nil == component || config == nil {
			continue
		}
		// 初始化0
		args := utils.ToMap(config["InitArgs"])
		if nil == args {
			continue
		}
		if init, ok := component.(Initial); ok {
			initFn(init, args)
		} else if init, ok := component.(StructuredInitial); ok {
			structInitFn(init, args)
		}
	}

}

func (re *Register) register0(keyAsTypeName string, item interface{}) (interface{}, map[string]interface{}) {
	component, componentType, config, ok := re.factory(keyAsTypeName, item)
	if !ok {
		return nil, nil
	}
	name := value.Of(config["name"]).String()
	switch component.(type) {

	case Driver:
		driver := component.(Driver)
		re.AddDriver(driver)
		if "" != name {
			driver.setName(name)
		} else {
			driver.setName(keyAsTypeName)
		}

	case Interceptor:
		it := component.(Interceptor)
		it.setPriority(int(value.Of(config["priority"]).MustInt64()))
		if "" != name {
			it.setName(name)
		} else {
			it.setName(keyAsTypeName)
		}
		re.AddInterceptor(it)

	case VirtualDevice:
		device := component.(VirtualDevice)
		device.setName(required(name,
			"VirtualDevice[%s::%s]配置项[name]是必填参数", componentType, keyAsTypeName))

		device.setUuid(required(value.Of(config["uuid"]).String(),
			"VirtualDevice[%s::%s]配置项[uuid]是必填参数", componentType, keyAsTypeName))

		if nil == device.GetEncoder() {
			encoder := required(value.Of(config["encoder"]).String(),
				"未设置默认Encoder时，Device[%s]配置项[encoder]是必填参数", componentType)
			if encoder, ok := re.namedEncoders[encoder]; ok {
				device.setEncoder(encoder)
			} else {
				log.Panicf("Encoder[%s]未注册", encoder)
			}
		}

		if nil == device.GetDecoder() {
			decoder := required(value.Of(config["decoder"]).String(),
				"未设置默认Decoder时，Device[%s]配置项[decoder]是必填参数", componentType)
			if decoder, ok := re.namedDecoders[decoder]; ok {
				device.setDecoder(decoder)
			} else {
				log.Panicf("Decoder[%s]未注册", decoder)
			}
		}

		if inputDevice, ok := device.(InputDevice); ok {
			inputDevice.setTopic(required(value.Of(config["topic"]).String(),
				"VirtualDevice[%s]配置项[topic]是必填参数", componentType))
			re.AddInputDevice(inputDevice)
		} else if outputDevice, ok := device.(OutputDevice); ok {
			re.AddOutputDevice(outputDevice)
		} else {
			log.Panicf("未知VirtualDevice类型： %s", utils.GetClassName(device))
		}

	case LogicDevice:
		logic := component.(LogicDevice)
		logic.setUuid(required(value.Of(config["uuid"]).String(),
			"LogicDevice[%s]配置项[uuid]是必填参数", componentType))

		logic.setName(required(name,
			"LogicDevice[%s]配置项[name]是必填参数", componentType))

		logic.setTopic(required(value.Of(config["topic"]).String(),
			"LogicDevice[%s]配置项[topic]是必填参数", componentType))

		masterUuid := required(value.Of(config["masterUuid"]).String(),
			"LogicDevice[%s]配置项[masterUuid]是必填参数", componentType)
		logic.setMasterUuid(masterUuid)

		// Add to input
		if input, ok := re.uuidInputs[masterUuid]; ok {
			if err := input.addLogic(logic); nil != err {
				log.Panic("LogicDevice挂载到MasterInputDevice发生错误", err)
			}
		} else {
			log.Panicf("LogicDevice[%s]配置项[masterUuid]是没找到对应设备", componentType)
		}

	default:
		if plg, ok := component.(Plugin); ok {
			re.AddPlugin(plg)
		} else {
			log.Panicf("未支持的组件类型：[%s::%s]. 你是否没有实现某个函数接口？", componentType, keyAsTypeName)
		}
	}

	// Interceptor / Driver 需要Topic过滤
	if tf, ok := component.(NeedTopicFilter); ok {
		if topics := utils.ToStringArray(config["topics"]); 0 == len(topics) {
			log.Panicw("配置项中[topics]必须是字符串数组", "type", componentType)
		} else {
			tf.setTopics(topics)
		}
	}

	return component, config
}

func required(value, template string, args ...interface{}) string {
	if "" == value {
		log.Panicf(template, args...)
	}
	return value
}
