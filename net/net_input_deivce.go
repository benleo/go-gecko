package net

import (
	"context"
	"github.com/parkingwang/go-conf"
	"github.com/pkg/errors"
	"github.com/yoojia/go-gecko"
	"net"
	"time"
)

func NewAbcNetInputDevice(network string) *AbcNetInputDevice {
	return &AbcNetInputDevice{
		AbcInputDevice: gecko.NewAbcInputDevice(),
		networkType:    network,
	}
}

// Socket服务器读取设备
type AbcNetInputDevice struct {
	*gecko.AbcInputDevice
	networkType    string
	networkAddress string
	maxBufferSize  int64
	readTimeout    time.Duration
	cancelCtx      context.Context
	serverCancelFn context.CancelFunc
	onServeHandler func(bytes []byte, ctx gecko.Context, deliverer gecko.InputDeliverer) error
	topic          string
}

func (d *AbcNetInputDevice) OnInit(config *cfg.Config, ctx gecko.Context) {
	d.AbcInputDevice.OnInit(config, ctx)
	d.maxBufferSize = config.GetInt64OrDefault("bufferSize", 512)
	d.readTimeout = config.GetDurationOrDefault("readTimeout", time.Second*3)
	d.topic = config.MustString("topic")
	d.networkAddress = config.MustString("networkAddress")
}

func (d *AbcNetInputDevice) OnStart(ctx gecko.Context) {
	d.cancelCtx, d.serverCancelFn = context.WithCancel(context.Background())
	zlog := gecko.ZapSugarLogger

	if d.networkAddress == "" || d.networkType == "" {
		zlog.Panicw("未设置网络通讯地址和网络类型", "address", d.networkAddress, "type", d.networkType)
	}
	if nil == d.onServeHandler {
		zlog.Warn("使用默认数据处理接口")
		if "" == d.topic {
			zlog.Panic("使用默认接口必须设置topic参数")
		}
		d.onServeHandler = func(bytes []byte, ctx gecko.Context, deliverer gecko.InputDeliverer) error {
			return deliverer.Broadcast(d.topic, gecko.FramePacket(bytes))
		}
	}
}

func (d *AbcNetInputDevice) OnStop(ctx gecko.Context) {
	d.serverCancelFn()
}

func (d *AbcNetInputDevice) Serve(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	if nil == d.onServeHandler {
		return errors.New("未设置onServeHandler接口")
	}
	gecko.ZapSugarLogger.Infof("使用%s服务端模式，监听端口: %s", d.networkType, d.networkAddress)
	if "udp" == d.networkType {
		return d.udpServe(ctx, deliverer)
	} else if "tcp" == d.networkType {
		return d.tcpServe(ctx, deliverer)
	} else {
		return errors.New("未识别的网络连接模式: " + d.networkType)
	}
}

func (d *AbcNetInputDevice) udpServe(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	if addr, err := net.ResolveUDPAddr("udp", d.networkAddress); err != nil {
		return errors.New("无法创建UDP地址: " + d.networkAddress)
	} else {
		if conn, err := net.ListenUDP("udp", addr); nil != err {
			return errors.WithMessage(err, "UDP连接监听失败")
		} else {
			return d.receiveLoop(conn, ctx, deliverer)
		}
	}
}

func (d *AbcNetInputDevice) tcpServe(ctx gecko.Context, deliverer gecko.InputDeliverer) error {
	zlog := gecko.ZapSugarLogger
	server, err := net.Listen("tcp", d.networkAddress)
	if nil != err {
		return errors.WithMessage(err, "TCP连接监听失败")
	}
	for {
		select {
		case <-d.cancelCtx.Done():
			if err := server.Close(); nil != err {
				zlog.Errorf("关闭%s服务器发生错误", d.networkType, err)
			}
			break

		default:
			if client, err := server.Accept(); nil != err {
				if !d.isNetTempErr(err) {
					zlog.Errorw("TCP服务端网络错误", "error", err)
					return err
				}
			} else {
				go func() {
					if err := d.receiveLoop(client, ctx, deliverer); nil != err {
						zlog.Errorw("TCP客户端发生错误", "error", err)
					}
				}()
			}
		}
	}
}

// 由于不需要返回响应数据到NetInputDevice，Encoder编码器可以不做业务处理
func (d *AbcNetInputDevice) GetEncoder() gecko.Encoder {
	return gecko.NopEncoder
}

func (d *AbcNetInputDevice) Topic() string {
	return d.topic
}

func (d *AbcNetInputDevice) receiveLoop(conn net.Conn, ctx gecko.Context, deliverer gecko.InputDeliverer) error {
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
				frame := gecko.NewFramePacket(buffer[:n])
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
