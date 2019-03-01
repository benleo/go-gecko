package gecko

import "context"

type SessionHandler func(session Session)

// Events是一个二级Channel的事件处理器
type Dispatcher struct {
	startChan    chan Session
	endChan      chan Session
	startHandler SessionHandler
	endHandler   SessionHandler
}

func NewDispatcher(capacity int) *Dispatcher {
	return &Dispatcher{
		startChan: make(chan Session, capacity),
		endChan:   make(chan Session, capacity),
	}
}

func (d *Dispatcher) SetStartHandler(handler SessionHandler) {
	d.startHandler = handler
}

func (d *Dispatcher) SetEndHandler(handler SessionHandler) {
	d.endHandler = handler
}

func (d *Dispatcher) StartC() chan<- Session {
	return d.startChan
}

func (d *Dispatcher) EndC() chan<- Session {
	return d.endChan
}

func (d *Dispatcher) Serve(shutdown context.Context) {
	var start <-chan Session = d.startChan
	var end <-chan Session = d.endChan
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
