package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Inbound 作为Gecko输入数据模型。
type Inbound struct {
	// 当前Topic
	Topic string
	// 数据字段
	Data map[string]interface{}
}
