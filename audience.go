// learnings from
// https://github.com/gorilla/websocket/blob/main/examples/chat

package main

import (
	"context"

	"github.com/gorilla/websocket"
)

type Node struct {
	audience *Audience
	conn     *websocket.Conn
}

type Audience struct {
	clients   map[*Node]bool
	register  chan *Node
	remove    chan *Node
	broadcast chan []byte
}

func SetupAudience() *Audience {
	return &Audience{
		clients:   make(map[*Node]bool),
		register:  make(chan *Node),
		broadcast: make(chan []byte),
	}
}

func (au *Audience) Broadcast(data []byte) {
	for client := range au.clients {
		client.conn.WriteMessage(websocket.BinaryMessage, data)
	}
}

func (au *Audience) Listen(ctx context.Context) {
	for {
		select {
		case node := <-au.register:
			{
				println("registering this user")

				// send previous *recorded* chats to newly connected users
				// db := ctx.Value("_DB").(*sqlx.DB)

				au.clients[node] = true
			}
		case node := <-au.remove:
			{
				println("removing this user")
				delete(au.clients, node) // remove item from node
			}
		}
	}
}
