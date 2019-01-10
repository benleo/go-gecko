package bundles

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"io"
	"net"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func UdpVirtualDeviceFactory() (string, gecko.BundleFactory) {
	return "UdpVirtualDevice", func() interface{} {
		return &UdpVirtualDevice{
			AbcVirtualDevice: new(gecko.AbcVirtualDevice),
		}
	}
}

// UDP虚拟设备
type UdpVirtualDevice struct {
	*gecko.AbcVirtualDevice
	//
	srvAddr      *net.UDPAddr
	udpConn      *net.UDPConn
	sendBuffSize int64
	recvBuffSize int64
	sendTimeout  time.Duration
	recvTimeout  time.Duration
	sendChan     chan *gecko.PacketFrame
	recvChan     chan *gecko.PacketFrame
	errChan      chan error
	//
	shutdownCompleted chan struct{}
}

func (uv *UdpVirtualDevice) GetProtoName() string {
	return "udp"
}

// 初始化
func (uv *UdpVirtualDevice) OnInit(args map[string]interface{}, ctx gecko.Context) {
	uv.sendChan = make(chan *gecko.PacketFrame, 1)
	uv.recvChan = make(chan *gecko.PacketFrame, 1)
	uv.errChan = make(chan error, 1)
	uv.shutdownCompleted = make(chan struct{}, 1)
	config := conf.MapToMap(args)
	uv.sendBuffSize = config.GetInt64OrDefault("sendBuffSizeKB", 1) * 1024
	uv.recvBuffSize = config.GetInt64OrDefault("recvBuffSizeKB", 1) * 1024
	uv.sendTimeout = config.GetDurationOrDefault("sendTimeout", time.Second*3)
	uv.recvTimeout = config.GetDurationOrDefault("recvTimeout", time.Second*3)
	// 使用GroupAddress作为UDP地址
	// 使用PhysicalAddress作为UDP端口
	udpAddr := uv.GetGroupAddress() + ":" + uv.GetPhyAddress()
	if addr, err := net.ResolveUDPAddr("udp", udpAddr); nil != err {
		uv.withTag(log.Panic).Err(err).Msgf("非法的UDP地址：" + udpAddr)
	} else {
		uv.srvAddr = addr
	}
	uv.withTag(log.Info).Msg("初始化UDP虚拟设备，通讯目标: " + udpAddr)
}

// 启动，建立与目标设备的连接
func (uv *UdpVirtualDevice) OnStart(ctx gecko.Context) {
	remoteAddr := uv.srvAddr.String()
	if conn, err := net.DialUDP("udp", nil, uv.srvAddr); nil != err {
		uv.withTag(log.Panic).Err(err).Str("remote", remoteAddr).Msgf("无法建立与目标设备UDP通讯连接： %s", remoteAddr)
	} else {
		conn.SetReadBuffer(int(uv.recvBuffSize))
		conn.SetWriteBuffer(int(uv.sendBuffSize))
		uv.udpConn = conn
	}
	// 读写处理
	go func(shouldBreak <-chan struct{}) {
		for {
			select {
			case <-shouldBreak:
				break

			case incoming := <-uv.sendChan:
				// 写
				uv.udpConn.SetWriteDeadline(time.Now().Add(uv.sendTimeout))
				if _, wErr := io.Copy(uv.udpConn, incoming.DataReader()); nil != wErr {
					uv.withTag(log.Error).Err(wErr).Str("remote", remoteAddr).Msgf("发送数据到目标设备出错:" + remoteAddr)
					uv.errChan <- wErr
				} else {
					// 读
					uv.udpConn.SetReadDeadline(time.Now().Add(uv.recvTimeout))
					buffer := make([]byte, uv.recvBuffSize)
					if n, rErr := uv.udpConn.Read(buffer); nil != rErr {
						uv.withTag(log.Error).Err(wErr).Str("remote", remoteAddr).Msgf("从目标设备读取响应数据出错:" + remoteAddr)
						uv.errChan <- rErr
					} else {
						uv.recvChan <- gecko.NewPackFrame(
							incoming.Id(),
							make(map[string]interface{}, 0),
							buffer[:n])
					}
				}
			}
		}
	}(uv.shutdownCompleted)
}

// 停止，关闭UDP通讯连接
func (uv *UdpVirtualDevice) OnStop(ctx gecko.Context) {
	if uv.udpConn != nil {
		uv.udpConn.Close()
	}
	uv.shutdownCompleted <- struct{}{}
}

// 处理事件
func (uv *UdpVirtualDevice) Process(in *gecko.PacketFrame, ctx gecko.Context) (out *gecko.PacketFrame, err error) {
	// Send and receive
	uv.sendChan <- in
	select {
	case out = <-uv.recvChan:
	case err = <-uv.errChan:
	}
	return
}

func (uv *UdpVirtualDevice) withTag(fun func() *zerolog.Event) *zerolog.Event {
	return fun().Str("tag", "NetworkServerTrigger")
}
