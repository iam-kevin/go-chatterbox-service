package service

import (
	"fmt"
	"io"
	"strings"
	"time"

	db "iam-kevin/chatterbox/data"

	"github.com/nrednav/cuid2"
)

// truncate the message box
const MAX_CHARACTERS = 160

func ProcessChat(cdb *db.ChatterboxDB, member *db.Member, message db.ChatMessage, wptr *io.WriteCloser) {
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
			fmt.Fprint(w, `{"message": "You are joining this group"}`)
			return
		}
	case strings.HasPrefix(txt, "/send"):
		{
			fmt.Println("Sending money")
			return
		}
	case strings.HasPrefix(txt, "/create"):
		{
			genid := cuid2.Generate()
			id := "room_" + genid
			name := "Room " + genid
			values := strings.Split(txt, " ")

			if len(values) > 2 {
				name = values[1]
			}

			// generate id
			cdb.CreateRoom(member, id, name, "", 10)
			fmt.Println("Room created")
			return
		}
	case txt == "/leave":
		{
			fmt.Println("Leave room")
			return
		}
	default:
		{
			fmt.Fprintf(w, `{"message": "%v", "time": "%v" }`, txt, time.Now().UTC())
			return
		}
	}
}
