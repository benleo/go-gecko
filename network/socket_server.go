package network

import (
	"context"
	"errors"
	"github.com/yoojia/go-gecko"
	"net"
	"sync"
	"time"
)

// 消息处理
type PacketHandler func(addr net.Addr, frame []byte) (resp []byte)

// Socket客户端
type SocketServer struct {
	conn         net.Conn
	config       SocketConfig
	closeContext context.Context
	closeFunc    context.CancelFunc
}

func (s *SocketServer) Init(config SocketConfig) {
	s.config = config
}

func (s *SocketServer) Config() SocketConfig {
	return s.config
}

func (s *SocketServer) BufferSize() uint {
	return s.config.BufferSize
}

func (s *SocketServer) Serve(handler PacketHandler) error {
	s.closeContext, s.closeFunc = context.WithCancel(context.Background())
	if "udp" == s.config.Type {
		if conn, err := OpenUdpConn(s.config.Addr); nil != err {
			return err
		} else {
			return s.acceptor(conn, handler)
		}
	} else if "tcp" == s.config.Type {
		if listener, err := net.Listen("tcp", s.config.Addr); nil != err {
			return err
		} else {
			return s.tcpAcceptLoop(listener, handler)
		}
	} else {
		return errors.New("未识别的网络连接模式: " + s.config.Type)
	}
}

func (s *SocketServer) tcpAcceptLoop(listen net.Listener, handler PacketHandler) error {
	clients := new(sync.Map)
	wg := new(sync.WaitGroup)
	for {
		select {
		case <-s.closeContext.Done():
			zlog := gecko.ZapSugarLogger
			clients.Range(func(_, c interface{}) bool {
				conn := c.(net.Conn)
				if err := conn.Close(); nil != err {
					zlog.Errorf("客户端关闭时发生错误", err)
				}
				return true
			})
			if err := listen.Close(); nil != err {
				zlog.Errorf("服务端关闭时发生错误", err)
			}
			return nil

		default:
			if conn, err := listen.Accept(); nil != err {
				if IsNetTempErr(err) {
					continue
				} else {
					return err
				}
			} else {
				addr := conn.RemoteAddr()
				clients.Store(addr, conn)
				wg.Add(1)
				zlog := gecko.ZapSugarLogger
				go func() {
					defer wg.Done()
					if err := s.acceptor(conn, handler); nil != err {
						zlog.Errorf("客户端中止通讯循环", err)
					}
				}()
			}
		}
	}
}

func (s *SocketServer) acceptor(conn net.Conn, handler PacketHandler) error {
	buffer := make([]byte, s.config.BufferSize)
	readTimeout := s.config.ReadTimeout
	writeTimeout := s.config.WriteTimeout
	for {
		if err := conn.SetReadDeadline(time.Now().Add(readTimeout));
			nil != err && !IsNetTempErr(err) {
			return err
		}

		if n, err := conn.Read(buffer); nil != err {
			if !IsNetTempErr(err) {
				return err
			}
		} else if n > 0 {
			data := handler(conn.RemoteAddr(), buffer[:n])
			if len(data) > 0 {
				if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); nil != err {
					return err
				}
				if _, err := conn.Write(data); nil != err {
					return err
				}
			}
		}
	}
}

func OpenUdpConn(addr string) (net.Conn, error) {
	if udpAddr, err := net.ResolveUDPAddr("udp", addr); err != nil {
		return nil, errors.New("无法创建UDP地址: " + addr)
	} else {
		if conn, err := net.ListenUDP("udp", udpAddr); nil != err {
			return nil, errors.New("UDP连接监听失败: " + addr)
		} else {
			return conn, nil
		}
	}
}

func (s *SocketServer) Close() {
	s.closeFunc()
}
