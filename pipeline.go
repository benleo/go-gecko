package gecko

import (
	"context"
	"errors"
	"github.com/parkingwang/go-conf"
	"github.com/yoojia/go-gecko/x"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
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
func (pl *Pipeline) Init(config *cfg.Config) {
	geckoCtx := newGeckoContext(config)
	pl.ctx = geckoCtx
	gecko := pl.ctx.gecko()
	capacity := gecko.GetInt64OrDefault("eventsCapacity", 8)
	pl.zap.Infof("事件通道容量： %d", capacity)
	pl.dispatcher = NewDispatcher(int(capacity))
	pl.dispatcher.SetStartHandler(pl.handleInterceptor)
	pl.dispatcher.SetEndHandler(pl.handleDriver)
	go pl.dispatcher.Serve(pl.shutdownCtx)

	// 初始化组件：根据配置文件指定项目
	initWithContext := func(it Initialize, args *cfg.Config) {
		it.OnInit(args, pl.ctx)
	}
	if !pl.registerIfHit(geckoCtx.plugins, initWithContext) {
		pl.zap.Warn("警告：未配置任何[Plugin]组件")
	}
	if !pl.registerIfHit(geckoCtx.outputs, initWithContext) {
		pl.zap.Panic("严重：未配置任何[OutputDevice]组件")
	}
	if !pl.registerIfHit(geckoCtx.interceptors, initWithContext) {
		pl.zap.Warn("警告：未配置任何[Interceptor]组件")
	}
	if !pl.registerIfHit(geckoCtx.drivers, initWithContext) {
		pl.zap.Warn("警告：未配置任何[Driver]组件")
	}
	if !pl.registerIfHit(geckoCtx.inputs, initWithContext) {
		pl.zap.Panic("严重：未配置任何[InputDevice]组件")
	}
	// show
	pl.showBundles()
}

// 启动Pipeline
func (pl *Pipeline) Start() {
	pl.zap.Infof("Pipeline启动...")
	// Hook first
	x.ForEach(pl.startBeforeHooks, func(it interface{}) {
		it.(HookFunc)(pl)
	})
	defer func() {
		x.ForEach(pl.startAfterHooks, func(it interface{}) {
			it.(HookFunc)(pl)
		})
		pl.zap.Infof("Pipeline启动...OK")
	}()
	// Plugins
	x.ForEach(pl.plugins, func(it interface{}) {
		pl.checkDefTimeout("Plugin.Start", it.(Plugin).OnStart)
	})
	// Outputs
	x.ForEach(pl.outputs, func(it interface{}) {
		pl.ctx.CheckTimeout("Output.Start", DefaultLifeCycleTimeout, func() {
			it.(OutputDevice).OnStart(pl.ctx)
		})
	})
	// Drivers
	x.ForEach(pl.drivers, func(it interface{}) {
		pl.checkDefTimeout("Driver.Start", it.(Driver).OnStart)
	})
	// Inputs
	x.ForEach(pl.inputs, func(it interface{}) {
		pl.ctx.CheckTimeout("Trigger.Start", DefaultLifeCycleTimeout, func() {
			it.(InputDevice).OnStart(pl.ctx)
		})
	})
	// Input Serve Last
	x.ForEach(pl.inputs, func(it interface{}) {
		device := it.(InputDevice)
		deliverer := pl.newInputDeliverer(device)
		go func() {
			if err := device.Serve(pl.ctx, deliverer); nil != err {
				pl.zap.Errorw("InputDevice服务运行错误：", "class", x.SimpleClassName(device))
			}
		}()
	})
}

// 停止Pipeline
func (pl *Pipeline) Stop() {
	pl.zap.Infof("Pipeline停止...")
	// Hook first
	x.ForEach(pl.stopBeforeHooks, func(it interface{}) {
		it.(HookFunc)(pl)
	})
	defer func() {
		x.ForEach(pl.stopAfterHooks, func(it interface{}) {
			it.(HookFunc)(pl)
		})
		// 最终发起关闭信息
		pl.shutdownFunc()
		pl.zap.Infof("Pipeline停止...OK")
	}()
	// Inputs
	x.ForEach(pl.inputs, func(it interface{}) {
		pl.ctx.CheckTimeout("Input.Stop", DefaultLifeCycleTimeout, func() {
			it.(InputDevice).OnStop(pl.ctx)
		})
	})
	// Drivers
	x.ForEach(pl.drivers, func(it interface{}) {
		pl.checkDefTimeout("Driver.Stop", it.(Driver).OnStop)
	})
	// Outputs
	x.ForEach(pl.outputs, func(it interface{}) {
		pl.ctx.CheckTimeout("Output.Stop", DefaultLifeCycleTimeout, func() {
			it.(OutputDevice).OnStop(pl.ctx)
		})
	})
	// Plugins
	x.ForEach(pl.plugins, func(it interface{}) {
		pl.checkDefTimeout("Plugin.Stop", it.(Plugin).OnStop)
	})
}

// 等待系统终止信息
func (pl *Pipeline) AwaitTermination() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	pl.zap.Warnf("接收到系统停止信号")
}

// 准备运行环境，初始化相关组件
func (pl *Pipeline) prepareEnv() {
	pl.shutdownCtx, pl.shutdownFunc = context.WithCancel(context.Background())
}

// 输出派发函数
// 根据Driver指定的目标输出设备地址，查找并处理数据包
func (pl *Pipeline) sendToOutput(address DeviceAddress, data PacketMap) (PacketMap, error) {
	if address.IsValid() {
		devAddr := address.GetUnionAddress()
		if device, ok := pl.namedOutputs[devAddr]; ok {
			frame, dErr := device.GetEncoder().Encode(data)
			if nil != dErr {
				return nil, dErr
			}
			ret, pErr := device.Process(frame, pl.ctx)
			if nil != pErr {
				return nil, pErr
			}
			if pm, err := device.GetDecoder().Decode(ret); nil != err {
				return nil, err
			} else {
				return pm, nil
			}
		} else {
			return nil, errors.New("指定地址的设备不存在:" + devAddr)
		}
	} else /*is Group Address*/ {
		groupAddr := address.Group
		for addr, device := range pl.namedOutputs {
			// 忽略GroupAddress不匹配的设备
			if !strings.HasPrefix(addr, groupAddr) {
				continue
			}
			frame, dErr := device.GetEncoder().Encode(data)
			if nil != dErr {
				return nil, dErr
			}
			if ret, err := device.Process(frame, pl.ctx); nil != err {
				pl.zap.Errorw("OutputDevice处理广播事件发生错误", "addr", addr, "error", err)
				return nil, err
			} else {
				if nil != ret {
					if pm, err := device.GetDecoder().Decode(ret); nil == err {
						pl.zap.Debugf("OutputDevice[%s]返回响应： %s", addr, pm)
					}
				}
			}
		}
		return nil, nil
	}
}

// 创建InputDeliverer函数
func (pl *Pipeline) newInputDeliverer(device InputDevice) InputDeliverer {
	return InputDeliverer(func(topic string, frame PacketFrame) (PacketFrame, error) {
		// 解码
		decoder := device.GetDecoder()
		inData, err := decoder(frame.Data())
		if nil != err {
			pl.zap.Errorw("InputDevice解码/Decode错误", "class", x.SimpleClassName(device))
			return nil, err
		}
		awaitResult := make(chan PacketMap, 1)
		// 处理
		pl.dispatcher.StartC() <- &_GeckoSession{
			timestamp:  time.Now(),
			attributes: make(map[string]interface{}),
			attrLock:   new(sync.RWMutex),
			topic:      topic,
			inbound: &Inbound{
				Topic: topic,
				Data:  inData,
			},
			outbound: &Outbound{
				Topic: topic,
				Data:  make(map[string]interface{}),
			},
			notifyCompletedFunc: func(data PacketMap) {
				// 通过 notifyCompletedFunc 返回处理结果
				awaitResult <- data
			},
		}
		// 编码
		encoder := device.GetEncoder()
		outData := <-awaitResult
		if bytes, err := encoder(outData); nil != err {
			pl.zap.Errorw("InputDevice编码/Encode错误", "class", x.SimpleClassName(device))
			return nil, err
		} else {
			return NewPackFrame(bytes), nil
		}
	})
}

// 处理拦截器过程
func (pl *Pipeline) handleInterceptor(session Session) {
	pl.ctx.OnIfLogV(func() {
		pl.zap.Debugf("Interceptor调度处理，Topic: %s", session.Topic())
	})
	defer func() {
		pl.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	// 查找匹配的拦截器，按优先级排序并处理
	matches := make(InterceptorSlice, 0)
	for el := pl.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		match := anyTopicMatches(interceptor.GetTopicExpr(), session.Topic())
		pl.ctx.OnIfLogV(func() {
			pl.zap.Debugf("拦截器调度： interceptor[%s], topic: %s, Matches: %s",
				x.SimpleClassName(interceptor),
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
		err := it.Handle(session, pl.ctx)
		if err == nil {
			continue
		}
		if err == ErrInterceptorDropped {
			pl.zap.Debugf("拦截器中断事件： %s", err.Error())
			session.Outbound().AddDataField("error", "InterceptorDropped")
			// 终止，输出处理
			session.AddAttribute("Since@Interceptor", session.Since())
			pl.output(session)
			return
		} else {
			pl.failFastLogger(err, "拦截器发生错误")
		}
	}
	// 继续
	session.AddAttribute("Since@Interceptor", session.Since())
	pl.dispatcher.EndC() <- session
}

// 处理驱动执行过程
func (pl *Pipeline) handleDriver(session Session) {
	pl.ctx.OnIfLogV(func() {
		pl.zap.Debugf("Driver调度处理，Topic: %s", session.Topic())
	})
	defer func() {
		pl.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()

	// 查找匹配的用户驱动，并处理
	for el := pl.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		match := anyTopicMatches(driver.GetTopicExpr(), session.Topic())
		pl.ctx.OnIfLogV(func() {
			pl.zap.Debugf("用户驱动处理： driver[%s], topic: %s, Matches: %s",
				x.SimpleClassName(driver),
				session.Topic(),
				strconv.FormatBool(match))
		})
		if match {
			err := driver.Handle(session, OutputDeliverer(pl.sendToOutput), pl.ctx)
			if nil != err {
				pl.failFastLogger(err, "用户驱动发生错误")
			}
		}
	}
	// 输出处理
	session.AddAttribute("Since@Driver", session.Since())
	pl.output(session)
}

func (pl *Pipeline) output(session Session) {
	pl.ctx.OnIfLogV(func() {
		pl.zap.Debugf("Output调度处理，Topic: %s", session.Topic())
		session.Attributes().ForEach(func(k string, v interface{}) {
			pl.zap.Debugf("SessionAttr: %s = %v", k, v)
		})
	})
	defer func() {
		pl.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	session.(*_GeckoSession).notifyCompletedFunc(session.Outbound().Data)
}

func (pl *Pipeline) checkDefTimeout(msg string, act func(Context)) {
	pl.ctx.CheckTimeout(msg, DefaultLifeCycleTimeout, func() {
		act(pl.ctx)
	})
}

func (pl *Pipeline) checkRecover(r interface{}, msg string) {
	if nil != r {
		if err, ok := r.(error); ok {
			pl.zap.Errorw(msg, "error", err)
		}
		pl.ctx.OnIfFailFast(func() {
			panic(r)
		})
	}
}

func (pl *Pipeline) failFastLogger(err error, msg string) {
	if pl.ctx.IsFailFastEnabled() {
		pl.zap.Panicw(msg, "error", err)
	} else {
		pl.zap.Errorw(msg, "error", err)
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
