package chatterbox

import (
	"fmt"
	"strings"
)

type MessageType int

const (
	Regular   MessageType = 0
	SendMoney MessageType = 1
)

func DecodeText(text string) {
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

type Message struct {
	Type    MessageType
	RawText string
	Sender  Member
}

type MessageContent struct {
	Message  string
	Receiver *Member
	Amount   int
}

func (m *Message) Content() MessageContent {
	switch m.Type {
	default:
		return MessageContent{
			Message: m.RawText,
		}
	case SendMoney:
		return MessageContent{
			Receiver: &Member{name: "John Doe", username: "john-doe"},
			Amount:   200,
			Message:  m.RawText,
		}
	}
}
