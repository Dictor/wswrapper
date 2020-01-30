// Original Code by : Copyright C) 2013 The Gorilla WebSocket Authors. All rights reserved.
// Modified by Dictor(kimdictor@gmail.com)

package wswrapper // import "github.com/dictor/wswrapper"

import (
	"net/http"
)

type WebsocketEventKind int

const (
	EVENT_RECIEVE WebsocketEventKind = iota
	EVENT_REGISTER
	EVENT_UNREGISTER
	EVENT_BROADCAST
	EVENT_ERROR
)

type WebsocketEvent struct {
	Kind   WebsocketEventKind
	Client *WebsocketClient
	Msg    *[]byte
	Err    error
}

type WebsocketHub struct {
	clients      map[*WebsocketClient]int
	clientsCount int
	broadcast    chan []byte
	event        chan *WebsocketEvent
	register     chan *WebsocketClient
	unregister   chan *WebsocketClient
}

func NewWebsocketHub() *WebsocketHub {
	return &WebsocketHub{
		broadcast:    make(chan []byte),
		event:        make(chan *WebsocketEvent),
		register:     make(chan *WebsocketClient),
		unregister:   make(chan *WebsocketClient),
		clients:      make(map[*WebsocketClient]int),
		clientsCount: 0,
	}
}

func (h *WebsocketHub) Run(event_callback func(*WebsocketEvent)) {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = h.clientsCount
			h.clientsCount++
			event_callback(&WebsocketEvent{EVENT_REGISTER, client, nil, nil})
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				event_callback(&WebsocketEvent{EVENT_UNREGISTER, client, nil, nil})
				h.closeClient(client)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				if h.Send(client, &message) {
					event_callback(&WebsocketEvent{EVENT_BROADCAST, client, nil, nil})
				} else {
					event_callback(&WebsocketEvent{EVENT_UNREGISTER, client, nil, nil})
				}
			}
		case evt := <-h.event:
			event_callback(evt)
		}
	}
}

func (h *WebsocketHub) closeClient(cli *WebsocketClient) {
	close(cli.send)
	delete(h.clients, cli)
}

func (h *WebsocketHub) SendSafe(cli *WebsocketClient, msg *[]byte) bool {
	select {
	case cli.send <- *msg:
		return true
	default: // when client.send closed, close and delete client
		h.closeClient(cli)
		return false
	}
}

func (h *WebsocketHub) AddClient(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.event <- &WebsocketEvent{EVENT_ERROR, nil, nil, err}
		return
	}
	client := &WebsocketClient{hub: h, conn: conn, send: make(chan []byte, 256)}
	h.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
