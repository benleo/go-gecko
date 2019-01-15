package gecko

import (
	"bytes"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 消息包载体
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
