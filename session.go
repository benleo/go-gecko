package gecko

import (
	"github.com/parkingwang/go-conf"
	"sync"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// EventSession 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type EventSession interface {
	// 返回属性列表
	Attributes() *cfg.Config

	// 添加属性
	AddAttribute(key string, value interface{})

	// 添加多个属性。相同Key的属性将被覆盖。
	AddAttributes(attributes map[string]interface{})

	// 创建的时间戳
	Timestamp() time.Time

	// 从创建起始，到当前时间的用时
	Since() time.Duration

	// 当前事件的Topic
	Topic() string

	// 返回输入端消息对象
	Inbound() *Message

	// 返回Outbound对象
	Outbound() *Message
}

////

type _EventSessionImpl struct {
	timestamp  time.Time
	attributes map[string]interface{}
	attrLock   *sync.RWMutex
	topic      string
	inbound    *Message
	outbound   *Message
	outputChan chan <- JSONPacket
}

func (si *_EventSessionImpl) Attributes() *cfg.Config {
	si.attrLock.RLock()
	defer si.attrLock.RUnlock()
	return cfg.Wrap(si.attributes)
}

func (si *_EventSessionImpl) AddAttribute(name string, value interface{}) {
	si.attrLock.Lock()
	defer si.attrLock.Unlock()
	si.attributes[name] = value
}

func (si *_EventSessionImpl) AddAttributes(attributes map[string]interface{}) {
	si.attrLock.Lock()
	defer si.attrLock.Unlock()
	for k, v := range attributes {
		si.AddAttribute(k, v)
	}
}

func (si *_EventSessionImpl) Timestamp() time.Time {
	return si.timestamp
}

func (si *_EventSessionImpl) Topic() string {
	return si.topic
}

func (si *_EventSessionImpl) Inbound() *Message {
	return si.inbound
}

func (si *_EventSessionImpl) Outbound() *Message {
	return si.outbound
}

func (si *_EventSessionImpl) Since() time.Duration {
	return time.Since(si.Timestamp())
}
