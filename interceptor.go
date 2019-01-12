package gecko

import "errors"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

var ErrInterceptorDropped = errors.New("INTERCEPTOR_DROPPED")

// 事件拦截器。在Gecko系统中，通过Trigger触发事件后，由 Interceptor 处理拦截。
// 负责对触发器发起的事件进行拦截处理，不符合规则的事件将被中断，丢弃。
type Interceptor interface {
	Initialize
	NeedTopicFilter
	// Interceptor可设置优先级
	GetPriority() int
	setPriority(p int)
	// 拦截处理过程。抛出 {@link DropException} 来中断拦截。
	Handle(session Session, ctx Context) error
}

// Interceptor抽象实现
type AbcInterceptor struct {
	Interceptor
	priority int
	topics   []*TopicExpr
}

func (ai *AbcInterceptor) GetPriority() int {
	return ai.priority
}

func (ai *AbcInterceptor) setPriority(priority int) {
	ai.priority = priority
}

func (ai *AbcInterceptor) setTopics(topics []string) {
	for _, t := range topics {
		ai.topics = append(ai.topics, newTopicExpr(t))
	}
}

func (ai *AbcInterceptor) GetTopicExpr() []*TopicExpr {
	return ai.topics
}

func (ai *AbcInterceptor) Drop() error {
	return InterceptedDropError()
}

func NewAbcInterceptor() *AbcInterceptor {
	return &AbcInterceptor{
		topics: make([]*TopicExpr, 0),
	}
}

// 拦截器抛弃事件操作
func InterceptedDropError() error {
	return ErrInterceptorDropped
}

////

// 拦截器排序
type InterceptorSlice []Interceptor

func (is InterceptorSlice) Len() int { return len(is) }

func (is InterceptorSlice) Less(i, j int) bool {
	return is[i].GetPriority() > is[j].GetPriority()
}

func (is InterceptorSlice) Swap(i, j int) { is[i], is[j] = is[j], is[i] }
