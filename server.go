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
	"strings"
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

const CHATTERBOX_DB string = "cbd"

type wildcard struct {
	prefix string
	suffix string
}

func (w wildcard) match(s string) bool {
	return len(s) >= len(w.prefix+w.suffix) && strings.HasPrefix(s, w.prefix) && strings.HasSuffix(s, w.suffix)
}

func main() {
	r := chi.NewRouter()

	// setting the CORS configuration
	corsOptions := (cors.Options{
		AllowedOrigins:     []string{"*"}, // []string{"https://*", "http://*", "http://localhost:5173"}, // or some reason.. that work, but // []string{"*"} // doesn't
		AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials:   true,
		OptionsPassthrough: true,
		MaxAge:             300,
	})

	// applying CORS configurations
	r.Use(cors.Handler(corsOptions))

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	sseServer := sse.New()
	sseServer.CreateStream("rooms")
	sseServer.EventTTL = time.Second * 1
	sseServer.AutoReplay = false
	sseServer.AutoStream = true
	sseServer.Headers = map[string]string{
		"Content-Type":  "text/event-stream",
		"Cache-Control": "no-cache",
		"Connection":    "keep-alive",
	}
	// setup WS
	upgrader := websocket.Upgrader{
		ReadBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.ToLower(r.Header.Get("Origin"))

			if len(corsOptions.AllowedOrigins) == 0 || corsOptions.AllowedOrigins == nil {
				if corsOptions.AllowOriginFunc == nil {
					// reject all
					return false
				}
			}

			// assuming there's the function
			if corsOptions.AllowOriginFunc != nil {
				// use the function
				return corsOptions.AllowOriginFunc(r, origin)
			}

			// assuming there's value
			for _, allowedOrigin := range corsOptions.AllowedOrigins {
				if allowedOrigin == "*" {
					// assumed you've allowed all
					return true
				}

				o := strings.ToLower(allowedOrigin)

				// if they have *, use regex
				if i := strings.IndexByte(o, '*'); i >= 0 {
					// Split the origin in two: start and end string without the *
					w := wildcard{o[0:i], o[i+1:]}
					if w.match(o) {
						return true
					}
				}

				// assume as stirng
				if o == origin {
					return true
				}
			}

			return false
		},
	}

	// init database
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
			log.Default().Print("[ws] ", err)
			w.WriteHeader(500)
			fmt.Fprintf(w, "Couldn't upgarde the server with WS connection")
			return
		}

		defer conn.Close()

		db := r.Context().Value("_DB").(*sqlx.DB)
		audience := r.Context().Value("_SOCKET_AUDIENCE").(*Audience)
		inmemory := r.Context().Value("_INMEMORY").(*data.ChatterboxDB)

		// register to channel
		client := &Node{audience: audience, conn: conn}
		audience.register <- client

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

			service.OnReceiveMessage(db, inmemory, body, &service.MsgHandlers{
				WriteToUser: func(data []byte) {
					conn.WriteMessage(websocket.BinaryMessage, data)
				},
				WriteToAll: func(data []byte) {
					println("sent to all")
					audience.Broadcast(data)
				},
			})
		}
	})

	server := http.Server{
		Handler:           r,
		Addr:              fmt.Sprintf(":%v", os.Getenv("APP_PORT")),
		ReadHeaderTimeout: 3 * time.Second,
		BaseContext: func(l net.Listener) context.Context {
			// Chat related interfacing
			audience := SetupAudience()

			ctx := context.Background()

			ctx = context.WithValue(ctx, "_DB", db)
			ctx = context.WithValue(ctx, "_SOCKET_AUDIENCE", audience)
			ctx = context.WithValue(ctx, "_INMEMORY", data.CreateInMemoryInstance())

			go audience.Listen(ctx)

			return ctx
		},
	}

	log.Fatal(server.ListenAndServe())
}
