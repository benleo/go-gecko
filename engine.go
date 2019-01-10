package gecko

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko/x"
	"os"
	"os/signal"
	"strconv"
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
	en.selector = func(proto string) (pl ProtoPipeline, ok bool) {
		pl, ok = en.pipelines[proto]
		return
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
		defer en.withTag(log.Info).Msg("已退出消息循环")
		for {
			select {
			case <-breakSig:
				return

			case ss := <-en.intChan:
				go en.handleInterceptor(ss)

			case ss := <-en.driChan:
				go en.handleDrivers(ss)

			case ss := <-en.outChan:
				go en.handleOutput(ss)
			}
		}
	}(en.shutdownCtx.Done())
}

// 初始化Engine
func (en *Engine) Init(args map[string]interface{}) {
	geckoCtx := newGeckoContext(args)
	en.ctx = geckoCtx
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
	itemInitWithContext := func(it Initialize, args map[string]interface{}) {
		it.OnInit(args, en.ctx)
	}

	if !en.registerBundlesIfHit(geckoCtx.pluginsConf, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Plugin]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.pipelinesConf, itemInitWithContext) {
		en.withTag(log.Panic).Msg("严重：未配置任何[Pipeline]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.devicesConf, itemInitWithContext) {
		en.withTag(log.Panic).Msg("严重：未配置任何[Devices]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.interceptorsConf, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Interceptor]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.driversConf, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Driver]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.triggersConf, itemInitWithContext) {
		en.withTag(log.Panic).Msg("严重：未配置任何[Trigger]组件")
	}
	// show
	en.showBundles()
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
		en.checkDefTimeout("Plugin.Start", it.(Plugin).OnStart)
	})
	// Pipeline
	for _, pipeline := range en.pipelines {
		en.checkDefTimeout("Pipeline.Start", pipeline.OnStart)
	}
	// Drivers
	x.ForEach(en.drivers, func(it interface{}) {
		en.checkDefTimeout("Driver.Start", it.(Driver).OnStart)
	})
	// Trigger
	x.ForEach(en.triggers, func(it interface{}) {
		en.ctx.CheckTimeout("Trigger.Start", DefaultLifeCycleTimeout, func() {
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
		en.ctx.CheckTimeout("Trigger.Stop", DefaultLifeCycleTimeout, func() {
			it.(Trigger).OnStop(en.ctx, en.invoker)
		})
	})
	// Drivers
	x.ForEach(en.drivers, func(it interface{}) {
		en.checkDefTimeout("Driver.Stop", it.(Driver).OnStop)
	})
	// Pipeline
	for _, pipeline := range en.pipelines {
		en.checkDefTimeout("Pipeline.Stop", pipeline.OnStop)
	}
	// Plugin
	x.ForEach(en.plugins, func(it interface{}) {
		en.checkDefTimeout("Plugin.Stop", it.(Plugin).OnStop)
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
	en.ctx.LogIfV(func() {
		en.withTag(log.Debug).Msgf("Interceptor调度处理，Topic: %s", session.Topic())
	})
	session.AddAttribute("Interceptor.Start", time.Now())
	defer func() {
		session.AddAttribute("Interceptor.End", time.Now())
		en.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	// 查找匹配的拦截器，按优先级排序并处理
	// TODO 排序
	for el := en.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		match := anyTopicMatches(interceptor.GetTopicExpr(), session.Topic())
		en.ctx.LogIfV(func() {
			en.withTag(log.Debug).Msgf("拦截器调度： interceptor[%s], topic: %s, Matches: %s",
				x.SimpleClassName(interceptor),
				session.Topic(),
				strconv.FormatBool(match))
		})
		if match {
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
	en.ctx.LogIfV(func() {
		en.withTag(log.Debug).Msgf("Driver调度处理，Topic: %s", session.Topic())
	})
	session.AddAttribute("Driver.Start", time.Now())
	defer func() {
		session.AddAttribute("Driver.End", time.Now())
		en.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()

	// 查找匹配的用户驱动，并处理
	for el := en.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		match := anyTopicMatches(driver.GetTopicExpr(), session.Topic())
		en.ctx.LogIfV(func() {
			en.withTag(log.Debug).Msgf("用户驱动处理： driver[%s], topic: %s, Matches: %s",
				x.SimpleClassName(driver),
				session.Topic(),
				strconv.FormatBool(match))
		})
		if match {
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
	en.ctx.LogIfV(func() {
		en.withTag(log.Debug).Msgf("Output调度处理，Topic: %s", session.Topic())
	})
	session.AddAttribute("Output.Start", time.Now())
	defer func() {
		session.AddAttribute("Output.End", time.Now())
		en.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	session.(*sessionImpl).onCompletedFunc(session.Outbound().Data)
}

func (en *Engine) checkDefTimeout(msg string, act func(Context)) {
	en.ctx.CheckTimeout(msg, DefaultLifeCycleTimeout, func() {
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

func newGeckoContext(config map[string]interface{}) *contextImpl {
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
