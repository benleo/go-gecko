package gecko

import (
	"bytes"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// JSON结构数据消息包
type PacketMap map[string]interface{}

func (pm PacketMap) Add(key string, value interface{}) {
	pm[key] = value
}

func NewPacketMap(m map[string]interface{}) PacketMap {
	return PacketMap(m)
}

////

// 字节数据消息包
type PacketFrame []byte

// 返回一个Reader接口
func (pf PacketFrame) DataReader() *bytes.Reader {
	return bytes.NewReader(pf)
}

// 返回Data数据
func (pf PacketFrame) Data() []byte {
	return pf
}

func NewPackFrame(frame []byte) PacketFrame {
	return PacketFrame(frame)
}
