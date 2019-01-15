package gecko

import (
	"context"
	"errors"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

var gSharedEngine = &Engine{
	Registration: prepare(),
}
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
	serve    func(device InputDevice) Deliverer
	executor OutputExecutor
	// 事件派发
	dispatcher *Dispatcher
	// Engine关闭的信号控制
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
}

// 准备运行环境，初始化相关组件
func (en *Engine) prepareEnv() {
	en.shutdownCtx, en.shutdownFunc = context.WithCancel(context.Background())
	// 创建输入调度的处理
	en.serve = func(device InputDevice) Deliverer {
		return func(topic string, frame PacketFrame) (PacketFrame, error) {
			// 解码
			input, deErr := device.GetDecoder()(frame.Data())
			if nil != deErr {
				en.withTag(log.Error).Err(deErr).Msgf("InputDevice解码/Decode错误： %s", x.SimpleClassName(device))
				return nil, deErr
			}
			// 处理
			output := make(chan map[string]interface{}, 1)
			en.dispatcher.Lv0() <- &_GeckoSession{
				timestamp:  time.Now(),
				attributes: make(map[string]interface{}),
				attrLock:   new(sync.RWMutex),
				topic:      topic,
				inbound: &Inbound{
					Topic: topic,
					Data:  input,
				},
				outbound: &Outbound{
					Topic: topic,
					Data:  make(map[string]interface{}),
				},
				onSessionCompleted: func(data map[string]interface{}) {
					output <- data
				},
			}
			// 编码
			if bytes, err := device.GetEncoder()(<-output); nil != err {
				en.withTag(log.Error).Err(err).Msgf("InputDevice编码/Encode错误： %s", x.SimpleClassName(device))
				return nil, err
			} else {
				return NewPackFrame(bytes), nil
			}
		}
	}
	// 搜索设备，并执行
	en.executor = OutputExecutor(func(unionOrGroupAddress string, isUnionAddress bool, frame PacketFrame) (PacketFrame, error) {
		if isUnionAddress {
			if device, ok := en.namedOutputs[unionOrGroupAddress]; ok {
				return device.Process(frame, en.ctx)
			} else {
				return nil, errors.New("OutputDeviceNotFound:" + unionOrGroupAddress)
			}
		} else /*is Group Address*/ {
			for addr, dev := range en.namedOutputs {
				if strings.HasPrefix(addr, "/"+unionOrGroupAddress) {
					if _, err := dev.Process(frame, en.ctx); nil != err {
						en.withTag(log.Error).Err(err).Msgf("OutputDevice处理广播错误： %s", x.SimpleClassName(dev))
					}
				}
			}
			return nil, nil
		}
	})
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
	gecko := en.ctx.gecko()
	capacity := gecko.GetInt64OrDefault("eventsCapacity", 8)
	en.withTag(log.Info).Msgf("事件通道容量： %d", capacity)
	en.dispatcher = NewDispatcher(int(capacity))
	en.dispatcher.SetLv0Handler(en.handleInterceptor)
	en.dispatcher.SetLv1Handler(en.handleDriver)
	go en.dispatcher.Serve(en.shutdownCtx)

	// 初始化组件：根据配置文件指定项目
	itemInitWithContext := func(it Initialize, args map[string]interface{}) {
		it.OnInit(args, en.ctx)
	}
	if !en.registerBundlesIfHit(geckoCtx.plugins, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Plugin]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.outputs, itemInitWithContext) {
		en.withTag(log.Panic).Msg("严重：未配置任何[OutputDevice]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.interceptors, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Interceptor]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.drivers, itemInitWithContext) {
		en.withTag(log.Warn).Msg("警告：未配置任何[Driver]组件")
	}
	if !en.registerBundlesIfHit(geckoCtx.inputs, itemInitWithContext) {
		en.withTag(log.Panic).Msg("严重：未配置任何[InputDevice]组件")
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
	// Drivers
	x.ForEach(en.drivers, func(it interface{}) {
		en.checkDefTimeout("Driver.Start", it.(Driver).OnStart)
	})
	// Inputs
	x.ForEach(en.inputs, func(it interface{}) {
		en.ctx.CheckTimeout("Trigger.Start", DefaultLifeCycleTimeout, func() {
			it.(InputDevice).OnStart(en.ctx)
		})
	})
	// Input Serve Last
	x.ForEach(en.inputs, func(it interface{}) {
		device := it.(InputDevice)
		go device.Serve(en.ctx, en.serve(device))
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
	// Inputs
	x.ForEach(en.inputs, func(it interface{}) {
		en.ctx.CheckTimeout("Trigger.Stop", DefaultLifeCycleTimeout, func() {
			it.(InputDevice).OnStop(en.ctx)
		})
	})
	// Drivers
	x.ForEach(en.drivers, func(it interface{}) {
		en.checkDefTimeout("Driver.Stop", it.(Driver).OnStop)
	})
	// Plugin
	x.ForEach(en.plugins, func(it interface{}) {
		en.checkDefTimeout("Plugin.Stop", it.(Plugin).OnStop)
	})
}

// 等待系统终止信息
func (en *Engine) AwaitTermination() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
	en.withTag(log.Warn).Msgf("接收到系统停止信号")
}

// 处理拦截器过程
func (en *Engine) handleInterceptor(session Session) {
	en.ctx.OnIfLogV(func() {
		en.withTag(log.Debug).Msgf("Interceptor调度处理，Topic: %s", session.Topic())
	})
	session.AddAttribute("Interceptor.Start", time.Now())
	defer func() {
		session.AddAttribute("Interceptor.End", session.Escaped())
		en.checkRecover(recover(), "Interceptor-Goroutine内部错误")
	}()
	// 查找匹配的拦截器，按优先级排序并处理
	matches := make(InterceptorSlice, 0)
	for el := en.interceptors.Front(); el != nil; el = el.Next() {
		interceptor := el.Value.(Interceptor)
		match := anyTopicMatches(interceptor.GetTopicExpr(), session.Topic())
		en.ctx.OnIfLogV(func() {
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
			// 终止，输出处理
			en.output(session)
			return
		} else {
			en.failFastLogger().Err(err).Msgf("拦截器发生错误： %s", err.Error())
		}
	}
	// 继续
	en.dispatcher.Lv1() <- session
}

// 处理驱动执行过程
func (en *Engine) handleDriver(session Session) {
	en.ctx.OnIfLogV(func() {
		en.withTag(log.Debug).Msgf("Driver调度处理，Topic: %s", session.Topic())
	})
	session.AddAttribute("Driver.Start", time.Now())
	defer func() {
		session.AddAttribute("Driver.End", session.Escaped())
		en.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()

	// 查找匹配的用户驱动，并处理
	for el := en.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		match := anyTopicMatches(driver.GetTopicExpr(), session.Topic())
		en.ctx.OnIfLogV(func() {
			en.withTag(log.Debug).Msgf("用户驱动处理： driver[%s], topic: %s, Matches: %s",
				x.SimpleClassName(driver),
				session.Topic(),
				strconv.FormatBool(match))
		})
		if match {
			err := driver.Handle(session, en.executor, en.ctx)
			if nil != err {
				en.failFastLogger().Err(err).Msgf("用户驱动发生错误： %s", err.Error())
			}
		} else {
			continue
		}
	}
	// 输出处理
	en.output(session)
}

func (en *Engine) output(session Session) {
	en.ctx.OnIfLogV(func() {
		en.withTag(log.Debug).Msgf("Output调度处理，Topic: %s", session.Topic())
		session.Attributes().ForEach(func(k string, v interface{}) {
			en.withTag(log.Debug).Msgf("SessionAttr: %s = %v", k, v)
		})
	})
	session.AddAttribute("Output.Start", time.Now())
	defer func() {
		session.AddAttribute("Output.End", session.Escaped())
		en.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	session.(*_GeckoSession).onSessionCompleted(session.Outbound().Data)
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
		en.ctx.OnIfFailFast(func() {
			panic(r)
		})
	}
}

func (en *Engine) failFastLogger() *zerolog.Event {
	if en.ctx.IsFailFastEnabled() {
		return en.withTag(log.Panic)
	} else {
		return en.withTag(log.Error)
	}
}

func newGeckoContext(config map[string]interface{}) *_GeckoContext {
	mapConf := conf.WrapImmutableMap(config)
	return &_GeckoContext{
		geckos:       mapConf.MustImmutableMap("GECKO"),
		globals:      mapConf.MustImmutableMap("GLOBALS"),
		interceptors: mapConf.MustImmutableMap("INTERCEPTORS"),
		drivers:      mapConf.MustImmutableMap("DRIVERS"),
		outputs:      mapConf.MustImmutableMap("OUTPUTS"),
		inputs:       mapConf.MustImmutableMap("INPUTS"),
		plugins:      mapConf.MustImmutableMap("PLUGINS"),
		magicKV:      make(map[interface{}]interface{}),
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
