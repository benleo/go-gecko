package gecko

import "github.com/yoojia/go-gecko/x"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 解码器，负责将字节数组转换成PacketMap对象
type Decoder func(bytes PacketFrame) (PacketMap, error)

func (d Decoder) Decode(bytes PacketFrame) (PacketMap, error) {
	return d(bytes)
}

// 编码器，负责将PacketMap对象转换成字节数组
type Encoder func(data PacketMap) (PacketFrame, error)

func (e Encoder) Encode(data PacketMap) (PacketFrame, error) {
	return e(data)
}

////

func JSONDefaultDecoderFactory() (string, CodecFactory) {
	return "JSONDefaultDecoder", func() interface{} {
		return Decoder(JSONDefaultDecoder)
	}
}

// 默认JSON解码器，将Byte数据解析成PacketMap对象
func JSONDefaultDecoder(bytes PacketFrame) (PacketMap, error) {
	json := make(map[string]interface{})
	err := x.UnmarshalJSON(bytes, &json)
	return json, err
}

func JSONDefaultEncoderFactory() (string, CodecFactory) {
	return "JSONDefaultEncoder", func() interface{} {
		return Encoder(JSONDefaultEncoder)
	}
}

// 默认JSON编码器，负责将PacketMap对象解析成Byte数组
func JSONDefaultEncoder(data PacketMap) (PacketFrame, error) {
	return x.MarshalJSON(data)
}
