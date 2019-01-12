package gecko

import (
	"bytes"
	"parkingwang.com/go-conf"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 消息包载体，包含一个Header信息、数据体的数据实体。
// 用于设备对象的请求和响应。
type PacketFrame struct {
	// id
	id int64

	// Headers
	header *conf.ImmutableMap

	// 数据
	frame []byte
}

// 返回数据帧编号
func (pf *PacketFrame) Id() int64 {
	return pf.id
}

// 返回一个Reader接口
func (pf *PacketFrame) DataReader() *bytes.Reader {
	return bytes.NewReader(pf.frame)
}

// 返回Data字段的Reader接口
func (pf *PacketFrame) Data() []byte {
	return pf.frame
}

// 返回Header字段
func (pf *PacketFrame) Header() *conf.ImmutableMap {
	return pf.header
}

func NewPackFrame(id int64, header map[string]interface{}, frame []byte) *PacketFrame {
	return NewPackFrame0(id, conf.WrapImmutableMap(header), frame)
}

func NewPackFrame0(id int64, header *conf.ImmutableMap, frame []byte) *PacketFrame {
	return &PacketFrame{
		id:     id,
		header: header,
		frame:  frame,
	}
}
