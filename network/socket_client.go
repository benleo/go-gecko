package network

import (
	"github.com/pkg/errors"
	"github.com/yoojia/go-gecko"
	"net"
	"sync"
	"time"
)

// Socket客户端
type SocketClient struct {
	conn   net.Conn
	config SocketConfig
	chLock *sync.Mutex
}

func NewSocketClient() *SocketClient {
	return &SocketClient{
		chLock: new(sync.Mutex),
	}
}

func (s *SocketClient) Init(config SocketConfig) {
	s.config = config
}

func (s *SocketClient) Config() SocketConfig {
	return s.config
}

func (s *SocketClient) BufferSize() uint {
	return s.config.BufferSize
}

// Open 创建数据连接
func (s *SocketClient) Open() error {
	switch s.config.Type {
	case "tcp", "TCP":
		if conn, err := net.Dial("tcp", s.config.Addr); nil != err {
			return errors.WithMessage(err, "TCP dial failed")
		} else {
			s.conn = conn
			return nil
		}

	case "udp", "UDP":
		if addr, err := net.ResolveUDPAddr("udp", s.config.Addr); nil != err {
			return errors.WithMessage(err, "Resolve udp address failed")
		} else if conn, err := net.DialUDP("udp", nil, addr); nil != err {
			return errors.WithMessage(err, "UDP dial failed")
		} else {
			s.conn = conn
			return nil
		}

	default:
		return errors.New("Unknown network type: " + s.config.Type)
	}
}

// Close 关闭数据连接
func (s *SocketClient) Close() error {
	if nil != s.conn {
		return s.conn.Close()
	} else {
		return nil
	}
}

// Receive 从数据连接中读取数据
func (s *SocketClient) Receive(buff []byte) (n int, err error) {
	if s.conn == nil {
		return 0, errors.New("Client connection is not ready")
	}
	if err := s.conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout)); nil != err {
		return 0, errors.WithMessage(err, "Set read timeout failed")
	}
	return s.conn.Read(buff)
}

// Send 向数据连接发送数据
func (s *SocketClient) Send(data []byte) (n int, err error) {
	if s.conn == nil {
		return 0, errors.New("Client connection is not ready")
	}
	if err := s.conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout)); nil != err {
		return 0, errors.WithMessage(err, "Set write timeout failed")
	}
	return s.conn.Write(data)
}

// Execute 是一个线程安全的Send-Receive原子操作
func (s *SocketClient) Execute(frame []byte) ([]byte, error) {
	s.chLock.Lock()
	defer s.chLock.Unlock()
	buffer := make([]byte, s.BufferSize())
	if _, err := s.Send(frame); nil != err {
		return nil, err
	}
	if n, err := s.Receive(buffer); nil != err {
		return nil, err
	} else {
		return gecko.FramePacket(buffer[:n]), nil
	}
}
