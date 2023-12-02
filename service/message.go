package chatterbox

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
