package gecko

import (
	"sync"
	"testing"
)

func BenchmarkNewEvents(b *testing.B) {

	breakChan := make(chan struct{}, 1)
	events := NewEvents(4, breakChan)

	wg := new(sync.WaitGroup)
	// 600 ns
	events.SetLv0Handler(func(in Session) {
		events.Lv1() <- in
	})
	// 1800 ns
	events.SetLv1Handler(func(in Session) {
		events.Lv2() <- in
	})
	// 3400 ns
	events.SetLv2Handler(func(in Session) {
		wg.Done()
	})

	go events.Serve()

	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		events.Lv0() <- nil
	}
	wg.Wait()
	breakChan <- struct{}{}
}
