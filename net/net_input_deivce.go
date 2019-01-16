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
	onServeHandler func(bytes []byte, ctx gecko.Context, deliverer gecko.InputDeliverer) error
	topic          string
}

func (d *AbcNetInputDevice) OnInit(args map[string]interface{}, ctx gecko.Context) {
	config := conf.WrapImmutableMap(args)
	d.maxBufferSize = config.GetInt64OrDefault("bufferSize", 512)
	d.readTimeout = config.GetDurationOrDefault("readTimeout", time.Second*3)
	d.topic = config.MustString("topic")
}

func (d *AbcNetInputDevice) OnStart(ctx gecko.Context) {
	d.cancelCtx, d.cancelFun = context.WithCancel(context.Background())
	if nil == d.onServeHandler {
		d.withTag(log.Warn).Msg("使用默认数据处理接口")
		if "" == d.topic {
			d.withTag(log.Panic).Msg("使用默认接口必须设置topic参数")
		}
		d.onServeHandler = func(bytes []byte, ctx gecko.Context, deliverer gecko.InputDeliverer) error {
			return deliverer.Broadcast(d.topic, gecko.PacketFrame(bytes))
		}
	}
}

func (d *AbcNetInputDevice) OnStop(ctx gecko.Context) {
	d.cancelFun()
}

func (d *AbcNetInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	if nil == d.onServeHandler {
		return errors.New("未设置onServeHandler接口")
	}
	address := d.GetUnionAddress()
	d.withTag(log.Info).Msgf("使用%s服务端模式，监听端口: %s", d.network, address)
	if "udp" == d.network {
		if addr, err := net.ResolveUDPAddr("udp", address); err != nil {
			return errors.New("无法创建UDP地址: " + address)
		} else {
			if conn, err := net.ListenUDP("udp", addr); nil != err {
				return err
			} else {
				return d.loop(conn, ctx, deliverer)
			}
		}
	} else if "tcp" == d.network {
		server, err := net.Listen("tcp", address)
		if nil != err {
			return err
		}
		wg := new(sync.WaitGroup)
		for {
			select {
			case <-d.cancelCtx.Done():
				server.Close()
				return nil

			default:
				if conn, err := server.Accept(); nil != err {
					if !d.isNetTempErr(err) {
						d.withTag(log.Error).Err(err).Msg("TCP服务端网络错误")
						return err
					}
				} else {
					go func() {
						defer wg.Done()
						wg.Add(1)
						if err := d.loop(conn, ctx, deliverer); nil != err {
							d.withTag(log.Error).Err(err).Msg("TCP客户端发生错误")
						}
					}()
				}
			}
		}
		wg.Wait()
		return nil
	} else {
		return errors.New("未识别的网络连接模式: " + d.network)
	}
}

// 由于不需要返回响应数据到NetInputDevice，Encoder编码器可以不做业务处理
func (d *AbcNetInputDevice) GetEncoder() gecko.Encoder {
	return func(data gecko.PacketMap) (gecko.PacketFrame, error) {
		return gecko.NewPackFrame([]byte{}), nil
	}
}

func (d *AbcNetInputDevice) Topic() string {
	return d.topic
}

func (d *AbcNetInputDevice) loop(conn net.Conn, ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	defer conn.Close()
	buffer := make([]byte, d.maxBufferSize)
	for {
		select {
		case <-d.cancelCtx.Done():
			return nil

		default:
			if err := conn.SetReadDeadline(time.Now().Add(d.readTimeout)); nil != err {
				if !d.isNetTempErr(err) {
					return err
				} else {
					continue
				}
			}

			if n, err := conn.Read(buffer); nil != err {
				if !d.isNetTempErr(err) {
					return err
				}
			} else if n > 0 {
				frame := gecko.NewPackFrame(buffer[:n])
				if err := d.onServeHandler(frame, ctx, deliverer); nil != err {
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
func (d *AbcNetInputDevice) SetServeHandler(handler func([]byte, gecko.Context, gecko.InputDeliverer) error) {
	d.onServeHandler = handler
}

func (d *AbcNetInputDevice) withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "AbcNetInputDevice")
}
