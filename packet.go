package gecko

import (
	"bytes"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// JSON结构数据消息包
type JSONPacket map[string]interface{}

func (pm JSONPacket) Add(key string, value interface{}) {
	pm[key] = value
}

func NewJSONPacket(m map[string]interface{}) JSONPacket {
	return JSONPacket(m)
}

func NewJSONPacketCapacity(capacity int) JSONPacket {
	return NewJSONPacket(make(map[string]interface{}, capacity))
}

////

// 字节数据消息包
type FramePacket []byte

// 返回一个Reader接口
func (p FramePacket) DataReader() *bytes.Reader {
	return bytes.NewReader(p)
}

// 返回Data数据
func (p FramePacket) Data() []byte {
	return p
}

func NewFramePacket(frame []byte) FramePacket {
	return FramePacket(frame)
}

func NewFramePacketSize(size int) FramePacket {
	return NewFramePacket(make([]byte, size))
}
