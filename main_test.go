package main

import (
	"io"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestDial(t *testing.T) {
	dialer := &websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://localhost:8546", nil)
	if err != nil {
		panic(err)
	}

	msg := []byte(`{"jsonrpc": "2.0","method": "eth_blockNumber","params": [],"id": 1}`)
	ticker := time.NewTicker(time.Second * 5)

	for i := 0; i < 3; i++ {
		err = conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			panic(err)
		}

		_, reader, err := conn.NextReader()
		if err != nil {
			panic(err)
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}

		t.Log(string(data))
		<-ticker.C
	}
}
