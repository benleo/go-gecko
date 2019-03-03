package network

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestSocketClient(t *testing.T) {
	client := new(SocketClient)
	client.Init(SocketConfig{
		Type:         "udp",
		Addr:         "127.0.0.1:3456",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		BufferSize:   32,
	})
	if err := client.Open(); nil != err {
		t.Fatal(err)
	}
	<-time.After(time.Second)
	for i := 0; i <= 10; i++ {
		<-time.After(time.Millisecond)
		fmt.Println(fmt.Sprintf("Client send hello: %d\n", i))
		if _, err := client.Send([]byte("hello\n")); nil != err {
			t.Error(err)
			t.Fail()
		}
	}
	if err := client.Close(); nil != err {
		t.Fatal(err)
	}
}

func TestSocketServer(t *testing.T) {
	server := NewSocketServer()
	server.Init(SocketConfig{
		Type:         "tcp",
		Addr:         "127.0.0.1:5555",
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
		BufferSize:   32,
	})

	time.AfterFunc(time.Second*10, func() {
		fmt.Println("Server shutdown")
		server.Shutdown()
	})

	if err := server.Serve(func(addr net.Addr, frame []byte) (resp []byte) {
		fmt.Println(fmt.Sprintf("Addr: %s, byte: %v", addr, frame))
		return []byte("SERVER_ECHO")
	}); nil != err {
		t.Fatal(err)
	}
}
