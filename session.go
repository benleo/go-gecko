package gecko

import (
	"github.com/yoojia/go-value"
	"sync"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Attributes interface {
	// 返回属性列表. 只读属性
	Attrs() map[string]interface{}

	// 添加属性
	AddAttr(key string, value interface{})

	// 获取属性
	GetAttr(key string) (interface{}, bool)

	// 获取属性值, 如果不存在,返回 nil
	GetAttrOrNil(key string) interface{}

	// 获取String类型属性
	GetAttrString(key string) (string, bool)

	// 获取Int64类型的属性
	GetAttrInt64(key string) (int64, bool)

	// 判断属性是否存在
	HasAttr(key string) bool
}

type attributesMap struct {
	Attributes
	values *sync.Map
}

func (a *attributesMap) Attrs() map[string]interface{} {
	rom := make(map[string]interface{})
	a.values.Range(func(key, value interface{}) bool {
		rom[key.(string)] = value
		return true
	})
	return rom
}

func (a *attributesMap) AddAttr(key string, value interface{}) {
	a.values.Store(key, value)
}

func (a *attributesMap) GetAttr(key string) (interface{}, bool) {
	return a.values.Load(key)
}

func (a *attributesMap) GetAttrOrNil(key string) interface{} {
	v, ok := a.values.Load(key)
	if ok {
		return v
	} else {
		return nil
	}
}

func (a *attributesMap) GetAttrString(key string) (string, bool) {
	val, ok := a.values.Load(key)
	if ok {
		return value.Of(val).String(), true
	} else {
		return "", false
	}
}

func (a *attributesMap) GetAttrInt64(key string) (int64, bool) {
	val, ok := a.values.Load(key)
	if ok {
		return value.Of(val).ToInt64()
	} else {
		return 0, false
	}
}

func (a *attributesMap) HasAttr(key string) bool {
	_, ok := a.values.Load(key)
	return ok
}

func newMapAttributesWith(attrs map[string]interface{}) *attributesMap {
	am := &attributesMap{values: new(sync.Map)}
	for k, v := range attrs {
		am.values.Store(k, v)
	}
	return am
}

////

// EventSession 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type EventSession interface {
	Attributes

	// 创建的时间戳
	Timestamp() time.Time

	// 从创建起始，到当前时间的用时
	Since() time.Duration

	// 当前事件的Topic
	Topic() string

	// 当前事件的设备地址
	Uuid() string

	// 返回输入端消息对象
	Inbound() *MessagePacket

	// 返回Outbound对象
	Outbound() *MessagePacket
}

////

type _EventSessionImpl struct {
	timestamp time.Time
	*attributesMap
	topic     string
	uuid      string
	inbound   *MessagePacket
	outbound  *MessagePacket
	completed chan *MessagePacket
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

func (s *_EventSessionImpl) Inbound() *MessagePacket {
	return s.inbound
}

func (s *_EventSessionImpl) Outbound() *MessagePacket {
	return s.outbound
}

func (s *_EventSessionImpl) Since() time.Duration {
	return time.Since(s.Timestamp())
}
