package gecko

type SessionHandler func(session Session)

//

// Events是一个三级Channel的事件处理器
type Events struct {
	shutdown   <-chan struct{}
	lv0Chan    chan Session
	lv1Chan    chan Session
	lv2Chan    chan Session
	lv0Handler SessionHandler
	lv1Handler SessionHandler
	lv2Handler SessionHandler
}

func NewEvents(capacity int, shutdown <-chan struct{}) *Events {
	return &Events{
		lv0Chan:  make(chan Session, capacity),
		lv1Chan:  make(chan Session, capacity),
		lv2Chan:  make(chan Session, capacity),
		shutdown: shutdown,
	}
}

func (es *Events) SetLv0Handler(handler SessionHandler) {
	es.lv0Handler = handler
}

func (es *Events) SetLv1Handler(handler SessionHandler) {
	es.lv1Handler = handler
}

func (es *Events) SetLv2Handler(handler SessionHandler) {
	es.lv2Handler = handler
}

func (es *Events) Lv0() chan<- Session {
	return es.lv0Chan
}

func (es *Events) Lv1() chan<- Session {
	return es.lv1Chan
}

func (es *Events) Lv2() chan<- Session {
	return es.lv2Chan
}

func (es *Events) Serve() {
	var c0 <-chan Session = es.lv0Chan
	var c1 <-chan Session = es.lv1Chan
	var c2 <-chan Session = es.lv2Chan
	for {
		select {
		case <-es.shutdown:
			return

		case v := <-c0:
			go es.lv0Handler(v)

		case v := <-c1:
			go es.lv1Handler(v)

		case v := <-c2:
			go es.lv2Handler(v)
		}
	}
}
