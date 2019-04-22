package gecko

import (
	"github.com/yoojia/go-value"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type Attributes interface {
	// 返回属性列表. 只读属性
	Map() map[string]interface{}

	// 添加属性
	Add(key string, value interface{})

	// 获取属性
	Get(key string) (interface{}, bool)

	// 获取属性值, 如果不存在,返回 nil
	GetOrNil(key string) interface{}

	// 获取String类型属性
	GetString(key string) (string, bool)

	// 获取Int64类型的属性
	GetInt64(key string) (int64, bool)

	// 判断属性是否存在
	HasAttr(key string) bool
}

type AttrMap struct {
	Attributes
	data map[string]interface{}
}

func (a *AttrMap) Map() map[string]interface{} {
	return a.data
}

func (a *AttrMap) Add(key string, value interface{}) {
	a.data[key] = value
}

func (a *AttrMap) Get(key string) (interface{}, bool) {
	v, ok := a.data[key]
	return v, ok
}

func (a *AttrMap) GetOrNil(key string) interface{} {
	if v, ok := a.data[key]; ok {
		return v
	} else {
		return nil
	}
}

func (a *AttrMap) GetString(key string) (string, bool) {
	if val, ok := a.data[key]; ok {
		return value.Of(val).String(), true
	} else {
		return "", false
	}
}

func (a *AttrMap) GetInt64(key string) (int64, bool) {
	if val, ok := a.data[key]; ok {
		return value.Of(val).ToInt64()
	} else {
		return 0, false
	}
}

func (a *AttrMap) HasAttr(key string) bool {
	_, ok := a.data[key]
	return ok
}

func newMapAttributesWith(attrs map[string]interface{}) *AttrMap {
	am := &AttrMap{data: make(map[string]interface{})}
	for k, v := range attrs {
		am.data[k] = v
	}
	return am
}

////

// session 是每次请求生成的上下文对象，服务于事件请求的整个生命周期。
type session interface {
	Attrs() Attributes

	// 创建的时间戳
	Timestamp() time.Time

	// 从创建起始，到当前时间的用时
	Since() time.Duration

	// 当前事件的Topic
	Topic() string

	// 当前事件的设备地址
	Uuid() string

	// 返回输入端消息对象
	GetInbound() *MessagePacket

	// 返回Outbound对象
	WriteOutbound(mp *MessagePacket)
}

////

type _Session struct {
	timestamp time.Time
	attrs     *AttrMap
	topic     string
	uuid      string
	inbound   *MessagePacket
	outbound  chan *MessagePacket
}

func (s *_Session) Attrs() Attributes {
	return s.attrs
}

func (s *_Session) Timestamp() time.Time {
	return s.timestamp
}

func (s *_Session) Topic() string {
	return s.topic
}

func (s *_Session) Uuid() string {
	return s.uuid
}

func (s *_Session) GetInbound() *MessagePacket {
	return s.inbound
}

func (s *_Session) WriteOutbound(mp *MessagePacket) {
	s.outbound <- mp
}

func (s *_Session) Since() time.Duration {
	return time.Since(s.Timestamp())
}
