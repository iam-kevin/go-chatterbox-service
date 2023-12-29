package messages

import (
	"encoding/json"
	"time"
)

type OutMessage struct {
	Type         string    `json:"type"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
	OnlyToSender bool      `json:"onlyme"`
	UserId       string    `json:"uid"`
	// optional
	RoomId *string `json:"rid"`
}

type simpleMessageToJson struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	UserId    string `json:"uid"`
	// optional
	OnlyToSender bool    `json:"onlyme"`
	RoomId       *string `json:"rid"`
}

func (sm *OutMessage) MarshalJSON() ([]byte, error) {
	var onlyuser bool = false
	if sm.OnlyToSender {
		onlyuser = sm.OnlyToSender
	}

	s := simpleMessageToJson{
		Type:         sm.Type,
		Message:      sm.Message,
		Timestamp:    sm.Timestamp.Format(time.DateTime),
		UserId:       sm.UserId,
		RoomId:       sm.RoomId,
		OnlyToSender: onlyuser,
	}

	out, err := json.Marshal(s)
	return out, err
}

const (
	TypeSimple     = "message"
	TypeCreateRoom = "create-room"
	TypeLeaveRoom  = "leave-room"
	TypeJoinRoom   = "join-room"
)

type Options struct {
	Message      string
	Type         string
	RoomId       *string
	UserId       string
	OnlyToSender bool
}

func New(obj *Options,
) OutMessage {
	var definedType string
	simple := TypeSimple

	if obj.Type != "" {
		definedType = obj.Type
	} else {
		definedType = simple
	}

	return OutMessage{
		Type:         definedType,
		Message:      obj.Message,
		RoomId:       obj.RoomId,
		UserId:       obj.UserId,
		Timestamp:    time.Now().UTC(),
		OnlyToSender: obj.OnlyToSender,
	}
}

func (o *OutMessage) Json() []byte {
	message, _ := json.Marshal(
		o)

	return message
}
