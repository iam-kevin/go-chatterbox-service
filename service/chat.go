package service

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	db "iam-kevin/chatterbox/data"
	messages "iam-kevin/chatterbox/message"

	"github.com/jmoiron/sqlx"
	"github.com/nrednav/cuid2"
)

// truncate the message box
const MAX_CHARACTERS = 160

type MsgHandlers struct {
	WriteToUser func(data []byte)
	WriteToAll  func(data []byte)
}

func OnReceiveMessage(db *sqlx.DB, cdb *db.ChatterboxDB, message db.ChatMessage, handle *MsgHandlers) {
	txt := strings.TrimSpace(message.Message)

	if len(txt) > MAX_CHARACTERS {
		txt = txt[:MAX_CHARACTERS]
	}

	switch {
	// shouts to everyone that user has joined group
	// pattern is /join <room-id>
	case strings.HasPrefix(txt, "/join"):
		{
			roomId := strings.Replace(txt, "/join", "", 1)

			// search the room
			var room Chatroom
			db.Get(&room, `select * from "room" where id = $1`, roomId)

			// TODO: save in cache

			// checks if user in room already
			out := messages.New(
				&messages.Options{
					Type:    messages.TypeJoinRoom,
					Message: fmt.Sprintf("%s has joined the room", message.SenderId),
					RoomId:  message.RoomId,
					UserId:  message.SenderId,
				},
			)

			handle.WriteToAll(out.Json())
			// fmt.Fprintf(w, `%s`, out.Json())
			return
		}
	// expected pattern of this chat message is
	//  /send <amount> <username>
	//  the <username> should not be of self
	case strings.HasPrefix(txt, "/send"):
		{
			if message.Pin == nil {
				out := messages.New(
					&messages.Options{
						Type:         messages.TypeJoinRoom,
						Message:      ("You need to send PIN before we can close this feature"),
						RoomId:       message.RoomId,
						UserId:       message.SenderId,
						OnlyToSender: true,
					},
				)
				handle.WriteToUser(out.Json())
			}

			out := messages.New(
				&messages.Options{
					Type:    messages.TypeJoinRoom,
					Message: ("We are still building out this feature"),
					// Message: fmt.Sprintf("%s has joined the room", message.SenderId),
					RoomId: message.RoomId,
					UserId: message.SenderId,
				},
			)

			handle.WriteToUser(out.Json())
			return
		}
	// expected pattern of this chat is
	//  /create <room-id>
	case strings.HasPrefix(txt, "/create"):
		{
			// extract string from text
			roomName := strings.Replace(txt, "/create", "", 1)

			// extract chat room string
			chatroom, err := ChatCreateRoom(db, CleanName(roomName), message.SenderId)
			if err != nil {
				log.Default().Print(err)
				msg := messages.New(
					&messages.Options{
						Type:         messages.TypeCreateRoom,
						Message:      fmt.Sprintf("Couldn't create room '%s'", chatroom.Name),
						UserId:       message.SenderId,
						RoomId:       message.RoomId,
						OnlyToSender: true,
					},
				)
				handle.WriteToUser(msg.Json())
				return
			}

			msg := messages.New(
				&messages.Options{
					Type:         messages.TypeCreateRoom,
					Message:      fmt.Sprintf("Room '%s' has been created", chatroom.Name),
					RoomId:       message.RoomId,
					UserId:       message.SenderId,
					OnlyToSender: true,
				},
			)
			handle.WriteToUser(msg.Json())
			return
		}
	case txt == "/leave":
		{
			out := messages.New(
				&messages.Options{
					Type:    messages.TypeLeaveRoom,
					Message: fmt.Sprintf("%s has left the room", message.SenderId),
					UserId:  message.SenderId,
					RoomId:  message.RoomId,
				},
			)
			handle.WriteToAll(out.Json())
			return
		}
	// echo's to everyone
	default:
		{
			out := messages.New(
				&messages.Options{
					Message: message.Message,
					UserId:  message.SenderId,
					RoomId:  message.RoomId,
				},
			)

			// save message

			handle.WriteToAll(out.Json())
			return
		}
	}
}

func CleanName(name string) string {
	roomName := strings.TrimSpace(name)
	re, _ := regexp.Compile(`\W`)
	return re.ReplaceAllString(roomName, "")
}

func ChatCreateRoom(db *sqlx.DB, name, senderId string) (*Chatroom, error) {
	genid := cuid2.Generate()
	id := "room_" + genid

	var chatroom Chatroom
	err := db.Get(&chatroom, `insert into "room" (id, name, user_id) values ($1, $2, $3) returning id, name, user_id`, id, name, senderId)
	if err != nil {
		return nil, err
	}

	return &chatroom, nil
}

type Chatroom struct {
	Name      string  `db:"name" json:"room_name"`
	OwnerId   string  `db:"user_id" json:"owner_nickname"`
	Rid       string  `db:"id" json:"room_id"`
	Timestamp *string `db:"created_at" json:"timestamp"`
}
