package gecko

import (
	"sync"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

////

// EventSession 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type EventSession interface {
	// 返回属性列表
	Attrs() map[string]interface{}

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
	Inbound() *MessagePacket

	// 返回Outbound对象
	Outbound() *MessagePacket
}

////

type _EventSessionImpl struct {
	timestamp time.Time
	attrs     *sync.Map
	topic     string
	uuid      string
	inbound   *MessagePacket
	outbound  *MessagePacket
	completed chan *MessagePacket
}

func (s *_EventSessionImpl) Attrs() map[string]interface{} {
	m := make(map[string]interface{}, 0)
	s.attrs.Range(func(key, value interface{}) bool {
		m[key.(string)] = value
		return true
	})
	return m
}

func (s *_EventSessionImpl) GetAttr(key string) (interface{}, bool) {
	return s.attrs.Load(key)
}

func (s *_EventSessionImpl) AddAttr(key string, value interface{}) {
	s.attrs.Store(key, value)
}

func (s *_EventSessionImpl) AddAttrs(attributes map[string]interface{}) {
	for k, v := range attributes {
		s.attrs.Store(k, v)
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

func (s *_EventSessionImpl) Inbound() *MessagePacket {
	return s.inbound
}

func (s *_EventSessionImpl) Outbound() *MessagePacket {
	return s.outbound
}

func (s *_EventSessionImpl) Since() time.Duration {
	return time.Since(s.Timestamp())
}
