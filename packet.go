package gecko

import (
	"bytes"
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
	header map[string]interface{}

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
func (pf *PacketFrame) Header() map[string]interface{} {
	return pf.header
}
