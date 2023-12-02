package chatterbox

import (
	"fmt"
	"io"
	"strings"
)

type Member struct {
	name     string
	username string
}

type Room struct {
	members    *[]Member
	name       string // Name of the chatroom
	desc       string // description of the chatroom
	id         string // chatroon id (used to join)
	MessageBox *MessageBox
}

/* Create a chatroom */
func CreateRoom(id, name, desc string, memberCount *int) *Room {
	var members []Member
	if memberCount != nil {
		members = make([]Member, *memberCount)
	} else {
		members = make([]Member, 10)
	}

	// messages
	messages := make([]Message, 10)

	return &Room{
		id:      id,
		name:    name,
		desc:    desc,
		members: &members,
		MessageBox: &MessageBox{
			count:    0,
			messages: &messages,
		},
	}
}

type MessageBox struct {
	count    int
	messages *[]Message
}

func (r *Room) ReadFromIndex(index int) *Message {
	p := *r.MessageBox.messages
	msg := p[index]
	return &msg
}

func (r *Room) NumOfMessages() int {
	return r.MessageBox.count
}

func ProcessRoomChat(roomId, text string, w *io.WriteCloser) {
	txt := strings.TrimSpace(text)
	switch {
	default:
		fmt.Printf("You typed \"%s\"\n", text)
		return
	case strings.HasPrefix(txt, "/send"):
		fmt.Println("Sending money")
		return
	case txt == "/leave":
		fmt.Println("Leave room")
		return
	case txt == "/leave":
		fmt.Println("Leave room")
		return
	}
}

func ProcessGlobalChat(text string) {
	txt := strings.TrimSpace(text)
	switch {
	default:
		fmt.Printf("You typed \"%s\"\n", text)
		return
	case strings.HasPrefix(txt, "/join"):
		fmt.Println("Joining room")
		return
	case strings.HasPrefix(txt, "/create"):
		fmt.Println("Creating a room")
		return
	}
}
