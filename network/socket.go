package network

import (
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

////

func IsNetTempErr(err error) bool {
	if nErr, ok := err.(net.Error); ok {
		return nErr.Timeout() || nErr.Temporary()
	} else {
		return false
	}
}
