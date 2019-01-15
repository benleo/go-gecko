package gecko

import "context"

type SessionHandler func(session Session)

//

// Events是一个三级Channel的事件处理器
type Dispatcher struct {
	lv0Chan    chan Session
	lv1Chan    chan Session
	lv0Handler SessionHandler
	lv1Handler SessionHandler
	lv2Handler SessionHandler
}

func NewDispatcher(capacity int) *Dispatcher {
	return &Dispatcher{
		lv0Chan: make(chan Session, capacity),
		lv1Chan: make(chan Session, capacity),
	}
}

func (es *Dispatcher) SetLv0Handler(handler SessionHandler) {
	es.lv0Handler = handler
}

func (es *Dispatcher) SetLv1Handler(handler SessionHandler) {
	es.lv1Handler = handler
}

func (es *Dispatcher) Lv0() chan<- Session {
	return es.lv0Chan
}

func (es *Dispatcher) Lv1() chan<- Session {
	return es.lv1Chan
}

func (es *Dispatcher) Serve(shutdown context.Context) {
	var c0 <-chan Session = es.lv0Chan
	var c1 <-chan Session = es.lv1Chan
	for {
		select {
		case <-shutdown.Done():
			return

		case v := <-c0:
			go es.lv0Handler(v)

		case v := <-c1:
			go es.lv1Handler(v)

		}
	}
}
