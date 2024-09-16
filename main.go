package main

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// struct implementing http.Handler
type WS struct {
	upgrader *websocket.Upgrader
	dst      string
}

func (ws *WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// websocket upgrade
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	// dialer for dialing to the destination
	dialer, _, err := websocket.DefaultDialer.Dial(ws.dst, nil)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}
	defer dialer.Close()

	for {
		// this data is from the user
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			conn.WriteMessage(messageType, []byte(err.Error()))
			return
		}

		// writes the user's message to the destination
		err = dialer.WriteMessage(messageType, data)
		if err != nil {
			conn.WriteMessage(messageType, []byte(err.Error()))
			return
		}

		// receives from the destination
		_, dstReader, err := dialer.NextReader()
		if err != nil {
			conn.WriteMessage(messageType, []byte(err.Error()))
			return
		}

		// read the destination's message
		dstData, err := io.ReadAll(dstReader)
		if err != nil {
			conn.WriteMessage(messageType, []byte(err.Error()))
			return
		}

		// writes the destination's message to the user
		err = conn.WriteMessage(messageType, dstData)
		if err != nil {
			conn.WriteMessage(messageType, []byte(err.Error()))
			return
		}
	}
}

func main() {
	// logger setting
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// url for proxying
	dst := "http://127.0.0.1:8545"
	u, err := url.Parse(dst)
	if err != nil {
		logger.Fatal().Err(err).Send()
	}

	// channel to receive a routing error
	ch := make(chan error, 1)

	// http route (port 3000)
	go func() {
		rp := &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				logger.Info().Any("request", map[string]any{
					"path":   r.URL.Path,
					"method": r.Method,
					"ip":     r.RemoteAddr,
				}).Send()
				r.URL = u
			},
		}
		http.Handle("/http", rp)
		if err := http.ListenAndServe(":3000", nil); err != nil {
			ch <- err
		}
	}()

	// websocket route (port 3001)
	go func() {
		ws := &WS{
			upgrader: &websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
			},
			dst: dst,
		}
		http.Handle("/ws", ws)
		if err := http.ListenAndServe(":3001", nil); err != nil {
			ch <- err
		}
	}()

	// exit when error caught
	err = <-ch
	logger.Error().Err(err).Send()
}
