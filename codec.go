package gecko

import "github.com/yoojia/go-gecko/x"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Codes组件负责设备消息的编码和解码任务。
// 当InputDevice读取/接收到系统外部的消息数据时，需要Decoder对外部消息协议格式数据进行解码；
// 当系统处理结束后，给InputDevice返回结果消息时，需要使用Encoder来编码成外部设备的消息协议格式；
// 同样，OutputDevice在消息的传递过程中，也需要Encoder和Decoder来转换数据格式；

// 解码器，负责将字节数组转换成PacketMap对象
type Decoder func(bytes PacketFrame) (PacketMap, error)

// 扩展Decoder的函数
func (d Decoder) Decode(bytes PacketFrame) (PacketMap, error) {
	return d(bytes)
}

// 编码器，负责将PacketMap对象转换成字节数组
type Encoder func(data PacketMap) (PacketFrame, error)

// 扩展Encoder的函数
func (e Encoder) Encode(data PacketMap) (PacketFrame, error) {
	return e(data)
}

//// 系统默认实现的编码和解码接口

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
