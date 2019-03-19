package gecko

import (
	"encoding/json"
	"github.com/pkg/errors"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 对象数据消息包
type MessagePacket map[string]interface{}

func (m MessagePacket) AddField(name string, value interface{}) MessagePacket {
	m[name] = value
	return m
}

func (m MessagePacket) Field(name string) (interface{}, bool) {
	v, ok := m[name]
	return v, ok
}

// 字节数据消息包
type FramePacket []byte

// Codes组件负责设备消息的编码和解码任务。
// 当InputDevice读取/接收到系统外部的消息数据时，需要Decoder对外部消息协议格式数据进行解码；
// 当系统处理结束后，给InputDevice返回结果消息时，需要使用Encoder来编码成外部设备的消息协议格式；
// 同样，OutputDevice在消息的传递过程中，也需要Encoder和Decoder来转换数据格式；

// 解码器，负责将字节数组转换成JSONPacket对象
type Decoder func(bytes FramePacket) (MessagePacket, error)

// 扩展Decoder的函数
func (d Decoder) Decode(frame FramePacket) (MessagePacket, error) {
	return d(frame)
}

// 编码器，负责将JSONPacket对象转换成字节数组
type Encoder func(data MessagePacket) (FramePacket, error)

// 扩展Encoder的函数
func (e Encoder) Encode(data MessagePacket) (FramePacket, error) {
	return e(data)
}

//// 系统默认实现的编码和解码接口

func NopEncoder(_ MessagePacket) (FramePacket, error) {
	return FramePacket([]byte{}), nil
}

func NopDecoder(_ FramePacket) (MessagePacket, error) {
	return MessagePacket(make(map[string]interface{}, 0)), nil
}

func JSONDefaultDecoderFactory() (string, CodecFactory) {
	return "JSONDefaultDecoder", func() interface{} {
		return Decoder(JSONDefaultDecoder)
	}
}

// 默认JSON解码器，将Byte数据解析成map[string]interface{}对象
func JSONDefaultDecoder(bytes FramePacket) (MessagePacket, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal(bytes, &m)
	return m, errors.Wrap(err, "json default decode failed")
}

func JSONDefaultEncoderFactory() (string, CodecFactory) {
	return "JSONDefaultEncoder", func() interface{} {
		return Encoder(JSONDefaultEncoder)
	}
}

// 默认JSON编码器，负责将map[string]interface{}对象解析成Byte数组
func JSONDefaultEncoder(data MessagePacket) (FramePacket, error) {
	return json.Marshal(data)
}
