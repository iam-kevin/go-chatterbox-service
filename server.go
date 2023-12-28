package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	data "iam-kevin/chatterbox/data"
	"iam-kevin/chatterbox/handlers"
	service "iam-kevin/chatterbox/service"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
}

const CHATTERBOX_DB string = "cbd"

func main() {
	r := chi.NewRouter()

	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}

	r.Get("/user", handlers.HandlerGetUser)
	r.Post("/create-user", handlers.HandleCreateUser)

	r.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "Couldn't upgarde the server with WS connection")
			return
		}

		defer conn.Close()
		db := r.Context().Value("_DB").(*sqlx.DB)
		inmemory := r.Context().Value("_INMEMORY").(*data.ChatterboxDB)

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

			var body data.ChatMessage
			d, _ := io.ReadAll(r)
			json.Unmarshal(d, &body)

			w, _ := conn.NextWriter(websocket.BinaryMessage)
			service.ProcessChat(db, inmemory, body, &w)
		}
	})

	server := http.Server{
		Handler:           r,
		Addr:              fmt.Sprintf(":%v", os.Getenv("APP_PORT")),
		ReadHeaderTimeout: 3 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			ctx := context.Background()
			ctx = context.WithValue(ctx, "_DB", db)
			return context.WithValue(ctx, "_INMEMORY", data.CreateInMemoryInstance())
		},
	}

	log.Fatal(server.ListenAndServe())
}
