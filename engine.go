package gecko

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko/x"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

////

// 默认组件生命周期超时时间：3秒
const DefaultLifeCycleTimeout = time.Second * 3

var gSharedEngine = new(Engine)
var gPrepareEnv = new(sync.Once)

// 全局Engine对象
func SharedEngine() *Engine {
	gPrepareEnv.Do(func() {
		gSharedEngine.prepareEnv()
	})
	return gSharedEngine
}

// Engine管理内部组件，处理事件。
type Engine struct {
	*Registration
	// ID生成器
	snowflake *Snowflake

	ctx      Context
	invoker  Invoker
	selector ProtoPipelineSelector
	// 事件通道
	intChan chan Session
	driChan chan Session
	outChan chan Session
	// Engine关闭的信号控制
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
}

// 准备运行环境，初始化相关组件
func (en *Engine) prepareEnv() {
	en.Registration = prepare()
	en.shutdownCtx, en.shutdownFunc = context.WithCancel(context.Background())
	// 查找Pipeline
	en.selector = func(proto string) ProtoPipeline {
		return en.pipelines[proto]
	}
	// 接收Trigger的输入事件
	en.invoker = func(income *TriggerEvent, cbFunc OnTriggerCompleted) {
		en.ctx.LogIfV(func() {
			en.withTag(log.Debug).Msgf("Invoker接收请求，Topic: %s", income.topic)
		})
		ss := &sessionImpl{
			timestamp:  time.Now(),
			attributes: make(map[string]interface{}),
			topic:      income.topic,
			contextId:  0,
			inbound: &Inbound{
				Topic: income.topic,
				Data:  income.data,
			},
			outbound: &Outbound{
				Topic: income.topic,
				Data:  make(map[string]interface{}),
			},
			onCompletedFunc: cbFunc,
		}
		en.intChan <- ss
	}
	// 消息循环
	go func(breakSig <-chan struct{}) {
		for {
			select {
			case <-breakSig:
				return

			case ctx := <-en.intChan:
				go en.handleInterceptor(ctx)

			case ctx := <-en.driChan:
				go en.handleDrivers(ctx)

			case ctx := <-en.outChan:
				go en.handleOutput(ctx)
			}
		}
	}(en.shutdownCtx.Done())
}

// 初始化Engine
func (en *Engine) Init(args map[string]interface{}) {
	config := conf.MapToMap(args)
	en.ctx = newGeckoContext(args)
	if sf, err := NewSnowflake(en.ctx.workerId()); nil != err {
		en.withTag(log.Panic).Err(err).Msg("初始化发生错误")
	} else {
		en.snowflake = sf
	}
	gecko := conf.MapToMap(en.ctx.gecko())
	intCapacity := gecko.GetInt64OrDefault("interceptorChannelCapacity", 8)
	driCapacity := gecko.GetInt64OrDefault("driverChannelCapacity", 8)
	outCapacity := gecko.GetInt64OrDefault("outputChannelCapacity", 8)
	en.intChan = make(chan Session, intCapacity)
	en.driChan = make(chan Session, driCapacity)
	en.outChan = make(chan Session, outCapacity)

	// 初始化组件：根据配置文件指定项目
	initWithScoped := func(it Initialize, args map[string]interface{}) {
		it.OnInit(args, en.ctx)
	}
	en.registerBundles(config.MustMap("PLUGINS"), initWithScoped)
	en.registerBundles(config.MustMap("PIPELINES"), initWithScoped)
	en.registerBundles(config.MustMap("INTERCEPTORS"), initWithScoped)
	en.registerBundles(config.MustMap("DRIVERS"), initWithScoped)
	en.registerBundles(config.MustMap("DEVICES"), initWithScoped)
	en.registerBundles(config.MustMap("TRIGGERS"), initWithScoped)
}

// 启动Engine
func (en *Engine) Start() {
	en.withTag(log.Info).Msgf("Engine启动...")
	// Hook first
	x.ForEach(en.startBeforeHooks, func(it interface{}) {
		it.(HookFunc)(en)
	})
	defer func() {
		x.ForEach(en.startAfterHooks, func(it interface{}) {
			it.(HookFunc)(en)
		})
		en.withTag(log.Info).Msgf("Engine启动...OK")
	}()

	// Plugin
	x.ForEach(en.plugins, func(it interface{}) {
		en.checkDefTimeout(it.(Plugin).OnStart)
	})
	// Pipeline
	for _, pipeline := range en.pipelines {
		en.checkDefTimeout(pipeline.OnStart)
	}
	// Drivers
	x.ForEach(en.drivers, func(it interface{}) {
		en.checkDefTimeout(it.(Driver).OnStart)
	})
	// Trigger
	x.ForEach(en.triggers, func(it interface{}) {
		en.ctx.CheckTimeout(DefaultLifeCycleTimeout, func() {
			it.(Trigger).OnStart(en.ctx, en.invoker)
		})
	})
}

// 停止Engine
func (en *Engine) Stop() {
	en.withTag(log.Info).Msgf("Engine停止...")
	// Hook first
	x.ForEach(en.stopBeforeHooks, func(it interface{}) {
		it.(HookFunc)(en)
	})
	defer func() {
		x.ForEach(en.stopAfterHooks, func(it interface{}) {
			it.(HookFunc)(en)
		})
		// 最终发起关闭信息
		en.shutdownFunc()
		en.withTag(log.Info).Msgf("Engine停止...OK")
	}()
	// Triggers
	x.ForEach(en.triggers, func(it interface{}) {
		en.ctx.CheckTimeout(DefaultLifeCycleTimeout, func() {
			it.(Trigger).OnStop(en.ctx, en.invoker)
		})
	})
	// Drivers
	x.ForEach(en.drivers, func(it interface{}) {
		en.checkDefTimeout(it.(Driver).OnStop)
	})
	// Pipeline
	for _, pipeline := range en.pipelines {
		en.checkDefTimeout(pipeline.OnStop)
	}
	// Plugin
	x.ForEach(en.plugins, func(it interface{}) {
		en.checkDefTimeout(it.(Plugin).OnStop)
	})
}

// 等待系统终止信息
func (en *Engine) AwaitTermination() {
	sysSignal := make(chan os.Signal, 1)
	signal.Notify(sysSignal, syscall.SIGINT, syscall.SIGTERM)
	<-sysSignal
	en.withTag(log.Warn).Msgf("接收到系统停止信号")
}

// 处理拦截器过程
func (en *Engine) handleInterceptor(session Session) {
	session.AddAttribute("Interceptor.Start", time.Now())
	defer func() {
		session.AddAttribute("Interceptor.End", time.Now())
		en.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	en.ctx.LogIfV(func() {
		en.withTag(log.Debug).Msgf("Interceptor调度处理，Topic: %s", session.Topic())
	})
	// 查找匹配的拦截器，按优先级排序并处理
	// TODO 排序
	for el := en.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		if anyTopicMatches(interceptor.GetTopicExpr(), session.Topic()) {
			err := interceptor.Handle(session, en.ctx)
			if err == nil {
				continue
			}
			if err == ErrInterceptorDropped {
				en.withTag(log.Debug).Err(err).Msgf("拦截器中断事件： %s", err.Error())
				session.Outbound().AddDataField("error", "InterceptorDropped")
				en.outChan <- session
				return
			} else {
				logger := en.withTag(log.Error)
				if en.ctx.failFastEnabled() {
					logger = en.withTag(log.Panic)
				}
				logger.Err(err).Msgf("拦截器发生错误： %s", err.Error())
			}
		}
	}
	// 继续驱动处理
	en.driChan <- session
}

// 处理驱动执行过程
func (en *Engine) handleDrivers(session Session) {
	session.AddAttribute("Driver.Start", time.Now())
	defer func() {
		session.AddAttribute("Driver.End", time.Now())
		en.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()
	en.ctx.LogIfV(func() {
		en.withTag(log.Debug).Msgf("Driver调度处理，Topic: %s", session.Topic())
	})
	// 查找匹配的用户驱动，并处理
	for el := en.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		if anyTopicMatches(driver.GetTopicExpr(), session.Topic()) {
			err := driver.Handle(session, en.selector, en.ctx)
			if nil != err {
				logger := en.withTag(log.Error)
				if en.ctx.failFastEnabled() {
					logger = en.withTag(log.Panic)
				}
				logger.Err(err).Msgf("用户驱动发生错误： %s", err.Error())
			}
		} else {
			continue
		}
	}
	// 输出处理
	en.outChan <- session
}

// 返回Trigger输出
func (en *Engine) handleOutput(session Session) {
	session.AddAttribute("Output.Start", time.Now())
	defer func() {
		session.AddAttribute("Output.End", time.Now())
		en.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	session.(*sessionImpl).onCompletedFunc(session.Outbound().Data)
}

func (en *Engine) checkDefTimeout(act func(Context)) {
	en.ctx.CheckTimeout(DefaultLifeCycleTimeout, func() {
		act(en.ctx)
	})
}

func (en *Engine) checkRecover(r interface{}, msg string) {
	if nil != r {
		if err, ok := r.(error); ok {
			en.withTag(log.Error).Err(err).Msg(msg)
		}
		if en.ctx.failFastEnabled() {
			panic(r)
		}
	}
}

func newGeckoContext(config map[string]interface{}) Context {
	mapConf := conf.MapToMap(config)
	return &contextImpl{
		geckoConf:        mapConf.MustMap("GECKO"),
		globalsConf:      mapConf.MustMap("GLOBALS"),
		pipelinesConf:    mapConf.MustMap("PIPELINES"),
		interceptorsConf: mapConf.MustMap("INTERCEPTORS"),
		driversConf:      mapConf.MustMap("DRIVERS"),
		devicesConf:      mapConf.MustMap("DEVICES"),
		triggersConf:     mapConf.MustMap("TRIGGERS"),
		pluginsConf:      mapConf.MustMap("PLUGINS"),
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
