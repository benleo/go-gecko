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

// 用来发起请求，并输出结果
type Invoker func(event *TriggerEvent, callback OnTriggerCompleted)

// Trigger是一个负责接收前端事件，并调用 {@link ContextInvoker} 方法函数来向系统内部发起触发事件通知；
// 内部系统处理完成后，将回调完成函数，返回输出
type Trigger interface {
	Initialize

	// 启动
	OnStart(scoped GeckoScoped, invoker Invoker)

	// 停止
	OnStop(scoped GeckoScoped, invoker Invoker)
}
