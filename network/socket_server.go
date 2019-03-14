package network

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yoojia/go-gecko"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

const (
	StateReady = iota
	StateClose
)

// 消息处理
type FrameHandler func(addr net.Addr, frame []byte) (resp []byte, err error)

// Socket客户端
type SocketServer struct {
	conn         net.Conn
	config       SocketConfig
	shutdown     context.Context
	shutdownFunc context.CancelFunc
	state        int32
}

func NewSocketServer() *SocketServer {
	c, f := context.WithCancel(context.Background())
	return &SocketServer{
		shutdown:     c,
		shutdownFunc: f,
		state:        StateReady,
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

func (ss *SocketServer) Serve(handler FrameHandler) error {
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

func (ss *SocketServer) udpServeLoop(udpConn *net.UDPConn, handler FrameHandler) error {
	// 收到终止服务端信号
	go func() {
		<-ss.shutdown.Done()
		if err := udpConn.Close(); nil != err {
			gecko.ZapSugarLogger.Errorf("UDP服务端关闭时发生错误", err)
		}
	}()
	err := ss.rwLoop("udp", udpConn, handler)
	if atomic.LoadInt32(&ss.state) == StateClose {
		return nil
	} else {
		return err
	}
}

func (ss *SocketServer) tcpServeLoop(server net.Listener, handler FrameHandler) error {
	zlog := gecko.ZapSugarLogger

	serve := func(clientAddr net.Addr, clientConn net.Conn) {
		err := ss.rwLoop("tcp", clientConn, handler)
		if nil != err && atomic.LoadInt32(&ss.state) == StateReady {
			zlog.Errorf("客户端中止通讯循环: %s", err)
		}
	}

	var delay time.Duration
	for {
		conn, err := server.Accept()
		if nil != err {
			// 检查关闭信号
			select {
			case <-ss.shutdown.Done():
				if err := server.Close(); nil != err {
					zlog.Errorf("TCP服务端关闭时发生错误", err)
				}
				return nil
			default:
			}
			// Accept超时，则适当延时
			if IsNetTempErr(err) {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay = 2 * delay
				}
				if max := time.Second; delay > max {
					delay = max
				}
				time.Sleep(delay)
				continue
			} else {

				return err
			}
		}
		addr := conn.RemoteAddr()
		zlog.Debugf("接受客户端连接: %s", addr)
		go serve(addr, conn)
	}
}

func (ss *SocketServer) rwLoop(protoType string, conn net.Conn, userHandler FrameHandler) error {
	listenAddr := conn.RemoteAddr()
	if "udp" == protoType {
		listenAddr = conn.LocalAddr()
	}
	zlog := gecko.ZapSugarLogger
	zlog.Debugf("开启数据通讯循环[%s]：%s", protoType, listenAddr)
	defer zlog.Debugf("中止数据通讯循环[%s]: %s", protoType, listenAddr)

	readFrame := func(c net.Conn, buf []byte, proto string) (n int, clientAddr net.Addr, err error) {
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
		err := conn.SetReadDeadline(time.Now().Add(readTimeout))
		if nil != err {
			if IsNetTempErr(err) {
				continue
			} else {
				return err
			}
		}

		n, clientAddr, err := readFrame(conn, buffer, protoType)
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

func (ss *SocketServer) Shutdown() {
	// 标记当前Server为Close状态
	atomic.StoreInt32(&ss.state, StateClose)
	// 发起Shutdown信号，其它依赖于shutdownContext的协程会中断内部循环
	ss.shutdownFunc()
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
