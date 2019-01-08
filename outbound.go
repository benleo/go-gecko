package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Outbound 作为Gecko的输出对象；
type Outbound struct {
	// Topic
	Topic string
	// 数据字段
	Data map[string]interface{}
}

// 添加Data字段。相同的Name将会被覆盖。
func (out *Outbound) AddDataField(name string, value interface{}) *Outbound {
	out.Data[name] = value
	return out
}
