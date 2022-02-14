// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"testtask/service"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WebsocketController is a middleman between the websocket connection and the pool.
type WebsocketController struct {
	request *http.Request

	pool *service.ConnectionPool

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte
}

// listenEvents pumps messages from the websocket connection to the pool.
//
// The application runs listenEvents in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *WebsocketController) listenEvents() {
	defer func() {
		c.pool.Unregister <- c.Send
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
	}
}

// buildResponse pumps messages from the pool to the websocket connection.
//
// A goroutine running buildResponse is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *WebsocketController) buildResponse() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The pool closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			w.Write(message)
			n := len(c.Send)

			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func HandleWs(pool *service.ConnectionPool, m *service.PriceManager, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	client := make(chan []byte)

	if err != nil {
		log.Println(err)
		return
	}

	controller := &WebsocketController{
		request: r,
		pool:    pool,
		conn:    conn,
		Send:    client,
	}
	controller.pool.Register <- controller.Send

	go func() {
		err := m.Broadcast(client, r.URL.Query()["fsyms"], r.URL.Query()["tsyms"])
		if err != nil {
			controller.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, "resource not available"))
			controller.conn.Close()
			if !service.IsClosed(client) {
				close(client)
			}
		}
	}()

	go controller.listenEvents()
	go controller.buildResponse()
}
