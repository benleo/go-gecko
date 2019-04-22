package gecko

import (
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/yoojia/go-value"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Fields interface {
	// 迭代
	RangeFields(consumer func(name string, value interface{}))

	// 返回属性列表. 只读属性
	GetFields() map[string]interface{}

	// 添加属性
	AddField(key string, value interface{})

	// 获取属性
	GetField(key string) (interface{}, bool)

	// 获取属性值, 如果不存在,返回 nil
	GetFieldOrNil(key string) interface{}

	// 获取String类型属性
	GetFieldString(key string) (string, bool)

	// 获取Int64类型的属性
	GetFieldInt64(key string) (int64, bool)

	// 判断属性是否存在
	HasField(key string) bool
}

type fieldsMap struct {
	Fields
	data map[string]interface{}
}

func (a *fieldsMap) RangeFields(consumer func(name string, value interface{})) {
	for k, v := range a.data {
		consumer(k, v)
	}
}

func (a *fieldsMap) GetFields() map[string]interface{} {
	rom := make(map[string]interface{})
	for k, v := range a.data {
		rom[k] = v
	}
	return rom
}

func (a *fieldsMap) AddField(key string, value interface{}) {
	a.data[key] = value
}

func (a *fieldsMap) GetField(key string) (interface{}, bool) {
	v, ok := a.data[key]
	return v, ok
}

func (a *fieldsMap) GetFieldOrNil(key string) interface{} {
	v, ok := a.data[key]
	if ok {
		return v
	} else {
		return nil
	}
}

func (a *fieldsMap) GetFieldString(key string) (string, bool) {
	val, ok := a.data[key]
	if ok {
		return value.ToString(val), true
	} else {
		return "", false
	}
}

func (a *fieldsMap) GetFieldInt64(key string) (int64, bool) {
	val, ok := a.data[key]
	if ok {
		return value.ToInt64(val)
	} else {
		return 0, false
	}
}

func (a *fieldsMap) HasField(key string) bool {
	_, ok := a.data[key]
	return ok
}

func newMapFields() *fieldsMap {
	return &fieldsMap{data: make(map[string]interface{})}
}

// 对象数据消息包
type MessagePacket struct {
	*fieldsMap
	frames []byte
}

func (m *MessagePacket) GetFrames() []byte {
	return m.frames
}

func (m *MessagePacket) GetFramesStr() string {
	return string(m.frames)
}

func (m *MessagePacket) SetFrames(b []byte) *MessagePacket {
	m.frames = b
	return m
}

func NewMessagePacketWith(fields map[string]interface{}, frames []byte) *MessagePacket {
	m := newMapFields()
	for k, v := range fields {
		m.data[k] = v
	}
	return &MessagePacket{
		fieldsMap: m,
		frames:    frames,
	}
}

func NewMessagePacketFields(fields map[string]interface{}) *MessagePacket {
	return NewMessagePacketWith(fields, nil)
}

func NewMessagePacketFrames(frames []byte) *MessagePacket {
	return NewMessagePacketWith(make(map[string]interface{}), frames)
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
	return json.Marshal(msg.data)
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
	return msg.frames, nil
}

func FrameDefaultEncoderFactory() (string, CodecFactory) {
	return "FrameDefaultEncoder", func() interface{} {
		return Encoder(FrameDefaultEncoder)
	}
}
