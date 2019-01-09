package gecko

import "time"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// GeckoContext 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type GeckoContext interface {
	// 返回属性列表
	Attributes() map[string]interface{}

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

type abcGeckoContext struct {
	GeckoContext
	timestamp       time.Time
	attributes      map[string]interface{}
	topic           string
	contextId       int64
	inbound         *Inbound
	outbound        *Outbound
	onCompletedFunc OnTriggerCompleted
}

func (gc *abcGeckoContext) Attributes() map[string]interface{} {
	return gc.attributes
}

func (gc *abcGeckoContext) AddAttribute(name string, value interface{}) {
	gc.attributes[name] = value
}

func (gc *abcGeckoContext) AddAttributes(attributes map[string]interface{}) {
	for k, v := range attributes {
		gc.AddAttribute(k, v)
	}
}

func (gc *abcGeckoContext) Timestamp() time.Time {
	return gc.timestamp
}

func (gc *abcGeckoContext) Topic() string {
	return gc.topic
}

func (gc *abcGeckoContext) ContextId() int64 {
	return gc.contextId
}

func (gc *abcGeckoContext) Inbound() *Inbound {
	return gc.inbound
}

func (gc *abcGeckoContext) Outbound() *Outbound {
	return gc.outbound
}

func (gc *abcGeckoContext) Escaped() time.Duration {
	return time.Now().Sub(gc.Timestamp())
}

func (gc *abcGeckoContext) NewPacketFrame(frame []byte) *PacketFrame {
	return &PacketFrame{
		id:     gc.ContextId(),
		header: gc.Attributes(),
		frame:  frame,
	}
}
