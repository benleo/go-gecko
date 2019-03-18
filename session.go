package gecko

import (
	"github.com/parkingwang/go-conf"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

const (
	DataTypeMap = iota
	DataTypeBytes
	DataTypeObject
)

// Inbound是Input的输入数据模型，对用户而言是只读属性。
// 数据类型根据
type Inbound struct {
	mapData    map[string]interface{}
	bytesData  []byte
	objectData interface{}
	dataType   int
}

func (i *Inbound) DataType() int {
	return i.dataType
}

func (i *Inbound) MapData() map[string]interface{} {
	return i.mapData
}

func (i *Inbound) MapField(name string) (interface{}, bool) {
	v, ok := i.mapData[name]
	return v, ok
}

func (i *Inbound) MapFields(iterator func(name string, value interface{})) {
	for k, v := range i.mapData {
		iterator(k, v)
	}
}

func (i *Inbound) BytesData() []byte {
	return i.bytesData
}

func (i *Inbound) ObjectData() interface{} {
	return i.objectData
}

func newInbound(data ObjectPacket) *Inbound {
	switch data.(type) {
	case map[string]interface{}:
		return &Inbound{
			mapData:  data.(map[string]interface{}),
			dataType: DataTypeMap,
		}

	case []byte:
		return &Inbound{
			bytesData: data.([]byte),
			dataType:  DataTypeBytes,
		}

	default:
		return &Inbound{
			objectData: data,
			dataType:   DataTypeObject,
		}
	}
}

////

// Message 作为Gecko内部传递数据的模型。
type Outbound struct {
	// 数据字段
	Data map[string]interface{}
}

// 添加Data字段。相同的Name将会被覆盖。
func (out *Outbound) AddField(name string, value interface{}) *Outbound {
	out.Data[name] = value
	return out
}

func newOutbound() *Outbound {
	return &Outbound{
		Data: make(map[string]interface{}),
	}
}

////

// EventSession 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type EventSession interface {
	// 返回属性列表
	Attrs() *cfg.Config

	// 添加属性
	AddAttr(key string, value interface{})

	// 添加多个属性。相同Key的属性将被覆盖。
	AddAttrs(attributes map[string]interface{})

	// 获取属性
	GetAttr(key string) (interface{}, bool)

	// 创建的时间戳
	Timestamp() time.Time

	// 从创建起始，到当前时间的用时
	Since() time.Duration

	// 当前事件的Topic
	Topic() string

	// 当前事件的设备地址
	Uuid() string

	// 返回输入端消息对象
	Inbound() *Inbound

	// 返回Outbound对象
	Outbound() *Outbound
}

////

type _EventSessionImpl struct {
	timestamp  time.Time
	attrs      map[string]interface{}
	topic      string
	uuid       string
	inbound    *Inbound
	outbound   *Outbound
	outputChan chan<- ObjectPacket
}

func (s *_EventSessionImpl) Attrs() *cfg.Config {
	return cfg.Wrap(s.attrs)
}

func (s *_EventSessionImpl) GetAttr(key string) (interface{}, bool) {
	v, ok := s.attrs[key]
	return v, ok
}

func (s *_EventSessionImpl) AddAttr(name string, value interface{}) {
	s.attrs[name] = value
}

func (s *_EventSessionImpl) AddAttrs(attributes map[string]interface{}) {
	for k, v := range attributes {
		s.AddAttr(k, v)
	}
}

func (s *_EventSessionImpl) Timestamp() time.Time {
	return s.timestamp
}

func (s *_EventSessionImpl) Topic() string {
	return s.topic
}

func (s *_EventSessionImpl) Uuid() string {
	return s.uuid
}

func (s *_EventSessionImpl) Inbound() *Inbound {
	return s.inbound
}

func (s *_EventSessionImpl) Outbound() *Outbound {
	return s.outbound
}

func (s *_EventSessionImpl) Since() time.Duration {
	return time.Since(s.Timestamp())
}
