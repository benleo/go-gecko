package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Event事件载体，包含一个Header信息、数据体的数据实体。
// 用于设备对象的请求和响应。
type EventFrame struct {
	// Id
	Id int64

	// Headers
	Header map[string]string

	// 数据
	Bytes []byte
}

// 创建EventFrame
func NewEventFrame(id int64, header map[string]string, bytes []byte) *EventFrame {
	return &EventFrame{
		Id:     id,
		Header: header,
		Bytes:  bytes,
	}
}
