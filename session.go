package gecko

import (
	"github.com/parkingwang/go-conf"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

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
	Inbound() MessagePacket

	// 返回Outbound对象
	Outbound() MessagePacket
}

////

type _EventSessionImpl struct {
	timestamp time.Time
	attrs     map[string]interface{}
	topic     string
	uuid      string
	inbound   MessagePacket
	outbound  MessagePacket
	completed chan MessagePacket
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

func (s *_EventSessionImpl) Inbound() MessagePacket {
	return s.inbound
}

func (s *_EventSessionImpl) Outbound() MessagePacket {
	return s.outbound
}

func (s *_EventSessionImpl) Since() time.Duration {
	return time.Since(s.Timestamp())
}
