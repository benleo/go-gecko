package gecko

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yoojia/go-gecko/structs"
	"github.com/yoojia/go-gecko/utils"
	"github.com/yoojia/go-value"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"
)

////

// 默认组件生命周期超时时间：3秒
const DefaultLifeCycleTimeout = time.Second * 3

// Pipeline管理内部组件，处理事件。
type Pipeline struct {
	*Register
	context Context
	// 事件派发
	interceptorChan chan *session
	driverChan      chan *session
	triggerChan     chan *session
	//
	dispatchCtx    context.Context
	dispatchCancel context.CancelFunc
}

// 初始化Pipeline
func (p *Pipeline) Init(config map[string]interface{}) {
	p.context = p.newGeckoContext(config)
	p.context.prepare()

	capacity := value.Of(p.context.gecko()["eventsCapacity"]).Int64OrDefault(64)
	if capacity <= 0 {
		capacity = 1
	}
	log.Infof("事件通道容量： %d", capacity)
	p.interceptorChan = make(chan *session, capacity)
	p.driverChan = make(chan *session, capacity)
	p.triggerChan = make(chan *session, capacity)

	// 初始化组件：根据配置文件指定项目
	initFn := func(it Initial, args map[string]interface{}) {
		it.OnInit(args, p.context)
	}
	// 使用结构化的参数来初始化
	structInitFn := func(it StructuredInitial, args map[string]interface{}) {
		structConfig := it.StructuredConfig()
		m2sDecoder, err := structs.NewDecoder(&structs.DecoderConfig{
			TagName: "toml",
			Result:  structConfig,
		})
		if nil != err {
			log.Panic("无法创建Map2Struct解码器", err)
		}
		if err := m2sDecoder.Decode(args); nil != err {
			log.Panic("Map2Struct解码出错", err)
		}
		it.Init(structConfig, p.context)
	}

	ctx := p.context.(*_GeckoContext)
	if 0 == len(ctx.cfgPlugins) {
		log.Warn("警告：未配置任何[Plugin]组件")
	} else {
		p.register(ctx.cfgPlugins, initFn, structInitFn)
	}
	if 0 == len(ctx.cfgOutputs) {
		log.Fatal("严重：未配置任何[OutputDevice]组件")
	} else {
		p.register(ctx.cfgOutputs, initFn, structInitFn)
	}
	if 0 == len(ctx.cfgInterceptors) {
		log.Warn("警告：未配置任何[Interceptor]组件")
	} else {
		p.register(ctx.cfgInterceptors, initFn, structInitFn)
	}
	if 0 == len(ctx.cfgDrivers) {
		log.Warn("警告：未配置任何[Driver]组件")
	} else {
		p.register(ctx.cfgDrivers, initFn, structInitFn)
	}
	if 0 == len(ctx.cfgTriggers) {
		log.Warn("警告：未配置任何[Trigger]组件")
	} else {
		p.register(ctx.cfgTriggers, initFn, structInitFn)
	}
	if 0 == len(ctx.cfgInputs) {
		log.Fatal("严重：未配置任何[InputDevice]组件")
	} else {
		p.register(ctx.cfgInputs, initFn, structInitFn)
	}
	if 0 == len(ctx.cfgLogics) {
		log.Warn("警告：未配置任何[LogicDevice]组件")
	} else {
		p.register(ctx.cfgLogics, initFn, structInitFn)
	}
	// show
	p.showComponents()
}

// 启动Pipeline
func (p *Pipeline) Start() {
	log.Info("Pipeline启动...")
	// 检查运行时依赖关系:
	// 每个Input产生的Topic,只允许单独一个driver处理, 不允许多个Driver处理同一个Topic
	utils.ForEach(p.inputs, func(it interface{}) {
		topic := it.(InputDevice).GetTopic()
		hits := make([]Driver, 0)
		utils.ForEach(p.drivers, func(dr interface{}) {
			driver := dr.(Driver)
			if anyTopicMatches(driver.GetTopicExpr(), topic) {
				hits = append(hits, driver)
			}
		})
		if len(hits) > 1 {
			for _, dr := range hits {
				log.Error("Topic被多个Driver处理, Driver: %s, Topic: %s", dr.GetName(), topic)
			}
			log.Panicf("禁止多个Driver处理相同的Topic")
		}
	})

	// Dispatch
	go func() {
		for {
			select {
			case <-p.dispatchCtx.Done():
				return

			case s := <-p.interceptorChan:
				go p.doInterceptor(s)

			case s := <-p.driverChan:
				go p.doDriver(s)

			case s := <-p.triggerChan:
				go p.doTrigger(s)

			}
		}
	}()

	// Hook first
	utils.ForEach(p.startBeforeHooks, func(it interface{}) { it.(HookFunc)(p) })
	// Plugins
	utils.ForEach(p.plugins, p.callStartFunc)
	// Outputs
	utils.ForEach(p.outputs, p.callStartFunc)
	// Drivers
	utils.ForEach(p.drivers, p.callStartFunc)
	// Triggers
	utils.ForEach(p.triggers, p.callStartFunc)
	// Inputs
	utils.ForEach(p.inputs, p.callStartFunc)
	// Then, Serve inputs
	utils.ForEach(p.inputs, func(it interface{}) {
		input := it.(InputDevice)
		uuid := input.GetUuid()
		go func() {
			defer log.Debugf("InputDevice已经停止：%s", uuid)
			if err := input.Serve(p.context, p.newInputDeliverer(input)); nil != err {
				log.Errorw("InputDevice服务运行错误",
					"uuid", uuid,
					"error", err,
					"class", utils.GetClassName(input))
			}
		}()
	})
	// Hook After
	utils.ForEach(p.startAfterHooks, func(it interface{}) { it.(HookFunc)(p) })

	log.Info("Pipeline启动...OK")
}

// 停止Pipeline
func (p *Pipeline) Stop() {
	log.Info("Pipeline停止...")
	// Hook first
	utils.ForEach(p.stopBeforeHooks, func(it interface{}) { it.(HookFunc)(p) })
	// Inputs
	utils.ForEach(p.inputs, p.callStopFunc)
	// Drivers
	utils.ForEach(p.drivers, p.callStopFunc)
	// Triggers
	utils.ForEach(p.triggers, p.callStopFunc)
	// Outputs
	utils.ForEach(p.outputs, p.callStopFunc)
	// Plugins
	utils.ForEach(p.plugins, p.callStopFunc)
	// Hook After
	utils.ForEach(p.stopAfterHooks, func(it interface{}) { it.(HookFunc)(p) })

	log.Info("Pipeline停止...OK")
	// 最终发起Dispatch停止信号
	p.dispatchCancel()
}

// 等待系统停止信号
func (p *Pipeline) AwaitTermination() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	log.Info("接收到系统停止信号")
}

// 准备运行环境，初始化相关组件
func (p *Pipeline) prepareEnv() {
	p.dispatchCtx, p.dispatchCancel = context.WithCancel(context.Background())
}

// 创建InputDeliverer函数
// InputDeliverer函数对于InputDevice对象是一个系统内部数据传输流程的代理函数。
// 每个Deliver请求，都会向系统发起请求，并获取系统处理结果响应数据。也意味着，InputDevice发起的每个请求
// 都会执行 Decode -> Deliver(GeckoKernelFlow) -> Encode 流程。
func (p *Pipeline) newInputDeliverer(masterInput InputDevice) InputDeliverer {
	return InputDeliverer(func(topic string, rawFrame FramePacket) (FramePacket, error) {
		// 从Input设备中读取Decode数据
		masterUuid := masterInput.GetUuid()
		if nil == rawFrame {
			return nil, errors.New("Input设备发起Deliver请求必须携带参数数据")
		}
		inputMessage, err := masterInput.GetDecoder()(rawFrame)
		if nil != err {
			return nil, errors.WithMessage(err, "Input设备Decode数据出错："+masterUuid)
		}
		attributes := make(map[string]interface{})
		attributes["@InputDevice.Type"] = utils.GetClassName(masterInput)
		attributes["@InputDevice.Name"] = masterInput.GetName()

		inputUuid := masterUuid
		inputTopic := topic

		var logic LogicDevice = nil
		// 查找符合条件的逻辑设备，并转换数据
		for _, item := range masterInput.GetLogicList() {
			if item.CheckIfMatch(inputMessage) {
				logic = item
				attributes["@InputDevice.Logic.Type"] = utils.GetClassName(logic)
				attributes["@InputDevice.Logic.Name"] = logic.GetName()
				break
			}
		}
		if logic != nil {
			inputUuid = logic.GetUuid()
			inputTopic = logic.GetTopic()
			inputMessage = logic.Transform(inputMessage)
		}
		// 发送到Dispatcher调度处理
		session := &session{
			attrs:     newMapAttributesWith(attributes),
			timestamp: time.Now(),
			topic:     inputTopic,
			uuid:      inputUuid,
			inbound:   inputMessage,
			outbound:  make(chan *MessagePacket, 1),
		}

		// 传递给interceptor通道来处理,并等待Session处理完成
		p.interceptorChan <- session
		// 等待
		output := <-session.outbound
		if nil == output {
			return nil, errors.New("Input设备发起Deliver请求必须返回结果数据")
		}
		// 输出调度Attr数据
		p.context.OnIfLogV(func() {
			for k, v := range session.Attrs().Map() {
				if '@' == k[0] {
					log.Debugf("||-> Session属性 %s = %v", k, v)
				}
			}
		})
		if encodedFrame, err := masterInput.GetEncoder()(output); nil != err {
			return nil, errors.WithMessage(err, "Input设备Encode数据出错："+masterUuid)
		} else {
			return FramePacket(encodedFrame), nil
		}
	})
}

// 输出派发函数
// 根据Driver指定的目标输出设备地址，查找并处理数据包
func (p *Pipeline) deliverToOutput(uuid string, rawJSON *MessagePacket) (*MessagePacket, error) {
	if output, ok := p.uuidOutputs[uuid]; ok {
		encodedFrame, encErr := output.GetEncoder().Encode(rawJSON)
		if nil != encErr {
			return nil, errors.WithMessage(encErr, "设备Encode数据出错: "+uuid)
		}
		respFrame, err := output.Process(encodedFrame, p.context)
		if nil != err {
			return nil, errors.WithMessage(err, "Output设备处理出错: "+uuid)
		}
		if decodedMessage, decErr := output.GetDecoder().Decode(respFrame); nil != decErr {
			return nil, errors.WithMessage(encErr, "设备Decode数据出错: "+uuid)
		} else {
			return decodedMessage, nil
		}
	} else {
		return nil, errors.New("指定地址的Output设备不存在:" + uuid)
	}
}

// 处理拦截器过程
func (p *Pipeline) doInterceptor(session *session) {
	topic := session.Topic()
	p.context.OnIfLogV(func() {
		log.Debugf("正在Interceptor调度过程，Topic: %s", topic)
	})
	// 查找匹配的拦截器，按优先级排序并处理
	matches := make(InterceptorSlice, 0)
	for el := p.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		name := interceptor.GetName()
		if anyTopicMatches(interceptor.GetTopicExpr(), topic) {
			matches = append(matches, interceptor)
		} else {
			p.context.OnIfLogV(func() {
				log.Debugf("拦截器[未匹配], Name: %s, topic: %s", name, topic)
			})
		}
	}
	sort.Sort(matches)
	// 按排序结果顺序执行
	defer func() {
		p.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()

	for _, it := range matches {
		itName := it.GetName()
		start := time.Now()
		err := it.Handle(session.Attrs(), session.Topic(), session.Uuid(), session.GetInbound(), p.context)
		session.Attrs().Add("@Interceptor.Cost."+itName, time.Since(start))
		if err == nil {
			continue
		}
		if err == ErrInterceptorDropped {
			log.Debugf("拦截器[%s]中断事件： %s", itName, err.Error())
			session.WriteOutbound(NewMessagePacketFields(map[string]interface{}{
				"error": "INTERCEPTOR_DROPPED",
			}))
			return // 直接Return, 终止后续处理过程
		} else {
			p.failFastLogger(err, "拦截器发生错误:"+itName)
		}
	}
	// 后续处理
	session.Attrs().Add("@Interceptor.Cost.TOTAL", session.Since())
	// 1. Driver驱动处理
	// 2. Trigger触发处理
	p.driverChan <- session
	p.triggerChan <- session
}

// 处理驱动执行过程
func (p *Pipeline) doDriver(session *session) {
	topic := session.Topic()
	p.context.OnIfLogV(func() {
		log.Debugf("Driver调度，Topic: %s", topic)
	})
	// 查找匹配的用户驱动
	var driver Driver
	for el := p.drivers.Front(); el != nil; el = el.Next() {
		d := el.Value.(Driver)
		if anyTopicMatches(d.GetTopicExpr(), topic) {
			driver = d
			// 只匹配一个Driver
			break
		} else {
			p.context.OnIfLogV(func() {
				log.Debugf("用户驱动[未匹配], Driver: %s, topic: %s", d.GetName(), topic)
			})
		}
	}

	if nil != driver {
		driName := driver.GetName()
		// Driver 处理
		log.Debugf("用户驱动正在处理, Driver: %s, topic: %s", driName, topic)
		defer func() {
			p.checkRecover(recover(), "Driver-Goroutine内部错误:"+driName)
		}()
		start := time.Now()
		outbound, err := driver.Drive(session.Attrs(), session.Topic(), session.Uuid(), session.GetInbound(),
			OutputDeliverer(p.deliverToOutput), p.context)
		session.Attrs().Add("@Driver.Cost."+driName, time.Since(start))

		if nil != err {
			p.failFastLogger(err, "用户驱动发生错误:"+driName)
		} else {
			session.WriteOutbound(outbound)
		}
	} else {
		log.Debugf("未找到匹配的用户驱动, topic: %s", topic)
		session.WriteOutbound(NewMessagePacketFields(map[string]interface{}{
			"error": "DRIVER_NOT_FOUND",
		}))
	}
}

// 处理驱动执行过程
func (p *Pipeline) doTrigger(session *session) {
	topic := session.Topic()
	p.context.OnIfLogV(func() {
		log.Debugf("正在Trigger调度过程，Topic: %s", topic)
	})
	// 查找匹配的用户触发器, 并发处理
	for el := p.triggers.Front(); el != nil; el = el.Next() {
		trigger := el.Value.(Trigger)
		if anyTopicMatches(trigger.GetTopicExpr(), topic) {
			go func() {
				defer func() {
					p.checkRecover(recover(), "Trigger-Goroutine内部错误:"+trigger.GetName())
				}()
				// Driver 处理
				log.Debugf("用户触发器正在处理, Trigger: %s, topic: %s", trigger.GetName(), topic)
				err := trigger.Touch(session.Attrs(), session.Topic(), session.Uuid(), session.GetInbound(),
					OutputDeliverer(p.deliverToOutput), p.context)
				if nil != err {
					p.failFastLogger(err, "用户触发器发生错误:"+trigger.GetName())
				}
			}()
		} else {
			p.context.OnIfLogV(func() {
				log.Debugf("用户触发器[未匹配], Trigger: %s, topic: %s", trigger.GetName(), topic)
			})
		}
	}

}

func (p *Pipeline) checkDefTimeout(msg string, fn func(Context)) {
	p.context.CheckTimeout(msg, DefaultLifeCycleTimeout, func() {
		fn(p.context)
	})
}

func (p *Pipeline) checkRecover(r interface{}, msg string) {
	if nil == r {
		return
	}
	if err, ok := r.(error); ok {
		log.Errorw(msg, "error", err)
	}
	p.context.OnIfFailFast(func() {
		log.Fatal(r)
	})
}

func (p *Pipeline) failFastLogger(err error, msg string) {
	if p.context.IsFailFastEnabled() {
		log.Fatalw(msg, "error", err)
	} else {
		log.Errorw(msg, "error", err)
	}
}

func (p *Pipeline) newGeckoContext(config map[string]interface{}) *_GeckoContext {
	return &_GeckoContext{
		cfgGeckos:       utils.ToMap(config["GECKO"]),
		cfgGlobals:      utils.ToMap(config["GLOBALS"]),
		cfgInterceptors: utils.ToMap(config["INTERCEPTORS"]),
		cfgDrivers:      utils.ToMap(config["DRIVERS"]),
		cfgTriggers:     utils.ToMap(config["TRIGGERS"]),
		cfgOutputs:      utils.ToMap(config["OUTPUTS"]),
		cfgInputs:       utils.ToMap(config["INPUTS"]),
		cfgPlugins:      utils.ToMap(config["PLUGINS"]),
		cfgLogics:       utils.ToMap(config["LOGICS"]),
		scopedKV:        make(map[interface{}]interface{}),
		plugins:         p.plugins,
		interceptors:    p.interceptors,
		drivers:         p.drivers,
		triggers:        p.triggers,
		outputs:         p.outputs,
		inputs:          p.inputs,
	}
}

func (p *Pipeline) callStopFunc(component interface{}) {
	if stops, ok := component.(LifeCycle); ok {
		p.context.CheckTimeout(utils.GetClassName(component)+".停止", DefaultLifeCycleTimeout, func() {
			stops.OnStop(p.context)
		})
	}
}

func (p *Pipeline) callStartFunc(component interface{}) {
	if stops, ok := component.(LifeCycle); ok {
		p.context.CheckTimeout(utils.GetClassName(component)+".启动", DefaultLifeCycleTimeout, func() {
			stops.OnStart(p.context)
		})
	}
}

func anyTopicMatches(expected []*TopicExpr, topic string) bool {
	for _, t := range expected {
		if t.matches(topic) {
			return true
		}
	}
	return false
}
