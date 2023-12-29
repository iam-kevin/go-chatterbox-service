package data

type MessageType int

const (
	Regular   MessageType = 0
	SendMoney MessageType = 1
)

type Message struct {
	Type    MessageType
	RawText string
	Sender  Member
}

// This is the shape of the message as sent by the user
type ChatMessage struct {
	RoomId   *string `json:"room"`
	Pin      *string `json:"pin"`
	Message  string  `json:"message"`
	SenderId string  `json:"senderId"`
}

type MessageContent struct {
	Message  string
	Receiver *Member
	Amount   int
}

type Member struct {
	Name     string
	Username string
}

type Room struct {
	members    []Member
	owner      Member
	name       string // Name of the chatroom
	id         string // chatroon id (used to join)
	MessageBox *MessageBox
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
