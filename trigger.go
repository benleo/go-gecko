package gecko

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 触发器事件
type TriggerEvent struct {
	topic string
	data  map[string]interface{}
}

func NewTriggerEvent(topic string, data map[string]interface{}) *TriggerEvent {
	return &TriggerEvent{
		topic: topic,
		data:  data,
	}
}

////

// 处理结果回调接口
type OnTriggerCompleted func(data map[string]interface{})

////

// Invoker
// 用来发起请求，并输出结果
type Invoker func(event *TriggerEvent, callback OnTriggerCompleted)

// 绑定函数来实现同步执行
func (in Invoker) Execute(event *TriggerEvent) map[string]interface{} {
	return <-in.Execute0(event)
}

// 绑定函数来实现同步执行
func (in Invoker) Execute0(event *TriggerEvent) <-chan map[string]interface{} {
	resp := make(chan map[string]interface{}, 1)
	in(event, func(data map[string]interface{}) {
		resp <- data
	})
	return resp
}

////

// Trigger是一个负责接收前端事件，并调用 {@link ContextInvoker} 方法函数来向系统内部发起触发事件通知；
// 内部系统处理完成后，将回调完成函数，返回输出
type Trigger interface {
	Initialize

	// Trigger需要设置Topic
	setTopic(topic string)
	GetTopic() string

	// 启动
	OnStart(ctx Context, invoker Invoker)

	// 停止
	OnStop(ctx Context, invoker Invoker)
}

//

type AbcTrigger struct {
	Trigger
	topic string
}

func (at *AbcTrigger) setTopic(topic string) {
	at.topic = topic
}

func (at *AbcTrigger) GetTopic() string {
	return at.topic
}
