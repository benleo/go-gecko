package gecko

import (
	"encoding/json"
	"github.com/pkg/errors"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 对象数据消息包
type MessagePacket struct {
	Fields map[string]interface{}
	Frames []byte
}

func (m *MessagePacket) AddField(name string, value interface{}) *MessagePacket {
	m.Fields[name] = value
	return m
}

func (m *MessagePacket) GetField(name string) (interface{}, bool) {
	v, ok := m.Fields[name]
	return v, ok
}

func (m *MessagePacket) Field(name string) interface{} {
	return m.Fields[name]
}

func (m *MessagePacket) GetFrames() []byte {
	return m.Frames
}

func (m *MessagePacket) GetFramesStr() string {
	return string(m.Frames)
}

func (m *MessagePacket) SetFrames(b []byte) *MessagePacket {
	m.Frames = b
	return m
}

func NewMessagePacketWith(fields map[string]interface{}, body []byte) *MessagePacket {
	return &MessagePacket{
		Fields: fields,
		Frames: body,
	}
}

func NewMessagePacketFields(fields map[string]interface{}) *MessagePacket {
	return NewMessagePacketWith(fields, nil)
}

func NewMessagePacket() *MessagePacket {
	return NewMessagePacketWith(
		make(map[string]interface{}),
		nil)
}

func NewEmptyMessagePack() *MessagePacket {
	return NewMessagePacketWith(
		make(map[string]interface{}, 0),
		nil)
}

////

// 字节数据消息包
type FramePacket []byte

// Codes组件负责设备消息的编码和解码任务。
// 当InputDevice读取/接收到系统外部的消息数据时，需要Decoder对外部消息协议格式数据进行解码；
// 当系统处理结束后，给InputDevice返回结果消息时，需要使用Encoder来编码成外部设备的消息协议格式；
// 同样，OutputDevice在消息的传递过程中，也需要Encoder和Decoder来转换数据格式；

// 解码器，负责将字节数组转换成JSONPacket对象
type Decoder func(frames FramePacket) (*MessagePacket, error)

// 扩展Decoder的函数
func (d Decoder) Decode(frames FramePacket) (*MessagePacket, error) {
	return d(frames)
}

// 编码器，负责将JSONPacket对象转换成字节数组
type Encoder func(data *MessagePacket) (FramePacket, error)

// 扩展Encoder的函数
func (e Encoder) Encode(data *MessagePacket) (FramePacket, error) {
	return e(data)
}

//// 系统默认实现的编码和解码接口

func NopEncoder(_ *MessagePacket) (FramePacket, error) {
	return FramePacket([]byte{}), nil
}

func NopDecoder(_ FramePacket) (*MessagePacket, error) {
	return NewEmptyMessagePack(), nil
}

////

func JSONDefaultDecoderFactory() (string, CodecFactory) {
	return "JSONDefaultDecoder", func() interface{} {
		return Decoder(JSONDefaultDecoder)
	}
}

// 默认JSON解码器，将Byte数据解析成MessagePacket.Fields对象
// 注意：JSON解码器会将JSON数据解析到Fields字段中，原始数据保存在Frame字段；
func JSONDefaultDecoder(bytes FramePacket) (*MessagePacket, error) {
	fields := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &fields); nil != err {
		return nil, errors.Wrap(err, "json default decode failed")
	} else {
		return NewMessagePacketWith(fields, bytes), nil
	}
}

func JSONDefaultEncoderFactory() (string, CodecFactory) {
	return "JSONDefaultEncoder", func() interface{} {
		return Encoder(JSONDefaultEncoder)
	}
}

// 默认JSON编码器，负责将MessagePacket.Fields对象解析成Byte数组
// 注意：JSON编码器将Fields字段转换成JSON字节数组，忽略Frames字段。
func JSONDefaultEncoder(msg *MessagePacket) (FramePacket, error) {
	return json.Marshal(msg.Fields)
}

////

// 字节帧解码器，负责将FramePacket打包到MessagePacket.Frame字段。其中Fields字段创建默认空对象。
func FrameDefaultDecoder(frame FramePacket) (*MessagePacket, error) {
	return NewMessagePacketWith(make(map[string]interface{}, 0), frame), nil
}

func FrameDefaultDecoderFactory() (string, CodecFactory) {
	return "FrameDefaultDecoder", func() interface{} {
		return Decoder(FrameDefaultDecoder)
	}
}

// 字节帧编码器，负责将MessagePacket.Frames转换成FramePacket。其中Fields字段被忽略。
func FrameDefaultEncoder(msg *MessagePacket) (FramePacket, error) {
	return msg.Frames, nil
}

func FrameDefaultEncoderFactory() (string, CodecFactory) {
	return "FrameDefaultEncoder", func() interface{} {
		return Encoder(FrameDefaultEncoder)
	}
}
