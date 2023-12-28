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
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
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

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}

	r.Get("/user", handlers.HandlerGetUser)
	r.Post("/set-user", handlers.HandleSetUser)
	// r.Get("/room")

	// SSE
	r.HandleFunc("/events/rooms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		db := r.Context().Value("_DB").(*sqlx.DB)
		for {
			// Send an event
			fmt.Fprintf(w, "data: The server time is %v\n\n", time.Now().Format(time.RFC1123))
			service.EvtCheckRooms(db, w)

			// Flush the data immediately instead of buffering it for later.
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			} else {
				log.Println("Warning: Streaming not supported!")
				break
			}

			time.Sleep(2 * time.Second) // Adjust the frequency of messages as needed

			// Check if the client is still connected
			if cn, ok := w.(http.CloseNotifier); ok {
				select {
				case <-cn.CloseNotify():
					log.Println("Client has closed the connection")
					return
				default:
					// continue
				}
			}
		}
	})

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
			service.OnReceiveMessage(db, inmemory, body, &w)
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
