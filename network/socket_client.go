package network

import (
	"github.com/pkg/errors"
	"net"
	"time"
)

// Socket配置
type SocketConfig struct {
	Type         string        // 网络类型
	Addr         string        // 地址
	ReadTimeout  time.Duration // 读超时
	WriteTimeout time.Duration // 写超时
	BufferSize   uint          // 读写缓存大小
}

func (c SocketConfig) IsValid() bool {
	return c.Type != "" && c.Addr != ""
}

type SocketClient struct {
	conn   net.Conn
	config SocketConfig
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

func (s *SocketClient) Open() error {
	if "tcp" == s.config.Type {
		if conn, err := net.Dial("tcp", s.config.Addr); nil != err {
			return errors.WithMessage(err, "TCP dial failed")
		} else {
			s.conn = conn
			return nil
		}
	} else if "udp" == s.config.Type {
		if addr, err := net.ResolveUDPAddr("udp", s.config.Addr); nil != err {
			return errors.WithMessage(err, "Resolve udp address failed")
		} else if conn, err := net.DialUDP("udp", nil, addr); nil != err {
			return errors.WithMessage(err, "UDP dial failed")
		} else {
			s.conn = conn
			return nil
		}
	} else {
		return errors.New("Unknown network type: " + s.config.Type)
	}
}

func (s *SocketClient) Receive(buff []byte) (n int, err error) {
	if s.conn == nil {
		return 0, errors.New("Client connection is not ready")
	}
	if err := s.conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout)); nil != err {
		return 0, errors.WithMessage(err, "Set read timeout failed")
	}
	return s.conn.Read(buff)
}

func (s *SocketClient) Send(data []byte) (n int, err error) {
	if s.conn == nil {
		return 0, errors.New("Client connection is not ready")
	}
	if err := s.conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout)); nil != err {
		return 0, errors.WithMessage(err, "Set write timeout failed")
	}
	return s.conn.Write(data)
}

func (s *SocketClient) Close() error {
	if nil != s.conn {
		return s.conn.Close()
	} else {
		return nil
	}
}
