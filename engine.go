package gecko

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko/x"
	"time"
)

////

// 默认组件生命周期超时时间：3秒
const DefaultLifeCycleTimeout = time.Second * 3

// Engine管理内部组件，处理事件。
type GeckoEngine struct {
	*RegisterEngine
	// ID生成器
	snowflake *Snowflake

	scoped   GeckoScoped
	invoker  TriggerInvoker
	selector ProtoPipelineSelector
	// 事件通道
	intChan chan GeckoContext
	driChan chan GeckoContext
	outChan chan GeckoContext
	// Engine已关闭的信号
	shutdownCompleted chan struct{}
}

// 准备运行环境，初始化相关组件
func (ge *GeckoEngine) PrepareEnv() {
	ge.RegisterEngine = prepare()
	// 查找Pipeline
	ge.selector = func(proto string) ProtoPipeline {
		return ge.pipelines[proto]
	}
	// 接收Trigger的输入事件
	ge.invoker = func(income *Income, callback TriggerCallback) {
		context := &abcGeckoContext{
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
			callback: callback,
		}
		ge.intChan <- context
	}
	// 事件循环
	go func(shouldBreak <-chan struct{}) {
		for {
			select {
			case <-shouldBreak:
				return

			case ctx := <-ge.intChan:
				go ge.handleInterceptor(ctx)

			case ctx := <-ge.driChan:
				go ge.handleDrivers(ctx)

			case ctx := <-ge.outChan:
				go ge.handleOutput(ctx)
			}
		}
	}(ge.shutdownCompleted)
}

// 初始化Engine
func (ge *GeckoEngine) Init(args map[string]interface{}) {
	config := conf.MapToMap(args)
	ge.scoped = newAbcGeckoScoped(args)
	if sf, err := NewSnowflake(ge.scoped.workerId()); nil != err {
		ge.withTag(log.Panic).Err(err).Msg("初始化发生错误")
	} else {
		ge.snowflake = sf
	}
	gecko := conf.MapToMap(ge.scoped.gecko())
	intCapacity := gecko.GetInt64OrDefault("interceptorChannelCapacity", 8)
	driCapacity := gecko.GetInt64OrDefault("driverChannelCapacity", 8)
	outCapacity := gecko.GetInt64OrDefault("outputChannelCapacity", 8)
	ge.intChan = make(chan GeckoContext, intCapacity)
	ge.driChan = make(chan GeckoContext, driCapacity)
	ge.outChan = make(chan GeckoContext, outCapacity)
	ge.shutdownCompleted = make(chan struct{}, 1)

	// 初始化组件：根据配置文件指定项目
	initWithScoped := func(it Initialize, args map[string]interface{}) {
		it.OnInit(args, ge.scoped)
	}
	ge.registerBundles(config.MustMap("PLUGINS"), initWithScoped)
	ge.registerBundles(config.MustMap("PIPELINES"), initWithScoped)
	ge.registerBundles(config.MustMap("INTERCEPTORS"), initWithScoped)
	ge.registerBundles(config.MustMap("DRIVERS"), initWithScoped)
	ge.registerBundles(config.MustMap("DEVICES"), initWithScoped)
	ge.registerBundles(config.MustMap("TRIGGERS"), initWithScoped)
}

// 启动Engine
func (ge *GeckoEngine) Start() {
	ge.withTag(log.Info).Msgf("Engine启动...")
	defer ge.withTag(log.Info).Msgf("Engine启动...OK")
	// Plugin
	x.ForEach(ge.plugins, func(it interface{}) {
		ge.checkDefTimeout(it.(Plugin).OnStart)
	})
	// Pipeline
	for _, pipeline := range ge.pipelines {
		ge.checkDefTimeout(pipeline.OnStart)
	}
	// Drivers
	x.ForEach(ge.drivers, func(it interface{}) {
		ge.checkDefTimeout(it.(Driver).OnStart)
	})
	// Trigger
	x.ForEach(ge.triggers, func(it interface{}) {
		ge.scoped.CheckTimeout(DefaultLifeCycleTimeout, func() {
			it.(Trigger).OnStart(ge.scoped, ge.invoker)
		})
	})
}

// 停止Engine
func (ge *GeckoEngine) Stop() {
	ge.withTag(log.Info).Msgf("Engine停止...")
	defer func() {
		// 最终发起关闭信息
		ge.shutdownCompleted <- struct{}{}
		ge.withTag(log.Info).Msgf("Engine停止...OK")
	}()
	// Triggers
	x.ForEach(ge.triggers, func(it interface{}) {
		ge.scoped.CheckTimeout(DefaultLifeCycleTimeout, func() {
			it.(Trigger).OnStop(ge.scoped, ge.invoker)
		})
	})
	// Drivers
	x.ForEach(ge.drivers, func(it interface{}) {
		ge.checkDefTimeout(it.(Driver).OnStop)
	})
	// Pipeline
	for _, pipeline := range ge.pipelines {
		ge.checkDefTimeout(pipeline.OnStop)
	}
	// Plugin
	x.ForEach(ge.plugins, func(it interface{}) {
		ge.checkDefTimeout(it.(Plugin).OnStop)
	})
}

// 处理拦截器过程
func (ge *GeckoEngine) handleInterceptor(ctx GeckoContext) {
	ctx.AddAttribute("Interceptor.Start", time.Now())
	defer func() {
		ctx.AddAttribute("Interceptor.End", time.Now())
		ge.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	ge.scoped.LogIfV(func() {
		ge.withTag(log.Debug).Msgf("Interceptor调度处理，Topic: %s", ctx.Topic())
	})
	// 查找匹配的拦截器，按优先级排序并处理
	// TODO 排序
	for el := ge.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		if anyTopicMatches(interceptor.GetTopicExpr(), ctx.Topic()) {
			err := interceptor.Handle(ctx, ge.scoped)
			if err == nil {
				continue
			}
			if err == ErrInterceptorDropped {
				ge.withTag(log.Debug).Err(err).Msgf("拦截器中断事件： %s", err.Error())
				ctx.Outbound().AddDataField("error", "InterceptorDropped")
				ge.outChan <- ctx
				return
			} else {
				logger := ge.withTag(log.Error)
				if ge.scoped.failFastEnabled() {
					logger = ge.withTag(log.Panic)
				}
				logger.Err(err).Msgf("拦截器发生错误： %s", err.Error())
			}
		}
	}
	// 继续驱动处理
	ge.driChan <- ctx
}

// 处理驱动执行过程
func (ge *GeckoEngine) handleDrivers(ctx GeckoContext) {
	ctx.AddAttribute("Driver.Start", time.Now())
	defer func() {
		ctx.AddAttribute("Driver.End", time.Now())
		ge.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()
	ge.scoped.LogIfV(func() {
		ge.withTag(log.Debug).Msgf("Driver调度处理，Topic: %s", ctx.Topic())
	})
	// 查找匹配的用户驱动，并处理
	for el := ge.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		if anyTopicMatches(driver.GetTopicExpr(), ctx.Topic()) {
			err := driver.Handle(ctx, ge.selector, ge.scoped)
			if nil != err {
				logger := ge.withTag(log.Error)
				if ge.scoped.failFastEnabled() {
					logger = ge.withTag(log.Panic)
				}
				logger.Err(err).Msgf("用户驱动发生错误： %s", err.Error())
			}
		} else {
			continue
		}
	}
	// 输出处理
	ge.outChan <- ctx
}

// 返回Trigger输出
func (ge *GeckoEngine) handleOutput(ctx GeckoContext) {
	ctx.AddAttribute("Output.Start", time.Now())
	defer func() {
		ctx.AddAttribute("Output.End", time.Now())
		ge.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	ctx.(*abcGeckoContext).callback(ctx.Topic(), ctx.Outbound().Data)
}

func (ge *GeckoEngine) checkDefTimeout(act func(GeckoScoped)) {
	ge.scoped.CheckTimeout(DefaultLifeCycleTimeout, func() {
		act(ge.scoped)
	})
}

func (ge *GeckoEngine) checkRecover(r interface{}, msg string) {
	if nil != r {
		if err, ok := r.(error); ok {
			ge.withTag(log.Error).Err(err).Msg(msg)
		}
		if ge.scoped.failFastEnabled() {
			panic(r)
		}
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
