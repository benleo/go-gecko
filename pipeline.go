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
	"sync"
	"syscall"
	"time"
)

////

// 默认组件生命周期超时时间：3秒
const DefaultLifeCycleTimeout = time.Second * 3

// Pipeline管理内部组件，处理事件。
type Pipeline struct {
	*Register
	geckoContext Context
	// 事件派发
	dispatcher     *Dispatcher
	dispatchCtx    context.Context
	dispatchCancel context.CancelFunc
}

// 初始化Pipeline
func (p *Pipeline) Init(config map[string]interface{}) {
	p.geckoContext = p.newGeckoContext(config)
	capacity := value.Of(p.geckoContext.gecko()["eventsCapacity"]).Int64OrDefault(100)
	if capacity <= 0 {
		capacity = 1
	}
	log.Infof("事件通道容量： %d", capacity)
	p.dispatcher = NewDispatcher(int(capacity))
	p.dispatcher.SetStartHandler(p.handleInterceptor)
	p.dispatcher.SetEndHandler(p.handleDriver)

	// 初始化组件：根据配置文件指定项目
	initFn := func(it Initial, args map[string]interface{}) {
		it.OnInit(args, p.geckoContext)
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
		it.Init(structConfig, p.geckoContext)
	}

	ctx := p.geckoContext.(*_GeckoContext)
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
	// Dispatcher运行
	go p.dispatcher.Serve(p.dispatchCtx)

	// Hook first
	utils.ForEach(p.startBeforeHooks, func(it interface{}) { it.(HookFunc)(p) })
	// Plugins
	utils.ForEach(p.plugins, p.callStartFunc)
	// Outputs
	utils.ForEach(p.outputs, p.callStartFunc)
	// Drivers
	utils.ForEach(p.drivers, p.callStartFunc)
	// Inputs
	utils.ForEach(p.inputs, p.callStartFunc)
	// Then, Serve inputs
	utils.ForEach(p.inputs, func(it interface{}) {
		input := it.(InputDevice)
		uuid := input.GetUuid()
		go func() {
			defer log.Debugf("InputDevice已经停止：%s", uuid)
			if err := input.Serve(p.geckoContext, p.newInputDeliverer(input)); nil != err {
				log.Errorw("InputDevice服务运行错误", "uuid", uuid,
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
	// Outputs
	utils.ForEach(p.outputs, p.callStopFunc)
	// Plugins
	utils.ForEach(p.plugins, p.callStopFunc)
	// Hook After
	utils.ForEach(p.stopAfterHooks, func(it interface{}) { it.(HookFunc)(p) })

	log.Info("Pipeline停止...OK")
	// 最终发起Dispatch关闭信息
	p.dispatchCancel()
}

// 等待系统终止信息
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
		session := &_EventSession{
			attributesMap: newMapAttributesWith(attributes),
			timestamp:     time.Now(),
			topic:         inputTopic,
			uuid:          inputUuid,
			inbound:       inputMessage,
			outbound:      NewMessagePacket(),
			completed:     make(chan *MessagePacket, 1),
		}
		p.dispatcher.StartC() <- session
		// 等待处理完成
		outputMessage := <-session.completed
		if nil == outputMessage {
			return nil, errors.New("Input设备发起Deliver请求必须返回结果数据")
		}
		if encodedFrame, err := masterInput.GetEncoder()(outputMessage); nil != err {
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
		respFrame, err := output.Process(encodedFrame, p.geckoContext)
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
func (p *Pipeline) handleInterceptor(session EventSession) {
	topic := session.Topic()
	p.geckoContext.OnIfLogV(func() {
		log.Debugf("正在Interceptor调度过程，Topic: %s", topic)
	})
	// 查找匹配的拦截器，按优先级排序并处理
	matches := make(InterceptorSlice, 0)
	for el := p.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		name := interceptor.GetName()
		if anyTopicMatches(interceptor.GetTopicExpr(), topic) {
			matches = append(matches, interceptor)
			log.Debugf("拦截器正在处理, Int: %s, topic: %s", name, topic)
		} else {
			p.geckoContext.OnIfLogV(func() {
				log.Debugf("拦截器[未匹配], Int: %s, topic: %s", name, topic)
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
		err := it.Handle(session, p.geckoContext)
		session.AddAttr("@Interceptor.Cost."+itName, session.Since())
		if err == nil {
			continue
		}
		if err == ErrInterceptorDropped {
			log.Debugf("拦截器[%s]中断事件： %s", itName, err.Error())
			session.Outbound().AddField("error", "InterceptorDropped")
			// 终止，输出处理
			session.AddAttr("@Interceptor.Cost.TOTAL", session.Since())
			p.output(session)
			return // 终止后续处理过程
		} else {
			p.failFastLogger(err, "拦截器发生错误:"+itName)
		}
	}
	// 后续处理
	session.AddAttr("@Interceptor.Cost.TOTAL", session.Since())
	p.dispatcher.EndC() <- session
}

// 处理驱动执行过程
func (p *Pipeline) handleDriver(session EventSession) {
	topic := session.Topic()
	p.geckoContext.OnIfLogV(func() {
		log.Debugf("正在Driver调度过程，Topic: %s", topic)
	})
	defer func() {
		p.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()

	// 输出处理
	defer func() {
		session.AddAttr("Driver.Cost.TOTAL", session.Since())
		p.output(session)
	}()

	// 查找匹配的用户驱动
	// Driver通常执行一些硬件驱动、远程事件等耗时操作，它们可以并行处理；
	// 使用WaitGroup归集后再输出；
	wg := new(sync.WaitGroup)
	for el := p.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		driName := driver.GetName()
		if anyTopicMatches(driver.GetTopicExpr(), topic) {
			wg.Add(1)
			go func() {
				defer wg.Done()
				log.Debugf("用户驱动正在处理, Driver: %s, topic: %s", driName, topic)
				if err := driver.Handle(session, OutputDeliverer(p.deliverToOutput), p.geckoContext); nil != err {
					p.failFastLogger(err, "用户驱动发生错误:"+driName)
				}
				session.AddAttr("Driver.Cost."+driName, session.Since())
			}()
		} else {
			p.geckoContext.OnIfLogV(func() {
				log.Debugf("用户驱动[未匹配], Driver: %s, topic: %s", driName, topic)
			})
		}
	}
	wg.Wait()
}

func (p *Pipeline) output(event EventSession) {
	p.geckoContext.OnIfLogV(func() {
		log.Debugf("正在Output调度过程，Topic: %s", event.Topic())
		for k, v := range event.Attrs() {
			log.Debugf("||-> SessionAttr: %s = %v", k, v)
		}
	})
	// 返回处理结果
	event.(*_EventSession).completed <- event.Outbound()
}

func (p *Pipeline) checkDefTimeout(msg string, act func(Context)) {
	p.geckoContext.CheckTimeout(msg, DefaultLifeCycleTimeout, func() {
		act(p.geckoContext)
	})
}

func (p *Pipeline) checkRecover(r interface{}, msg string) {
	if nil == r {
		return
	}
	if err, ok := r.(error); ok {
		log.Errorw(msg, "error", err)
	}
	p.geckoContext.OnIfFailFast(func() {
		log.Fatal(r)
	})
}

func (p *Pipeline) failFastLogger(err error, msg string) {
	if p.geckoContext.IsFailFastEnabled() {
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
		cfgOutputs:      utils.ToMap(config["OUTPUTS"]),
		cfgInputs:       utils.ToMap(config["INPUTS"]),
		cfgPlugins:      utils.ToMap(config["PLUGINS"]),
		cfgLogics:       utils.ToMap(config["LOGICS"]),
		scopedKV:        make(map[interface{}]interface{}),
		plugins:         p.plugins,
		interceptors:    p.interceptors,
		drivers:         p.drivers,
		outputs:         p.outputs,
		inputs:          p.inputs,
	}
}

func (p *Pipeline) callStopFunc(component interface{}) {
	if stops, ok := component.(LifeCycle); ok {
		p.geckoContext.CheckTimeout(utils.GetClassName(component)+".Stop", DefaultLifeCycleTimeout, func() {
			stops.OnStop(p.geckoContext)
		})
	}
}

func (p *Pipeline) callStartFunc(component interface{}) {
	if stops, ok := component.(LifeCycle); ok {
		p.geckoContext.CheckTimeout(utils.GetClassName(component)+".Start", DefaultLifeCycleTimeout, func() {
			stops.OnStart(p.geckoContext)
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
