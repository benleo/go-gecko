package bundles

import (
	conf2 "github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-gecko"
	"net"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 支持UDP通讯的设备。
// 在此处，UDP设备作为客户端，另一方作为服务端，接收此Device的消息，并返回响应。
// 使用UDP设备时，需要指定：
// - GroupAddress：目标UDP通讯设备的IP地址；
// - PhysicalAddress: 目标UDP通讯设备的端口；
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
	//
	shutdownCompleted chan struct{}
}

func (udv *UdpVirtualDevice) GetProtoName() string {
	return "udp"
}

// 初始化
func (udv *UdpVirtualDevice) OnInit(args map[string]interface{}, scoped gecko.GeckoScoped) {
	udv.sendChan = make(chan *gecko.PacketFrame, 1)
	udv.recvChan = make(chan *gecko.PacketFrame, 1)
	udv.shutdownCompleted = make(chan struct{}, 1)
	conf := conf2.MapToMap(args)
	udv.sendBuffSize = conf.GetInt64OrDefault("sendBuffSizeKB", 1) * 1024
	udv.recvBuffSize = conf.GetInt64OrDefault("recvBuffSizeKB", 1) * 1024
	udv.sendTimeout = conf.GetDurationOrDefault("sendTimeout", time.Second*3)
	udv.recvTimeout = conf.GetDurationOrDefault("recvTimeout", time.Second*3)
	// 使用GroupAddress作为UDP地址
	// 使用PhysicalAddress作为UDP端口
	udpAddr := udv.GetGroupAddress() + ":" + udv.GetPhyAddress()
	if addr, err := net.ResolveUDPAddr("udp", udpAddr); nil != err {
		log.Panic().Err(err).Msgf("非法的UDP地址：" + udpAddr)
	} else {
		udv.srvAddr = addr
	}
	log.Info().Msg("初始化UDP虚拟设备客户端: " + udpAddr)
}

// 启动设备
func (udv *UdpVirtualDevice) OnStart(scoped gecko.GeckoScoped) {
	if conn, err := net.DialUDP("udp", nil, udv.srvAddr); nil != err {
		log.Panic().Err(err).Msgf("无法建立UDP连接: " + udv.srvAddr.String())
	} else {
		conn.SetReadBuffer(int(udv.recvBuffSize))
		conn.SetWriteBuffer(int(udv.sendBuffSize))
		udv.udpConn = conn

	}
	// 读写协程
	go func(shouldBreak <-chan struct{}) {
		for {
			select {
			case <-shouldBreak:
				break

			case frame := <-udv.sendChan:
				udv.udpConn.SetWriteDeadline(time.Now().Add(udv.sendTimeout))
				if _, err := udv.udpConn.Write(frame.Bytes); nil != err {
					log.Error().Err(err).Msgf("发送数据到UDP服务端错误: " + udv.srvAddr.String())
				}
				// TODO 读取处理结果
				udv.udpConn.SetWriteDeadline(time.Now().Add(udv.recvTimeout))
				//if a, err := udv.udpConn.Read

			}
		}
	}(udv.shutdownCompleted)
}

// 停止设备
func (udv *UdpVirtualDevice) OnStop(scoped gecko.GeckoScoped) {
	if udv.udpConn != nil {
		udv.udpConn.Close()
	}
	udv.shutdownCompleted <- struct{}{}
}

// 处理远程控制事件
func (udv *UdpVirtualDevice) Process(frame *gecko.PacketFrame, scoped gecko.GeckoScoped) (*gecko.PacketFrame, error) {
	// 将请求发送到SendBuf，等待接收RecvBuf
	udv.sendChan <- frame

	return nil, nil
}
