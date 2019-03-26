package gecko

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/pkg/errors"
	"github.com/yoojia/go-gecko/structs"
	"github.com/yoojia/go-gecko/utils"
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

var gSharedPipeline = &Pipeline{
	Register: prepare(),
}
var gPrepareEnv = new(sync.Once)

// 全局Pipeline对象
func SharedPipeline() *Pipeline {
	gPrepareEnv.Do(func() {
		gSharedPipeline.prepareEnv()
	})
	return gSharedPipeline
}

// Pipeline管理内部组件，处理事件。
type Pipeline struct {
	*Register
	context Context
	// 事件派发
	dispatcher *Dispatcher
	// Pipeline关闭的信号控制
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
}

// 初始化Pipeline
func (p *Pipeline) Init(config *cfg.Config) {
	zlog := ZapSugarLogger
	ctx := p.newGeckoContext(config)
	p.context = ctx
	gecko := p.context.gecko()
	capacity := gecko.GetInt64OrDefault("eventsCapacity", 8)
	zlog.Infof("事件通道容量： %d", capacity)
	p.dispatcher = NewDispatcher(int(capacity))
	p.dispatcher.SetStartHandler(p.handleInterceptor)
	p.dispatcher.SetEndHandler(p.handleDriver)

	go p.dispatcher.Serve(p.shutdownCtx)

	// 初始化组件：根据配置文件指定项目
	initFn := func(it Initial, args *cfg.Config) {
		it.OnInit(args, p.context)
	}
	// 使用结构化的参数来初始化
	structInitFn := func(it StructuredInitial, args *cfg.Config) {
		structConfig := it.StructuredConfig()
		m2sDecoder, err := structs.NewDecoder(&structs.DecoderConfig{
			TagName: "toml",
			Result:  structConfig,
		})
		if nil != err {
			zlog.Panic("无法创建Map2Struct解码器", err)
		}
		if err := m2sDecoder.Decode(args.RefMap()); nil != err {
			zlog.Panic("Map2Struct解码出错", err)
		}
		it.Init(structConfig, p.context)
	}

	if ctx.cfgPlugins.IsEmpty() {
		zlog.Warn("警告：未配置任何[Plugin]组件")
	} else {
		p.register(ctx.cfgPlugins, initFn, structInitFn)
	}
	if ctx.cfgOutputs.IsEmpty() {
		zlog.Fatal("严重：未配置任何[OutputDevice]组件")
	} else {
		p.register(ctx.cfgOutputs, initFn, structInitFn)
	}
	if ctx.cfgInterceptors.IsEmpty() {
		zlog.Warn("警告：未配置任何[Interceptor]组件")
	} else {
		p.register(ctx.cfgInterceptors, initFn, structInitFn)
	}
	if ctx.cfgDrivers.IsEmpty() {
		zlog.Warn("警告：未配置任何[Driver]组件")
	} else {
		p.register(ctx.cfgDrivers, initFn, structInitFn)
	}
	if ctx.cfgInputs.IsEmpty() {
		zlog.Fatal("严重：未配置任何[InputDevice]组件")
	} else {
		p.register(ctx.cfgInputs, initFn, structInitFn)
	}
	if !ctx.cfgLogics.IsEmpty() {
		p.register(ctx.cfgLogics, initFn, structInitFn)
	} else {
		zlog.Warn("警告：未配置任何[LogicDevice]组件")
	}
	// show
	p.showBundles()
}

// 启动Pipeline
func (p *Pipeline) Start() {
	zlog := ZapSugarLogger
	zlog.Info("Pipeline启动...")
	// Hook first
	utils.ForEach(p.startBeforeHooks, func(it interface{}) {
		it.(HookFunc)(p)
	})
	defer func() {
		utils.ForEach(p.startAfterHooks, func(it interface{}) {
			it.(HookFunc)(p)
		})
		zlog.Info("Pipeline启动...OK")
	}()

	startFn := func(component interface{}) {
		if starts, ok := component.(LifeCycle); ok {
			p.context.CheckTimeout(utils.GetClassName(component)+".Start", DefaultLifeCycleTimeout, func() {
				starts.OnStart(p.context)
			})
		}
	}

	// Plugins
	utils.ForEach(p.plugins, startFn)
	// Outputs
	utils.ForEach(p.outputs, startFn)
	// Drivers
	utils.ForEach(p.drivers, startFn)
	// Inputs
	utils.ForEach(p.inputs, startFn)
	// Then, Serve inputs
	utils.ForEach(p.inputs, func(it interface{}) {
		input := it.(InputDevice)
		deliverer := p.newInputDeliverer(input)
		go func() {
			uuid := input.GetUuid()
			defer zlog.Debugf("InputDevice已经停止：%s", uuid)
			err := input.Serve(p.context, deliverer)
			if nil != err {
				zlog.Errorw("InputDevice服务运行错误",
					"uuid", uuid,
					"error", err,
					"class", utils.GetClassName(input))
			}
		}()
	})
}

// 停止Pipeline
func (p *Pipeline) Stop() {
	zlog := ZapSugarLogger
	zlog.Info("Pipeline停止...")
	// Hook first
	utils.ForEach(p.stopBeforeHooks, func(it interface{}) {
		it.(HookFunc)(p)
	})
	defer func() {
		utils.ForEach(p.stopAfterHooks, func(it interface{}) {
			it.(HookFunc)(p)
		})
		// 最终发起关闭信息
		p.shutdownFunc()
		zlog.Info("Pipeline停止...OK")
	}()

	stopFn := func(component interface{}) {
		if stops, ok := component.(LifeCycle); ok {
			p.context.CheckTimeout(utils.GetClassName(component)+".Stop", DefaultLifeCycleTimeout, func() {
				stops.OnStop(p.context)
			})
		}
	}
	// Inputs
	utils.ForEach(p.inputs, stopFn)
	// Drivers
	utils.ForEach(p.drivers, stopFn)
	// Outputs
	utils.ForEach(p.outputs, stopFn)
	// Plugins
	utils.ForEach(p.plugins, stopFn)
}

// 等待系统终止信息
func (p *Pipeline) AwaitTermination() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ZapSugarLogger.Info("接收到系统停止信号")
}

func (p *Pipeline) init() {

}

// 准备运行环境，初始化相关组件
func (p *Pipeline) prepareEnv() {
	p.shutdownCtx, p.shutdownFunc = context.WithCancel(context.Background())
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
		session := &_EventSessionImpl{
			timestamp: time.Now(),
			attrs:     attributes,
			topic:     inputTopic,
			uuid:      inputUuid,
			inbound:   inputMessage,
			outbound:  NewMessagePacket(),
			completed: make(chan *MessagePacket, 1),
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
func (p *Pipeline) handleInterceptor(session EventSession) {
	zlog := ZapSugarLogger
	topic := session.Topic()
	p.context.OnIfLogV(func() {
		zlog.Debugf("正在Interceptor调度过程，Topic: %s", topic)
	})
	defer func() {
		p.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	// 查找匹配的拦截器，按优先级排序并处理
	matches := make(InterceptorSlice, 0)
	for el := p.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		name := interceptor.GetName()
		if anyTopicMatches(interceptor.GetTopicExpr(), topic) {
			matches = append(matches, interceptor)
			zlog.Debugf("拦截器正在处理, Int: %s, topic: %s", name, topic)
		} else {
			p.context.OnIfLogV(func() {
				zlog.Debugf("拦截器[未匹配], Int: %s, topic: %s", name, topic)
			})
		}
	}
	sort.Sort(matches)
	// 按排序结果顺序执行
	for _, it := range matches {
		err := it.Handle(session, p.context)
		if err == nil {
			continue
		}
		if err == ErrInterceptorDropped {
			zlog.Debugf("拦截器[%s]中断事件： %s", it.GetName(), err.Error())
			session.Outbound().AddField("error", "InterceptorDropped")
			// 终止，输出处理
			session.AddAttr("拦截过程用时", session.Since())
			p.output(session)
			return
		} else {
			p.failFastLogger(err, "拦截器发生错误:"+it.GetName())
		}
	}
	// 继续
	session.AddAttr("拦截过程用时", session.Since())
	p.dispatcher.EndC() <- session
}

// 处理驱动执行过程
func (p *Pipeline) handleDriver(session EventSession) {
	zlog := ZapSugarLogger
	topic := session.Topic()
	p.context.OnIfLogV(func() {
		zlog.Debugf("正在Driver调度过程，Topic: %s", topic)
	})
	defer func() {
		p.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()

	defer func() {
		// 输出处理
		session.AddAttr("驱动过程用时", session.Since())
		p.output(session)
	}()
	// 查找匹配的用户驱动
	for el := p.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		name := driver.GetName()
		if anyTopicMatches(driver.GetTopicExpr(), topic) {
			zlog.Debugf("用户驱动正在处理, Driver: %s, topic: %s", name, topic)
			// Drivers不要并行处理：每个输入消息本身已被协程异步调度；在一个Session周期内，它的执行过程应当是串行的。
			// 如果Driver内部存在与Session无关的异步操作，可以由Driver内部自身实现。
			if err := driver.Handle(session, OutputDeliverer(p.deliverToOutput), p.context); nil != err {
				p.failFastLogger(err, "用户驱动发生错误:"+name)
			}
		} else {
			p.context.OnIfLogV(func() {
				zlog.Debugf("用户驱动[未匹配], Driver: %s, topic: %s", name, topic)
			})
		}
	}
}

func (p *Pipeline) output(event EventSession) {
	p.context.OnIfLogV(func() {
		zlog := ZapSugarLogger
		zlog.Debugf("正在Output调度过程，Topic: %s", event.Topic())
		event.Attrs().ForEach(func(k string, v interface{}) {
			zlog.Debugf("SessionAttr: %s = %v", k, v)
		})
	})
	defer func() {
		p.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	// 返回处理结果
	event.(*_EventSessionImpl).completed <- event.Outbound()
}

func (p *Pipeline) checkDefTimeout(msg string, act func(Context)) {
	p.context.CheckTimeout(msg, DefaultLifeCycleTimeout, func() {
		act(p.context)
	})
}

func (p *Pipeline) checkRecover(r interface{}, msg string) {
	if nil == r {
		return
	}
	zlog := ZapSugarLogger
	if err, ok := r.(error); ok {
		zlog.Errorw(msg, "error", err)
	}
	p.context.OnIfFailFast(func() {
		zlog.Fatal(r)
	})
}

func (p *Pipeline) failFastLogger(err error, msg string) {
	zlog := ZapSugarLogger
	if p.context.IsFailFastEnabled() {
		zlog.Fatalw(msg, "error", err)
	} else {
		zlog.Errorw(msg, "error", err)
	}
}

func (p *Pipeline) newGeckoContext(config *cfg.Config) *_GeckoContext {
	return &_GeckoContext{
		cfgGeckos:       config.MustConfig("GECKO"),
		cfgGlobals:      config.MustConfig("GLOBALS"),
		cfgInterceptors: config.MustConfig("INTERCEPTORS"),
		cfgDrivers:      config.MustConfig("DRIVERS"),
		cfgOutputs:      config.MustConfig("OUTPUTS"),
		cfgInputs:       config.MustConfig("INPUTS"),
		cfgPlugins:      config.MustConfig("PLUGINS"),
		cfgLogics:       config.MustConfig("LOGICS"),
		scopedKV:        make(map[interface{}]interface{}),
		plugins:         p.plugins,
		interceptors:    p.interceptors,
		drivers:         p.drivers,
		outputs:         p.outputs,
		inputs:          p.inputs,
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
