package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Message 作为Gecko内部传递数据的模型。
type Message struct {
	// 当前Topic
	Topic string
	// 数据字段
	Data map[string]interface{}
}

// 添加Data字段。相同的Name将会被覆盖。
func (out *Message) AddDataField(name string, value interface{}) *Message {
	out.Data[name] = value
	return out
}
