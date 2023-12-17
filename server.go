package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	db "iam-kevin/chatterbox/data"
	service "iam-kevin/chatterbox/service"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
}

const CHATTERBOX_DB string = "cbd"

func main() {
	r := chi.NewRouter()

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
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
		cdb := r.Context().Value(CHATTERBOX_DB).(db.ChatterboxDB)

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

			var body db.ChatMessage
			d, _ := io.ReadAll(r)
			json.Unmarshal(d, &body)

			w, _ := conn.NextWriter(websocket.BinaryMessage)

			mm := db.Member{Name: "Kevin", Username: "iam-kevin"}
			service.ProcessChat(&cdb, &mm, body, &w)
		}
	})

	server := http.Server{
		Handler:           r,
		Addr:              ":8080",
		ReadHeaderTimeout: 3 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			ctx := context.WithValue(context.Background(), CHATTERBOX_DB, db.CreateInMemoryInstance())
			return ctx
		},
	}

	log.Fatal(server.ListenAndServe())
}
