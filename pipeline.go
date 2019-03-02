package gecko

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/pkg/errors"
	"github.com/yoojia/go-gecko/utils"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"sync"
	"syscall"
	"time"
)

////

// 默认组件生命周期超时时间：3秒
const DefaultLifeCycleTimeout = time.Second * 3

var gSharedPipeline = &Pipeline{
	Registration: prepare(),
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
	*Registration
	ctx Context
	// 事件派发
	dispatcher *Dispatcher
	// Pipeline关闭的信号控制
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
}

// 初始化Pipeline
func (p *Pipeline) Init(config *cfg.Config) {
	zlog := ZapSugarLogger
	geckoCtx := newGeckoContext(config)
	p.ctx = geckoCtx
	gecko := p.ctx.gecko()
	capacity := gecko.GetInt64OrDefault("eventsCapacity", 8)
	zlog.Infof("事件通道容量： %d", capacity)
	p.dispatcher = NewDispatcher(int(capacity))
	p.dispatcher.SetStartHandler(p.handleInterceptor)
	p.dispatcher.SetEndHandler(p.handleDriver)
	go p.dispatcher.Serve(p.shutdownCtx)

	// 初始化组件：根据配置文件指定项目
	initWithContext := func(it Initialize, args *cfg.Config) {
		it.OnInit(args, p.ctx)
	}
	if !p.registerIfHit(geckoCtx.plugins, initWithContext) {
		zlog.Warn("警告：未配置任何[Plugin]组件")
	}
	if !p.registerIfHit(geckoCtx.outputs, initWithContext) {
		zlog.Fatal("严重：未配置任何[OutputDevice]组件")
	}
	if !p.registerIfHit(geckoCtx.interceptors, initWithContext) {
		zlog.Warn("警告：未配置任何[Interceptor]组件")
	}
	if !p.registerIfHit(geckoCtx.drivers, initWithContext) {
		zlog.Warn("警告：未配置任何[Driver]组件")
	}
	if !p.registerIfHit(geckoCtx.inputs, initWithContext) {
		zlog.Fatal("严重：未配置任何[InputDevice]组件")
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
	// Plugins
	utils.ForEach(p.plugins, func(it interface{}) {
		p.checkDefTimeout("Plugin.Start", it.(Plugin).OnStart)
	})
	// Outputs
	utils.ForEach(p.outputs, func(it interface{}) {
		p.ctx.CheckTimeout("Output.Start", DefaultLifeCycleTimeout, func() {
			it.(OutputDevice).OnStart(p.ctx)
		})
	})
	// Drivers
	utils.ForEach(p.drivers, func(it interface{}) {
		p.checkDefTimeout("Driver.Start", it.(Driver).OnStart)
	})
	// Inputs
	utils.ForEach(p.inputs, func(it interface{}) {
		p.ctx.CheckTimeout("Trigger.Start", DefaultLifeCycleTimeout, func() {
			it.(InputDevice).OnStart(p.ctx)
		})
	})
	// Input Serve Last
	utils.ForEach(p.inputs, func(it interface{}) {
		device := it.(InputDevice)
		deliverer := p.newInputDeliverer(device)
		go func() {
			if err := device.Serve(p.ctx, deliverer); nil != err {
				zlog.Errorw("InputDevice服务运行错误：", "error", err, "class", utils.GetClassName(device))
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
	// Inputs
	utils.ForEach(p.inputs, func(it interface{}) {
		p.ctx.CheckTimeout("Input.Stop", DefaultLifeCycleTimeout, func() {
			it.(InputDevice).OnStop(p.ctx)
		})
	})
	// Drivers
	utils.ForEach(p.drivers, func(it interface{}) {
		p.checkDefTimeout("Driver.Stop", it.(Driver).OnStop)
	})
	// Outputs
	utils.ForEach(p.outputs, func(it interface{}) {
		p.ctx.CheckTimeout("Output.Stop", DefaultLifeCycleTimeout, func() {
			it.(OutputDevice).OnStop(p.ctx)
		})
	})
	// Plugins
	utils.ForEach(p.plugins, func(it interface{}) {
		p.checkDefTimeout("Plugin.Stop", it.(Plugin).OnStop)
	})
}

// 等待系统终止信息
func (p *Pipeline) AwaitTermination() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	ZapSugarLogger.Info("接收到系统停止信号")
}

// 准备运行环境，初始化相关组件
func (p *Pipeline) prepareEnv() {
	p.shutdownCtx, p.shutdownFunc = context.WithCancel(context.Background())
}

// 输出派发函数
// 根据Driver指定的目标输出设备地址，查找并处理数据包
func (p *Pipeline) deliverToOutput(address string, broadcast bool, data JSONPacket) (JSONPacket, error) {
	// 广播给相同组地址的设备
	if broadcast {
		zlog := ZapSugarLogger
		for addr, device := range p.outputsMap {
			// 忽略GroupAddress不匹配的设备
			if address != device.GetAddress().Group {
				continue
			}
			frame, encErr := device.GetEncoder().Encode(data)
			if nil != encErr {
				return nil, errors.WithMessage(encErr, "设备Encode数据出错: "+addr)
			}
			if ret, procErr := device.Process(frame, p.ctx); nil != procErr {
				zlog.Errorw("OutputDevice处理广播事件发生错误", "addr", addr, "error", procErr)
				return nil, errors.WithMessage(procErr, "Output broadcast of device: "+addr)
			} else {
				if nil != ret {
					if json, decErr := device.GetDecoder().Decode(ret); nil != decErr {
						return nil, errors.WithMessage(encErr, "设备Decode数据出错: "+addr)
					} else {
						zlog.Debugf("OutputDevice[%s]返回响应： %s", addr, json)
					}
				}
			}
		}
		return nil, nil
	} else {
		// 发送给精确地址的设备
		if device, ok := p.outputsMap[address]; ok {
			frame, encErr := device.GetEncoder().Encode(data)
			if nil != encErr {
				return nil, errors.WithMessage(encErr, "设备Encode数据出错: "+address)
			}
			ret, procErr := device.Process(frame, p.ctx)
			if nil != procErr {
				return nil, errors.WithMessage(procErr, "Output设备处理出错: "+address)
			}
			if json, decErr := device.GetDecoder().Decode(ret); nil != decErr {
				return nil, errors.WithMessage(encErr, "设备Decode数据出错: "+address)
			} else {
				return json, nil
			}
		} else {
			return nil, errors.New("指定地址的Output设备不存在:" + address)
		}
	}
}

// 创建InputDeliverer函数
func (p *Pipeline) newInputDeliverer(device InputDevice) InputDeliverer {
	return InputDeliverer(func(topic string, frame FramePacket) (FramePacket, error) {
		// 从Input设备中读取Decode数据
		decoder := device.GetDecoder()
		input, err := decoder(frame.Data())
		if nil != err {
			return nil, errors.WithMessage(err, "Input设备Decode数据出错："+device.GetAddress().UUID)
		}
		output := make(chan JSONPacket, 1)
		// 发送到Dispatcher调度处理
		p.dispatcher.StartC() <- &_GeckoEventContext{
			timestamp:  time.Now(),
			attributes: make(map[string]interface{}),
			attrLock:   new(sync.RWMutex),
			topic:      topic,
			inbound: &Message{
				Topic: topic,
				Data:  input,
			},
			outbound: &Message{
				Topic: topic,
				Data:  make(map[string]interface{}),
			},
			completedNotifier: output,
		}
		// Encode返回到Input设备
		encoder := device.GetEncoder()
		if bytes, err := encoder(<-output); nil != err {
			return nil, errors.WithMessage(err, "Input设备Encode数据出错："+device.GetAddress().UUID)
		} else {
			return NewFramePacket(bytes), nil
		}
	})
}

// 处理拦截器过程
func (p *Pipeline) handleInterceptor(session EventSession) {
	zlog := ZapSugarLogger
	p.ctx.OnIfLogV(func() {
		zlog.Debugf("Interceptor调度处理，Topic: %s", session.Topic())
	})
	defer func() {
		p.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	// 查找匹配的拦截器，按优先级排序并处理
	matches := make(InterceptorSlice, 0)
	for el := p.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		match := anyTopicMatches(interceptor.GetTopicExpr(), session.Topic())
		p.ctx.OnIfLogV(func() {
			zlog.Debugf("拦截器调度： interceptor[%s], topic: %s, Matches: %s",
				utils.GetClassName(interceptor),
				session.Topic(),
				strconv.FormatBool(match))
		})
		if match {
			matches = append(matches, interceptor)
		}
	}
	sort.Sort(matches)
	// 按排序结果顺序执行
	for _, it := range matches {
		err := it.Handle(session, p.ctx)
		if err == nil {
			continue
		}
		if err == ErrInterceptorDropped {
			zlog.Debugf("拦截器中断事件： %s", err.Error())
			session.Outbound().AddDataField("error", "InterceptorDropped")
			// 终止，输出处理
			session.AddAttribute("Since@Interceptor", session.Since())
			p.output(session)
			return
		} else {
			p.failFastLogger(err, "拦截器发生错误")
		}
	}
	// 继续
	session.AddAttribute("Since@Interceptor", session.Since())
	p.dispatcher.EndC() <- session
}

// 处理驱动执行过程
func (p *Pipeline) handleDriver(session EventSession) {
	p.ctx.OnIfLogV(func() {
		ZapSugarLogger.Debugf("Driver调度处理，Topic: %s", session.Topic())
	})
	defer func() {
		p.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()

	// 查找匹配的用户驱动，并处理
	for el := p.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		match := anyTopicMatches(driver.GetTopicExpr(), session.Topic())
		p.ctx.OnIfLogV(func() {
			ZapSugarLogger.Debugf("用户驱动处理： driver[%s], topic: %s, Matches: %s",
				utils.GetClassName(driver),
				session.Topic(),
				strconv.FormatBool(match))
		})
		if match {
			err := driver.Handle(session, OutputDeliverer(p.deliverToOutput), p.ctx)
			if nil != err {
				p.failFastLogger(err, "用户驱动发生错误")
			}
		}
	}
	// 输出处理
	session.AddAttribute("Since@Driver", session.Since())
	p.output(session)
}

func (p *Pipeline) output(event EventSession) {
	p.ctx.OnIfLogV(func() {
		zlog := ZapSugarLogger
		zlog.Debugf("Output调度处理，Topic: %s", event.Topic())
		event.Attributes().ForEach(func(k string, v interface{}) {
			zlog.Debugf("SessionAttr: %s = %v", k, v)
		})
	})
	defer func() {
		p.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	event.(*_GeckoEventContext).completedNotifier <- event.Outbound().Data
}

func (p *Pipeline) checkDefTimeout(msg string, act func(Context)) {
	p.ctx.CheckTimeout(msg, DefaultLifeCycleTimeout, func() {
		act(p.ctx)
	})
}

func (p *Pipeline) checkRecover(r interface{}, msg string) {
	if nil != r {
		zlog := ZapSugarLogger
		if err, ok := r.(error); ok {
			zlog.Errorw(msg, "error", err)
		}
		p.ctx.OnIfFailFast(func() {
			zlog.Fatal(r)
		})
	}
}

func (p *Pipeline) failFastLogger(err error, msg string) {
	zlog := ZapSugarLogger
	if p.ctx.IsFailFastEnabled() {
		zlog.Fatalw(msg, "error", err)
	} else {
		zlog.Errorw(msg, "error", err)
	}
}

func newGeckoContext(config *cfg.Config) *_GeckoContext {
	return &_GeckoContext{
		geckos:       config.MustConfig("GECKO"),
		globals:      config.MustConfig("GLOBALS"),
		interceptors: config.MustConfig("INTERCEPTORS"),
		drivers:      config.MustConfig("DRIVERS"),
		outputs:      config.MustConfig("OUTPUTS"),
		inputs:       config.MustConfig("INPUTS"),
		plugins:      config.MustConfig("PLUGINS"),
		scopedKV:     make(map[interface{}]interface{}),
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
