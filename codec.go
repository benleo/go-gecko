package gecko

import "github.com/yoojia/go-gecko/x"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 解码器
type Decoder func(bytes []byte) (map[string]interface{}, error)

// 编码器
type Encoder func(data map[string]interface{}) ([]byte, error)

////

// 将Byte数据解析成JSON对象
func JSONDefaultDecoder(bytes []byte) (map[string]interface{}, error) {
	json := make(map[string]interface{})
	err := x.UnmarshalJSON(bytes, json)
	return json, err
}

// 将JSON对象解析成Bytes对象
func JSONDefaultEncoder(data map[string]interface{}) ([]byte, error) {
	return x.MarshalJSON(data)
}
