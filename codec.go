package gecko

import (
	"encoding/json"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Codes组件负责设备消息的编码和解码任务。
// 当InputDevice读取/接收到系统外部的消息数据时，需要Decoder对外部消息协议格式数据进行解码；
// 当系统处理结束后，给InputDevice返回结果消息时，需要使用Encoder来编码成外部设备的消息协议格式；
// 同样，OutputDevice在消息的传递过程中，也需要Encoder和Decoder来转换数据格式；

// 解码器，负责将字节数组转换成PacketMap对象
type Decoder func(bytes FramePacket) (JSONPacket, error)

// 扩展Decoder的函数
func (d Decoder) Decode(frame FramePacket) (JSONPacket, error) {
	return d(frame)
}

// 编码器，负责将PacketMap对象转换成字节数组
type Encoder func(data JSONPacket) (FramePacket, error)

// 扩展Encoder的函数
func (e Encoder) Encode(data JSONPacket) (FramePacket, error) {
	return e(data)
}

//// 系统默认实现的编码和解码接口

func NopEncoder(_ JSONPacket) (FramePacket, error) {
	return NewFramePacket([]byte{}), nil
}

func NopDecoder(_ FramePacket) (JSONPacket, error) {
	return NewJSONPacket(make(map[string]interface{}, 0)), nil
}

func JSONDefaultDecoderFactory() (string, CodecFactory) {
	return "JSONDefaultDecoder", func() interface{} {
		return Decoder(JSONDefaultDecoder)
	}
}

// 默认JSON解码器，将Byte数据解析成PacketMap对象
func JSONDefaultDecoder(bytes FramePacket) (JSONPacket, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal(bytes, &m)
	return m, err
}

func JSONDefaultEncoderFactory() (string, CodecFactory) {
	return "JSONDefaultEncoder", func() interface{} {
		return Encoder(JSONDefaultEncoder)
	}
}

// 默认JSON编码器，负责将PacketMap对象解析成Byte数组
func JSONDefaultEncoder(data JSONPacket) (FramePacket, error) {
	return json.Marshal(data)
}
