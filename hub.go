// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import "log"

// Body 是消息正文，包含房间 id
type Body struct {
	roomID  string
	message []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *Body

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	rooms map[string][]*Client
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan *Body),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string][]*Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			sendMessage(h, &Body{client.roomID, []byte("Someone came in")})
			h.clients[client] = true
			h.rooms[client.roomID] = append(h.rooms[client.roomID], client)
			log.Println("Someone came in. now client count is", len(h.clients), " and room count is", len(h.rooms))
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				removeClient4Room(h, client)
				close(client.send)
				log.Println("Someone is gone.")
				sendMessage(h, &Body{client.roomID, []byte("Someone is gone")})
			}
		case body := <-h.broadcast:
			sendMessage(h, body)
		}
	}
}

func sendMessage(h *Hub, body *Body) {
	for _, client := range h.rooms[body.roomID] {
		select {
		case client.send <- body.message:
		default:
			close(client.send)
			delete(h.clients, client)
			removeClient4Room(h, client)
		}
	}
}

func removeClient4Room(h *Hub, c *Client) {
	index := -1
	clients := h.rooms[c.roomID]
	for i, client := range clients {
		if c == client {
			index = i
			break
		}
	}
	clients = append(clients[:index], clients[index+1:]...)
	if 0 == len(clients) {
		delete(h.rooms, c.roomID)
	} else {
		h.rooms[c.roomID] = clients
	}
}
