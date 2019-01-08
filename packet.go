package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 消息包载体，包含一个Header信息、数据体的数据实体。
// 用于设备对象的请求和响应。
type PacketFrame struct {
	// Id
	Id int64

	// Headers
	Header map[string]string

	// 数据
	Bytes []byte
}

func NewPacketFrame(id int64, header map[string]string, bytes []byte) *PacketFrame {
	return &PacketFrame{
		Id:     id,
		Header: header,
		Bytes:  bytes,
	}
}
