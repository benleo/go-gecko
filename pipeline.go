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
	ctx      Context
	serve    func(device InputDevice) Deliverer
	executor OutputExecutor
	// 事件派发
	dispatcher *Dispatcher
	// Pipeline关闭的信号控制
	shutdownCtx  context.Context
	shutdownFunc context.CancelFunc
}

// 准备运行环境，初始化相关组件
func (pl *Pipeline) prepareEnv() {
	pl.shutdownCtx, pl.shutdownFunc = context.WithCancel(context.Background())
	// 创建输入调度的处理
	pl.serve = func(device InputDevice) Deliverer {
		return func(topic string, frame PacketFrame) (PacketFrame, error) {
			// 解码
			input, deErr := device.GetDecoder()(frame.Data())
			if nil != deErr {
				pl.withTag(log.Error).Err(deErr).Msgf("InputDevice解码/Decode错误： %s", x.SimpleClassName(device))
				return nil, deErr
			}
			// 处理
			output := make(chan map[string]interface{}, 1)
			pl.dispatcher.Lv0() <- &_GeckoSession{
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
				pl.withTag(log.Error).Err(err).Msgf("InputDevice编码/Encode错误： %s", x.SimpleClassName(device))
				return nil, err
			} else {
				return NewPackFrame(bytes), nil
			}
		}
	}
	// 搜索设备，并执行
	pl.executor = OutputExecutor(func(unionOrGroupAddress string, isUnionAddress bool, frame PacketFrame) (PacketFrame, error) {
		if isUnionAddress {
			if device, ok := pl.namedOutputs[unionOrGroupAddress]; ok {
				return device.Process(frame, pl.ctx)
			} else {
				return nil, errors.New("OutputDeviceNotFound:" + unionOrGroupAddress)
			}
		} else /*is Group Address*/ {
			groupAddr := unionOrGroupAddress
			for addr, dev := range pl.namedOutputs {
				if strings.HasPrefix(addr, groupAddr) {
					if _, err := dev.Process(frame, pl.ctx); nil != err {
						pl.withTag(log.Error).Err(err).Msgf("OutputDevice处理广播错误： %s", x.SimpleClassName(dev))
						return nil, err
					}
				}
			}
			return nil, nil
		}
	})
}

// 初始化Pipeline
func (pl *Pipeline) Init(args map[string]interface{}) {
	geckoCtx := newGeckoContext(args)
	pl.ctx = geckoCtx
	gecko := pl.ctx.gecko()
	capacity := gecko.GetInt64OrDefault("eventsCapacity", 8)
	pl.withTag(log.Info).Msgf("事件通道容量： %d", capacity)
	pl.dispatcher = NewDispatcher(int(capacity))
	pl.dispatcher.SetLv0Handler(pl.handleInterceptor)
	pl.dispatcher.SetLv1Handler(pl.handleDriver)
	go pl.dispatcher.Serve(pl.shutdownCtx)

	// 初始化组件：根据配置文件指定项目
	itemInitWithContext := func(it Initialize, args map[string]interface{}) {
		it.OnInit(args, pl.ctx)
	}
	if !pl.registerBundlesIfHit(geckoCtx.plugins, itemInitWithContext) {
		pl.withTag(log.Warn).Msg("警告：未配置任何[Plugin]组件")
	}
	if !pl.registerBundlesIfHit(geckoCtx.outputs, itemInitWithContext) {
		pl.withTag(log.Panic).Msg("严重：未配置任何[OutputDevice]组件")
	}
	if !pl.registerBundlesIfHit(geckoCtx.interceptors, itemInitWithContext) {
		pl.withTag(log.Warn).Msg("警告：未配置任何[Interceptor]组件")
	}
	if !pl.registerBundlesIfHit(geckoCtx.drivers, itemInitWithContext) {
		pl.withTag(log.Warn).Msg("警告：未配置任何[Driver]组件")
	}
	if !pl.registerBundlesIfHit(geckoCtx.inputs, itemInitWithContext) {
		pl.withTag(log.Panic).Msg("严重：未配置任何[InputDevice]组件")
	}
	// show
	pl.showBundles()
}

// 启动Pipeline
func (pl *Pipeline) Start() {
	pl.withTag(log.Info).Msgf("Pipeline启动...")
	// Hook first
	x.ForEach(pl.startBeforeHooks, func(it interface{}) {
		it.(HookFunc)(pl)
	})
	defer func() {
		x.ForEach(pl.startAfterHooks, func(it interface{}) {
			it.(HookFunc)(pl)
		})
		pl.withTag(log.Info).Msgf("Pipeline启动...OK")
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
		go func() {
			if err := device.Serve(pl.ctx, pl.serve(device)); nil != err {
				pl.withTag(log.Error).Err(err).Msgf("InputDevice[%s]服务运行错误：", x.SimpleClassName(device))
			}
		}()
	})
}

// 停止Pipeline
func (pl *Pipeline) Stop() {
	pl.withTag(log.Info).Msgf("Pipeline停止...")
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
		pl.withTag(log.Info).Msgf("Pipeline停止...OK")
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
	pl.withTag(log.Warn).Msgf("接收到系统停止信号")
}

// 处理拦截器过程
func (pl *Pipeline) handleInterceptor(session Session) {
	pl.ctx.OnIfLogV(func() {
		pl.withTag(log.Debug).Msgf("Interceptor调度处理，Topic: %s", session.Topic())
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
			pl.withTag(log.Debug).Msgf("拦截器调度： interceptor[%s], topic: %s, Matches: %s",
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
			pl.withTag(log.Debug).Err(err).Msgf("拦截器中断事件： %s", err.Error())
			session.Outbound().AddDataField("error", "InterceptorDropped")
			// 终止，输出处理
			session.AddAttribute("Escaped@Interceptor", session.Escaped())
			pl.output(session)
			return
		} else {
			pl.failFastLogger().Err(err).Msgf("拦截器发生错误： %s", err.Error())
		}
	}
	// 继续
	session.AddAttribute("Escaped@Interceptor", session.Escaped())
	pl.dispatcher.Lv1() <- session
}

// 处理驱动执行过程
func (pl *Pipeline) handleDriver(session Session) {
	pl.ctx.OnIfLogV(func() {
		pl.withTag(log.Debug).Msgf("Driver调度处理，Topic: %s", session.Topic())
	})
	defer func() {
		pl.checkRecover(recover(), "Driver-Goroutine内部错误")
	}()

	// 查找匹配的用户驱动，并处理
	for el := pl.drivers.Front(); el != nil; el = el.Next() {
		driver := el.Value.(Driver)
		match := anyTopicMatches(driver.GetTopicExpr(), session.Topic())
		pl.ctx.OnIfLogV(func() {
			pl.withTag(log.Debug).Msgf("用户驱动处理： driver[%s], topic: %s, Matches: %s",
				x.SimpleClassName(driver),
				session.Topic(),
				strconv.FormatBool(match))
		})
		if match {
			err := driver.Handle(session, pl.executor, pl.ctx)
			if nil != err {
				pl.failFastLogger().Err(err).Msgf("用户驱动发生错误： %s", err.Error())
			}
		}
	}
	// 输出处理
	session.AddAttribute("Escaped@Driver", session.Escaped())
	pl.output(session)
}

func (pl *Pipeline) output(session Session) {
	pl.ctx.OnIfLogV(func() {
		pl.withTag(log.Debug).Msgf("Output调度处理，Topic: %s", session.Topic())
		session.Attributes().ForEach(func(k string, v interface{}) {
			pl.withTag(log.Debug).Msgf("SessionAttr: %s = %v", k, v)
		})
	})
	defer func() {
		pl.checkRecover(recover(), "Output-Goroutine内部错误")
	}()
	session.(*_GeckoSession).onSessionCompleted(session.Outbound().Data)
}

func (pl *Pipeline) checkDefTimeout(msg string, act func(Context)) {
	pl.ctx.CheckTimeout(msg, DefaultLifeCycleTimeout, func() {
		act(pl.ctx)
	})
}

func (pl *Pipeline) checkRecover(r interface{}, msg string) {
	if nil != r {
		if err, ok := r.(error); ok {
			pl.withTag(log.Error).Err(err).Msg(msg)
		}
		pl.ctx.OnIfFailFast(func() {
			panic(r)
		})
	}
}

func (pl *Pipeline) failFastLogger() *zerolog.Event {
	if pl.ctx.IsFailFastEnabled() {
		return pl.withTag(log.Panic)
	} else {
		return pl.withTag(log.Error)
	}
}

func (pl *Pipeline) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "Pipeline")
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
