package service

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"time"

	db "iam-kevin/chatterbox/data"

	"github.com/jmoiron/sqlx"
	"github.com/nrednav/cuid2"
)

// truncate the message box
const MAX_CHARACTERS = 160

func ProcessChat(db *sqlx.DB, cdb *db.ChatterboxDB, message db.ChatMessage, wptr *io.WriteCloser) {
	w := *wptr
	txt := strings.TrimSpace(message.Message)

	if len(txt) > MAX_CHARACTERS {
		txt = txt[:MAX_CHARACTERS]
	}

	// close connection
	defer w.Close()

	switch {
	case strings.HasPrefix(txt, "/join"):
		{
			fmt.Fprint(w, `{"message": "You are joining this group", "room": {
				"id": "123",
				"name": "TheCoolKidsClub"
			}}`)
			return
		}
	case strings.HasPrefix(txt, "/send"):
		{
			fmt.Println("Sending money")
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
				msg := OutgoingMessage{
					Type:       TypeMessage,
					Message:    "failed",
					Timestamp:  time.Now().UTC().Format(time.DateTime),
					UserId:     message.SenderId,
					OnlyByUser: true,
				}
				fmt.Fprintf(w, `%s`, msg.Json())
				return
			}

			msg := OutgoingMessage{
				Type:       TypeCreateRoom,
				Message:    fmt.Sprintf("Room '%s' has been created", chatroom.Name),
				Timestamp:  time.UTC.String(),
				RoomId:     &chatroom.Rid,
				UserId:     message.SenderId,
				OnlyByUser: true,
			}
			fmt.Fprintf(w, "%s", msg.Json())
			return
		}
	case txt == "/leave":
		{
			fmt.Println("Leave room")
			return
		}
	default:
		{
			out := OutgoingMessage{
				Type:       TypeMessage,
				Message:    message.Message,
				Timestamp:  time.Now().UTC().Format(time.DateTime),
				UserId:     message.SenderId,
				OnlyByUser: false,
			}
			fmt.Fprintf(w, `%s`, out.Json())
			return
		}
	}
}

const (
	TypeMessage    = "message"
	TypeCreateRoom = "create-room"
)

type OutgoingMessage struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	// optional
	RoomId     *string `json:"rid"`
	UserId     string  `json:"uid"`
	OnlyByUser bool    `json:"onlyme"`
}

func (o *OutgoingMessage) Json() []byte {
	message, _ := json.Marshal(
		o)

	return message
}

func CleanName(name string) string {
	roomName := strings.TrimSpace(name)
	re, _ := regexp.Compile(`\W`)
	return re.ReplaceAllString(roomName, "")
}

func ChatCreateRoom(db *sqlx.DB, name, senderId string) (*Chatroom, error) {
	genid := cuid2.Generate()
	id := "room_" + genid
	// name := "Room " + genid
	// values := strings.Split(txt, " ")

	// if len(values) > 2 {
	// 	name = values[1]
	// }

	// // generate id
	// cdb.CreateRoom(member, id, name, "", 10)
	var chatroom Chatroom
	err := db.Get(&chatroom, `insert into "room" (id, name, user_id) values ($1, $2, $3) returning id, name, user_id`, id, name, senderId)
	if err != nil {
		return nil, err
	}

	return &chatroom, nil
}

type Chatroom struct {
	Name    string `db:"name" json:"room"`
	OwnerId string `db:"user_id" json:"ownerId"`
	Rid     string `db:"id" json:"room_id"`
}
