package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize: 1024,
	}
	mux = http.NewServeMux()
)

func main() {
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Sanity checking...")
	})

	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Couldn't upgarde the server with WS connection")
			return
		}

		defer conn.Close()

		//  reaf from the chat endpoint
		for {
			msgType, r, err := conn.NextReader()
			if err != nil {
				fmt.Fprintf(w, "There was a problem when reading the message")
				return
			}

			if msgType != websocket.TextMessage {
				panic("Couldn't read the message dude! Fail badly")
			}

			var body ChatMessage
			d, _ := io.ReadAll(r)
			json.Unmarshal(d, &body)

			// show the received message
			fmt.Printf("received: %s\n", body.Message)
		}
	})

	server := &http.Server{
		Addr: ":8080",
		BaseContext: func(l net.Listener) context.Context {
			return context.Background()
		},
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}

type ChatMessage struct {
	Message string `json:"message"`
}
