package gecko

import "context"

type SessionHandler func(session EventSession)

// Events是一个二级Channel的事件处理器
type Dispatcher struct {
	startChan    chan EventSession
	endChan      chan EventSession
	startHandler SessionHandler
	endHandler   SessionHandler
}

func NewDispatcher(capacity int) *Dispatcher {
	return &Dispatcher{
		startChan: make(chan EventSession, capacity),
		endChan:   make(chan EventSession, capacity),
	}
}

func (d *Dispatcher) SetStartHandler(handler SessionHandler) {
	d.startHandler = handler
}

func (d *Dispatcher) SetEndHandler(handler SessionHandler) {
	d.endHandler = handler
}

func (d *Dispatcher) StartC() chan<- EventSession {
	return d.startChan
}

func (d *Dispatcher) EndC() chan<- EventSession {
	return d.endChan
}

func (d *Dispatcher) Serve(shutdown context.Context) {
	var start <-chan EventSession = d.startChan
	var end <-chan EventSession = d.endChan
	for {
		select {
		case <-shutdown.Done():
			return

		case v := <-start:
			go d.startHandler(v)

		case v := <-end:
			go d.endHandler(v)

		}
	}
}
