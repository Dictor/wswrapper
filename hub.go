// Original Code by : Copyright C) 2013 The Gorilla WebSocket Authors. All rights reserved.
// Modified by Dictor(kimdictor@gmail.com)

package wswrapper // import "github.com/dictor/wswrapper"

import (
	"github.com/gorilla/websocket"
	"net/http"
)

//Explicit type for directing websocket event kind.
type WebsocketEventKind int

const (
	EVENT_RECIEVE WebsocketEventKind = iota
	EVENT_REGISTER
	EVENT_UNREGISTER
	EVENT_ERROR
)

//Websocket event object which passing to event handler
type WebsocketEvent struct {
	Kind   WebsocketEventKind //event kind
	Client *WebsocketClient   //client address who causes this event
	Msg    *[]byte            //message data when event kind is EVENT_RECIEVE
	Err    error              //error data when event kind is EVENT_ERROR
}

//Websocket server object. No public field.
type WebsocketHub struct {
	clients      map[*WebsocketClient]int
	clientsCount int
	broadcast    chan []byte
	event        chan *WebsocketEvent
	register     chan *WebsocketClient
	unregister   chan *WebsocketClient
	upgrader     websocket.Upgrader
}

//Create new websocket server object
func NewHub() *WebsocketHub {
	return &WebsocketHub{
		broadcast:    make(chan []byte, 1),
		event:        make(chan *WebsocketEvent),
		register:     make(chan *WebsocketClient),
		unregister:   make(chan *WebsocketClient),
		clients:      make(map[*WebsocketClient]int),
		clientsCount: 0,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
	}
}

//Listen client registering and data receiving. Passed callback will be called when any events cause
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
				if !h.Send(client, message) {
					event_callback(&WebsocketEvent{EVENT_UNREGISTER, client, nil, nil})
				}
			}
		case evt := <-h.event:
			event_callback(evt)
		}
	}
}

//Disconnect client and drop client record.
func (h *WebsocketHub) closeClient(cli *WebsocketClient) {
	close(cli.send)
	delete(h.clients, cli)
}

//Send message to passed client address. When passed client is inactive state then it will be unregisterd.
func (h *WebsocketHub) Send(cli *WebsocketClient, msg []byte) bool {
	select {
	case cli.send <- msg:
		return true
	default: // when client.send closed, close and delete client
		h.closeClient(cli)
		return false
	}
}

//Repeat Send() for all registerd clients.
func (h *WebsocketHub) SendAll(msg []byte) {
	// Sending in separated goroutine for preventing deadlock
	go func() {
		h.broadcast <- msg
	}()
}

//Register client from http request.
func (h *WebsocketHub) AddClient(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
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

//Set origin header when will be sended websocket registering (upgrading)
func (h *WebsocketHub) AddUpgraderOrigin(origin []string) {
	h.upgrader.CheckOrigin = func(r *http.Request) bool {
		for _, val := range origin {
			if val == r.Host {
				return true
			}
		}
		return false
	}
}

//Get all registerd clients
func (h *WebsocketHub) Clients() map[*WebsocketClient]int {
	return h.clients
}
