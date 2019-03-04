package network

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yoojia/go-gecko"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 消息处理
type PacketHandler func(addr net.Addr, frame []byte) (resp []byte, err error)

// Socket客户端
type SocketServer struct {
	conn            net.Conn
	config          SocketConfig
	closeServer     context.Context
	closeServerFunc context.CancelFunc
	closeState      int32
}

func NewSocketServer() *SocketServer {
	c, f := context.WithCancel(context.Background())
	return &SocketServer{
		closeServer:     c,
		closeServerFunc: f,
		closeState:      0,
	}
}

func (ss *SocketServer) Init(config SocketConfig) {
	ss.config = config
}

func (ss *SocketServer) Config() SocketConfig {
	return ss.config
}

func (ss *SocketServer) BufferSize() uint {
	return ss.config.BufferSize
}

func (ss *SocketServer) Serve(handler PacketHandler) error {
	networkType := ss.config.Type
	networkAddr := ss.config.Addr
	gecko.ZapSugarLogger.Debugf("启动服务端：Type=%s, Addr=%s", networkType, networkAddr)
	if strings.HasPrefix(networkType, "udp") {
		if conn, err := OpenUdpConn(networkAddr); nil != err {
			return err
		} else {
			return ss.udpServeLoop(conn, handler)
		}
	} else if strings.HasPrefix(networkType, "tcp") {
		if listener, err := OpenTcpListener(networkType, networkAddr); nil != err {
			return err
		} else {
			return ss.tcpServeLoop(listener, handler)
		}
	} else {
		return errors.New("未识别的网络连接模式: " + networkType)
	}
}

func (ss *SocketServer) udpServeLoop(udpConn *net.UDPConn, handler PacketHandler) error {
	go func() {
		// 收到终止服务端信号
		<-ss.closeServer.Done()
		if err := udpConn.Close(); nil != err {
			gecko.ZapSugarLogger.Errorf("UDP服务端关闭时发生错误", err)
		}
	}()
	err := ss.rwLoop("udp", udpConn, handler)
	if atomic.LoadInt32(&ss.closeState) > 0 {
		return nil
	} else {
		return err
	}
}

func (ss *SocketServer) tcpServeLoop(listen net.Listener, handler PacketHandler) error {
	clients := new(sync.Map)
	zlog := gecko.ZapSugarLogger
	go func() {
		<-ss.closeServer.Done()
		zlog.Debug("关闭客户端列表")
		clients.Range(func(_, c interface{}) bool {
			conn := c.(*net.TCPConn)
			if err := conn.Close(); nil != err {
				zlog.Errorf("TCP客户端关闭时发生错误", err)
			}
			return true
		})
		if err := listen.Close(); nil != err {
			zlog.Errorf("TCP服务端关闭时发生错误", err)
		}
	}()
	for {
		if conn, err := listen.Accept(); nil != err {
			if IsNetTempErr(err) {
				continue
			} else {
				if atomic.LoadInt32(&ss.closeState) > 0 {
					return nil
				} else {
					return err
				}
			}
		} else {
			addr := conn.RemoteAddr()
			clients.Store(addr, conn)
			zlog.Debugf("接受客户端连接: %s", addr)
			go func(clientAddr net.Addr, clientConn net.Conn) {
				defer clients.Delete(clientAddr) // 客户端主动中断时，删除注册
				if err := ss.rwLoop("tcp", clientConn, handler); nil != err {
					if atomic.LoadInt32(&ss.closeState) == 0 {
						zlog.Errorf("客户端中止通讯循环: %s", err)
					}
				}
			}(addr, conn)
		}
	}
}

func (ss *SocketServer) rwLoop(protoType string, conn net.Conn, userHandler PacketHandler) error {
	listenAddr := conn.RemoteAddr()
	if "udp" == protoType {
		listenAddr = conn.LocalAddr()
	}
	zlog := gecko.ZapSugarLogger
	zlog.Debugf("开启数据通讯循环[%s]：%s", protoType, listenAddr)
	defer zlog.Debugf("中止数据通讯循环[%s]: %s", protoType, listenAddr)

	readFromClient := func(c net.Conn, buf []byte, proto string) (n int, clientAddr net.Addr, err error) {
		if "udp" == proto {
			return c.(*net.UDPConn).ReadFromUDP(buf)
		} else /*if "tcp" == proto*/ {
			n, err := c.Read(buf)
			return n, c.RemoteAddr(), err
		}
	}

	buffer := make([]byte, ss.config.BufferSize)
	readTimeout := ss.config.ReadTimeout
	writeTimeout := ss.config.WriteTimeout

	for {
		if err := conn.SetReadDeadline(time.Now().Add(readTimeout)); nil != err {
			if IsNetTempErr(err) {
				continue
			} else {
				return err
			}
		}

		n, clientAddr, err := readFromClient(conn, buffer, protoType)
		if nil != err {
			if IsNetTempErr(err) {
				continue
			} else {
				return err
			}
		}

		if n <= 0 {
			continue
		}

		data, err := userHandler(clientAddr, buffer[:n])
		if err != nil {
			zlog.Errorw("用户处理函数内部错误", "err", err)
			continue
		}
		if len(data) <= 0 {
			continue
		}

		if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); nil != err {
			return err
		}
		if "udp" == protoType {
			if _, err := conn.(*net.UDPConn).WriteTo(data, clientAddr); nil != err {
				return err
			}
		} else /*if "tcp" == protoType*/ {
			if _, err := conn.Write(data); nil != err {
				return err
			}
		}
	}
}

func OpenUdpConn(addr string) (*net.UDPConn, error) {
	if udpAddr, err := net.ResolveUDPAddr("udp", addr); err != nil {
		return nil, errors.WithMessage(err, "无法创建UDP地址: "+addr)
	} else {
		if conn, err := net.ListenUDP("udp", udpAddr); nil != err {
			return nil, errors.WithMessage(err, "UDP连接监听失败: "+addr)
		} else {
			return conn, nil
		}
	}
}

func OpenTcpListener(network, addr string) (*net.TCPListener, error) {
	if tcpAddr, err := net.ResolveTCPAddr(network, addr); err != nil {
		return nil, errors.WithMessage(err, "无法创建TCP地址: "+addr)
	} else {
		if ln, err := net.ListenTCP(network, tcpAddr); nil != err {
			return nil, errors.WithMessage(err, "TCP连接监听失败: "+addr)
		} else {
			return ln, nil
		}
	}
}

func (ss *SocketServer) Shutdown() {
	atomic.StoreInt32(&ss.closeState, 1)
	ss.closeServerFunc()
}
