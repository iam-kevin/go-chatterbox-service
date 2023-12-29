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
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/r3labs/sse/v2"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize: 1024,
}

const CHATTERBOX_DB string = "cbd"

func main() {
	r := chi.NewRouter()

	sseServer := sse.New()
	sseServer.CreateStream("rooms")
	sseServer.AutoReplay = false
	sseServer.AutoStream = true
	sseServer.Headers = map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	// enable cors
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*", "http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	db, err := sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalln(err)
	}

	r.Get("/user", handlers.HandlerGetUser)
	r.Post("/set-user", handlers.HandleSetUser)
	// r.Get("/room")

	// SSE
	r.HandleFunc("/events", sseServer.ServeHTTP)

	go func(db *sqlx.DB, sseServer *sse.Server) {
		for {
			if sseServer.StreamExists("rooms") {
				var rooms []service.Chatroom
				err := db.Select(&rooms, `select id, user_id, name, created_at from "room"`)
				if err != nil {
					log.Print(err)
					// w.WriteHeader(501)
					// fmt.Fprintf(w, `{ "error": "DB_FETCH", "message": "unable to fetch the results" }`)
					return
				}

				data, err := json.Marshal(rooms)
				if err != nil {
					log.Fatal(err)
				}

				// log.Print(rooms)
				sseServer.Publish("rooms", &sse.Event{
					Data: data,
				})
			}

			time.Sleep(time.Second * 5)
		}
	}(db, sseServer)

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
