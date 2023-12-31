package data

import (
	"fmt"
	"sync"
)

type ChatterboxDB struct {
	Rooms *RoomDatastore
}

func CreateInMemoryInstance() *ChatterboxDB {
	return &ChatterboxDB{
		Rooms: CreateRoomDB(),
	}
}

func (r *ChatterboxDB) CreateRoom(owner *Member, id, name, desc string, memberCount int) *Room {
	// messages
	messages := make([]Message, 10)

	room := Room{
		id:    id,
		name:  name,
		owner: *owner,
		MessageBox: &MessageBox{
			count:    0,
			messages: &messages,
		},
	}

	// add rooms
	r.Rooms.rooms[id] = room
	return &room
}

// generic definition
type RoomDatastore struct {
	count int
	mu    sync.Mutex
	// rooms that storage posses
	rooms map[string]Room
}

func CreateRoomDB() *RoomDatastore {
	roomsdb := RoomDatastore{count: 0, mu: sync.Mutex{}, rooms: make(map[string]Room, 10)}
	return &roomsdb
}

func (db *RoomDatastore) GetAllRooms() []*Room {
	refs := make([]*Room, 10)
	for _, r := range db.rooms {
		refs = append(refs, &r)
	}

	return refs
}

func (db *RoomDatastore) GetRoom(roomId string) (*Room, error) {
	room, ok := db.rooms[roomId]
	if !ok {
		return nil, fmt.Errorf("there's no such room")
	}

	return &room, nil
}
