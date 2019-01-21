package gecko

import (
	"github.com/parkingwang/go-conf"
	"sync"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// Session 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type Session interface {
	// 返回属性列表
	Attributes() *cfg.Config

	// 添加属性
	AddAttribute(key string, value interface{})

	// 添加多个属性。相同Key的属性将被覆盖。
	AddAttributes(attributes map[string]interface{})

	// Context创建的时间戳
	Timestamp() time.Time

	// 从创建起始，到当前时间的用时
	Since() time.Duration

	// 当前事件的Topic
	Topic() string

	// 返回Inbound对象
	Inbound() *Inbound

	// 返回Outbound对象
	Outbound() *Outbound
}

////

type _GeckoSession struct {
	timestamp          time.Time
	attributes         map[string]interface{}
	attrLock           *sync.RWMutex
	topic              string
	inbound            *Inbound
	outbound           *Outbound
	onSessionCompleted func(PacketMap)
}

func (si *_GeckoSession) Attributes() *cfg.Config {
	si.attrLock.RLock()
	defer si.attrLock.RUnlock()
	return cfg.WrapConfig(si.attributes)
}

func (si *_GeckoSession) AddAttribute(name string, value interface{}) {
	si.attrLock.Lock()
	defer si.attrLock.Unlock()
	si.attributes[name] = value
}

func (si *_GeckoSession) AddAttributes(attributes map[string]interface{}) {
	si.attrLock.Lock()
	defer si.attrLock.Unlock()
	for k, v := range attributes {
		si.AddAttribute(k, v)
	}
}

func (si *_GeckoSession) Timestamp() time.Time {
	return si.timestamp
}

func (si *_GeckoSession) Topic() string {
	return si.topic
}

func (si *_GeckoSession) Inbound() *Inbound {
	return si.inbound
}

func (si *_GeckoSession) Outbound() *Outbound {
	return si.outbound
}

func (si *_GeckoSession) Since() time.Duration {
	return time.Since(si.Timestamp())
}
