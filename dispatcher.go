package gecko

import "context"

type SessionHandler func(session Session)

// Events是一个三级Channel的事件处理器
type Dispatcher struct {
	chan00    chan Session
	chan11    chan Session
	handler00 SessionHandler
	handler11 SessionHandler
}

func NewDispatcher(capacity int) *Dispatcher {
	return &Dispatcher{
		chan00: make(chan Session, capacity),
		chan11: make(chan Session, capacity),
	}
}

func (es *Dispatcher) Set00Handler(handler SessionHandler) {
	es.handler00 = handler
}

func (es *Dispatcher) Set11Handler(handler SessionHandler) {
	es.handler11 = handler
}

func (es *Dispatcher) Channel00() chan<- Session {
	return es.chan00
}

func (es *Dispatcher) Channel11() chan<- Session {
	return es.chan11
}

func (es *Dispatcher) Serve(shutdown context.Context) {
	var c0 <-chan Session = es.chan00
	var c1 <-chan Session = es.chan11
	for {
		select {
		case <-shutdown.Done():
			return

		case v := <-c0:
			go es.handler00(v)

		case v := <-c1:
			go es.handler11(v)

		}
	}
}
