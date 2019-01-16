package gecko

import "github.com/yoojia/go-gecko/x"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 解码器，负责将字节数组转换成JSON Map对象
type Decoder func(bytes []byte) (map[string]interface{}, error)

// 编码器，负责将JSON Map对象转换成字节数组
type Encoder func(data map[string]interface{}) ([]byte, error)

////

func JSONDefaultDecoderFactory() (string, CodecFactory) {
	return "JSONDefaultDecoder", func() interface{} {
		return Decoder(JSONDefaultDecoder)
	}
}

// 默认JSON解码器，将Byte数据解析成JSON Map对象
func JSONDefaultDecoder(bytes []byte) (map[string]interface{}, error) {
	json := make(map[string]interface{})
	err := x.UnmarshalJSON(bytes, &json)
	return json, err
}

func JSONDefaultEncoderFactory() (string, CodecFactory) {
	return "JSONDefaultEncoder", func() interface{} {
		return Encoder(JSONDefaultEncoder)
	}
}

// 默认JSON编码器，负责将JSON Map对象解析成Byte数组
func JSONDefaultEncoder(data map[string]interface{}) ([]byte, error) {
	return x.MarshalJSON(data)
}
