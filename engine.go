package gecko

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko/x"
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
		select {
		case en.intChan <- ss:
			en.withTag(log.Debug).Msgf("Trigger已发送事件： %s", income.topic)

		case <-time.After(time.Second):
			en.failFastLogger().Msgf("Trigger事件阻塞超时： %s", income.topic)
		}

	}
	// 消息循环
	go func(breakSig <-chan struct{}, inr <-chan Session, dri <-chan Session, out <-chan Session) {
		defer en.withTag(log.Info).Msg("已退出消息循环")
		for {
			en.withTag(log.Info).Msg("进入消息循环。。。。")
			select {
			case <-breakSig:
				en.withTag(log.Info).Msg("已退出消息循环")
				return

			case ss := <-inr:
				en.withTag(log.Debug).Msgf("主循环接收Int事件： %s", ss.Topic())
				go en.handleInterceptor(ss, en.driChan)

				case ss := <-dri:
					en.withTag(log.Debug).Msgf("主循环接收Dri事件： %s", ss.Topic())
					go en.handleDrivers(ss, en.outChan)

				case ss := <-out:
					en.withTag(log.Debug).Msgf("主循环接收Out事件： %s", ss.Topic())
					go en.handleOutput(ss)
			}
		}
	}(en.shutdownCtx.Done(), en.intChan, en.driChan, en.outChan)
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
	intCap := gecko.GetInt64OrDefault("interceptorChannelCapacity", 8)
	driCap := gecko.GetInt64OrDefault("driverChannelCapacity", 8)
	outCap := gecko.GetInt64OrDefault("outputChannelCapacity", 8)
	en.intChan = make(chan Session, intCap)
	en.driChan = make(chan Session, driCap)
	en.outChan = make(chan Session, outCap)
	en.withTag(log.Info).Msgf("拦截通道容量： %d", intCap)
	en.withTag(log.Info).Msgf("驱动通道容量： %d", driCap)
	en.withTag(log.Info).Msgf("输出通道容量： %d", outCap)

	// 初始化组件：根据配置文件指定项目
	itemInitWithContext := func(it Initialize, args map[string]interface{}) {
		it.OnInit(args, en.ctx)
	}

	if !en.registerBundlesIfHit(geckoCtx.confPlugins, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Plugin]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.confPipelines, itemInitWithContext) {
		en.withTag(log.Panic).Msg("严重：未配置任何[Pipeline]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.confDevices, itemInitWithContext) {
		en.withTag(log.Panic).Msg("严重：未配置任何[Devices]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.confInterceptors, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Interceptor]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.confDrivers, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Driver]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.confTriggers, itemInitWithContext) {
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
func (en *Engine) handleInterceptor(session Session, nextChan chan<- Session) {
	en.ctx.LogIfV(func() {
		en.withTag(log.Debug).Msgf("Interceptor调度处理，Topic: %s", session.Topic())
	})
	session.AddAttribute("Interceptor.Start", time.Now())
	defer func() {
		session.AddAttribute("Interceptor.End", time.Now())
		en.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	// 查找匹配的拦截器，按优先级排序并处理
	matches := make(InterceptorSlice, 0)
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
			matches = append(matches, interceptor)
		}
	}
	sort.Sort(matches)
	// 按排序结果顺序执行
	for _, it := range matches {
		err := it.Handle(session, en.ctx)
		if err == nil {
			continue
		}
		if err == ErrInterceptorDropped {
			en.withTag(log.Debug).Err(err).Msgf("拦截器中断事件： %s", err.Error())
			session.Outbound().AddDataField("error", "InterceptorDropped")
			en.outChan <- session
			return
		} else {
			en.failFastLogger().Err(err).Msgf("拦截器发生错误： %s", err.Error())
		}
	}
	// 继续驱动处理
	select {
	case nextChan <- session:

	case <-time.After(time.Second):
		en.failFastLogger().Msgf("驱动处理事件阻塞超时： %s", session.Topic())
	}
}

// 处理驱动执行过程
func (en *Engine) handleDrivers(session Session, nextChan chan<- Session) {
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
				en.failFastLogger().Err(err).Msgf("用户驱动发生错误： %s", err.Error())
			}
		} else {
			continue
		}
	}
	// 输出处理
	select {
	case nextChan <- session:

	case <-time.After(time.Second):
		en.failFastLogger().Msgf("输出处理事件阻塞超时： %s", session.Topic())
	}
}

// 返回Trigger输出
func (en *Engine) handleOutput(session Session) {
	en.ctx.LogIfV(func() {
		en.withTag(log.Debug).Msgf("Output调度处理，Topic: %s", session.Topic())
		for k, v := range session.Attributes() {
			en.withTag(log.Debug).Msgf("SessionAttr: %s = %v", k, v)
		}
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

func (en *Engine) failFastLogger() *zerolog.Event {
	if en.ctx.failFastEnabled() {
		return en.withTag(log.Panic)
	} else {
		return en.withTag(log.Error)
	}
}

func newGeckoContext(config map[string]interface{}) *contextImpl {
	mapConf := conf.MapToMap(config)
	return &contextImpl{
		confGecko:        mapConf.MustMap("GECKO"),
		confGlobals:      mapConf.MustMap("GLOBALS"),
		confPipelines:    mapConf.MustMap("PIPELINES"),
		confInterceptors: mapConf.MustMap("INTERCEPTORS"),
		confDrivers:      mapConf.MustMap("DRIVERS"),
		confDevices:      mapConf.MustMap("DEVICES"),
		confTriggers:     mapConf.MustMap("TRIGGERS"),
		confPlugins:      mapConf.MustMap("PLUGINS"),
		magicKV:          make(map[interface{}]interface{}),
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
