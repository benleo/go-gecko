package gecko

import (
	"time"
	"sync"
	"parkingwang.com/go-conf"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Session 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type Session interface {
	// 返回属性列表
	Attributes() *conf.ImmutableMap

	// 添加属性
	AddAttribute(key string, value interface{})

	// 添加多个属性。相同Key的属性将被覆盖。
	AddAttributes(attributes map[string]interface{})

	// Context创建的时间戳
	Timestamp() time.Time

	// 从创建起始，到当前时间的用时
	Escaped() time.Duration

	// 当前事件的Topic
	Topic() string

	// 当前事件的ContextId。每个事件具有唯一ID。
	ContextId() int64

	// 返回Inbound对象
	Inbound() *Inbound

	// 返回Outbound对象
	Outbound() *Outbound

	// 创建数据帧对象
	NewPacketFrame(frame []byte) *PacketFrame
}

////

type sessionImpl struct {
	timestamp       time.Time
	attributes      map[string]interface{}
	attrLock        *sync.RWMutex
	topic           string
	contextId       int64
	inbound         *Inbound
	outbound        *Outbound
	onCompletedFunc OnTriggerCompleted
}

func (si *sessionImpl) Attributes() *conf.ImmutableMap {
	si.attrLock.RLock()
	defer si.attrLock.RUnlock()
	return conf.WrapImmutableMap(si.attributes)
}

func (si *sessionImpl) AddAttribute(name string, value interface{}) {
	si.attrLock.Lock()
	defer si.attrLock.Unlock()
	si.attributes[name] = value
}

func (si *sessionImpl) AddAttributes(attributes map[string]interface{}) {
	si.attrLock.Lock()
	defer si.attrLock.Unlock()
	for k, v := range attributes {
		si.AddAttribute(k, v)
	}
}

func (si *sessionImpl) Timestamp() time.Time {
	return si.timestamp
}

func (si *sessionImpl) Topic() string {
	return si.topic
}

func (si *sessionImpl) ContextId() int64 {
	return si.contextId
}

func (si *sessionImpl) Inbound() *Inbound {
	return si.inbound
}

func (si *sessionImpl) Outbound() *Outbound {
	return si.outbound
}

func (si *sessionImpl) Escaped() time.Duration {
	return time.Now().Sub(si.Timestamp())
}

func (si *sessionImpl) NewPacketFrame(frame []byte) *PacketFrame {
	return &PacketFrame{
		id:     si.ContextId(),
		header: si.Attributes(),
		frame:  frame,
	}
}
