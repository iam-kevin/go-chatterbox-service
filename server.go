package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	chatterbox "iam-kevin/chatterbox/service"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Sanity checking...")
	})

	r.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
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

			chatterbox.DecodeText(body.Message)

			// show the received message
			fmt.Printf("received: %s\n", body.Message)
		}
	})

	r.HandleFunc("/room/{id}/chat", func(w http.ResponseWriter, r *http.Request) {
		// ..
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}

type ChatMessage struct {
	Message string `json:"message"`
}
