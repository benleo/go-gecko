package net

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"net"
	"sync"
	"time"
)

func NewAbcNetInputDevice(network string) *AbcNetInputDevice {
	return &AbcNetInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
		network:        network,
	}
}

// UDP服务器读取设备
type AbcNetInputDevice struct {
	*gecko.AbcInputDevice
	network        string
	maxBufferSize  int64
	readTimeout    time.Duration
	cancelCtx      context.Context
	cancelFun      context.CancelFunc
	onServeHandler func(bytes []byte, ctx gecko.Context, deliverer gecko.Deliverer) error
	topic          string
}

func (ui *AbcNetInputDevice) OnInit(args map[string]interface{}, ctx gecko.Context) {
	config := conf.WrapImmutableMap(args)
	ui.maxBufferSize = config.GetInt64OrDefault("bufferSize", 512)
	ui.readTimeout = config.GetDurationOrDefault("readTimeout", time.Second*3)
	ui.topic = config.MustString("topic")
}

func (ui *AbcNetInputDevice) OnStart(ctx gecko.Context) {
	ui.cancelCtx, ui.cancelFun = context.WithCancel(context.Background())
	if nil == ui.onServeHandler {
		ui.withTag(log.Warn).Msg("使用默认数据处理接口")
		if "" == ui.topic {
			ui.withTag(log.Panic).Msg("使用默认接口必须设置topic参数")
		}
		ui.onServeHandler = func(bytes []byte, ctx gecko.Context, deliverer gecko.Deliverer) error {
			return deliverer.Broadcast(ui.topic, gecko.PacketFrame(bytes))
		}
	}
}

func (ui *AbcNetInputDevice) OnStop(ctx gecko.Context) {
	ui.cancelFun()
}

func (ui *AbcNetInputDevice) Serve(ctx gecko.Context, deliverer gecko.Deliverer) error {
	if nil == ui.onServeHandler {
		return errors.New("未设置onServeHandler接口")
	}
	address := ui.GetUnionAddress()
	ui.withTag(log.Info).Msgf("使用%s服务端模式，监听端口: %s", ui.network, address)
	if "udp" == ui.network {
		if addr, err := net.ResolveUDPAddr("udp", address); err != nil {
			return errors.New("无法创建UDP地址: " + address)
		} else {
			if conn, err := net.ListenUDP("udp", addr); nil != err {
				return err
			} else {
				return ui.loop(conn, ctx, deliverer)
			}
		}
	} else if "tcp" == ui.network {
		server, err := net.Listen("tcp", address)
		if nil != err {
			return err
		}
		wg := new(sync.WaitGroup)
		for {
			select {
			case <-ui.cancelCtx.Done():
				server.Close()
				return nil

			default:
				if conn, err := server.Accept(); nil != err {
					if !ui.isNetTempErr(err) {
						ui.withTag(log.Error).Err(err).Msg("TCP服务端网络错误")
						return err
					}
				} else {
					go func() {
						defer wg.Done()
						wg.Add(1)
						if err := ui.loop(conn, ctx, deliverer); nil != err {
							ui.withTag(log.Error).Err(err).Msg("TCP客户端发生错误")
						}
					}()
				}
			}
		}
		wg.Wait()
		return nil
	} else {
		return errors.New("未识别的网络连接模式: " + ui.network)
	}
}

func (ui *AbcNetInputDevice) loop(conn net.Conn, ctx gecko.Context, deliverer gecko.Deliverer) error {
	defer conn.Close()
	buffer := make([]byte, ui.maxBufferSize)
	for {
		select {
		case <-ui.cancelCtx.Done():
			return nil

		default:
			if err := conn.SetReadDeadline(time.Now().Add(ui.readTimeout)); nil != err {
				if !ui.isNetTempErr(err) {
					return err
				} else {
					continue
				}
			}

			if n, err := conn.Read(buffer); nil != err {
				if !ui.isNetTempErr(err) {
					return err
				}
			} else if n > 0 {
				frame := gecko.NewPackFrame(buffer[:n])
				if err := ui.onServeHandler(frame, ctx, deliverer); nil != err {
					return err
				}
			}
		}
	}
}

func (*AbcNetInputDevice) isNetTempErr(err error) bool {
	if nErr, ok := err.(net.Error); ok {
		return nErr.Timeout() || nErr.Temporary()
	} else {
		return false
	}
}

// 设置Serve处理函数
func (ui *AbcNetInputDevice) SetServeHandler(handler func([]byte, gecko.Context, gecko.Deliverer) error) {
	ui.onServeHandler = handler
}

func (ui *AbcNetInputDevice) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "AbcNetInputDevice")
}
