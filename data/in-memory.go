package data

import (
	"fmt"
	"sync"
)

type ChatterboxDB struct {
	M     *MemberDatastore
	Rooms *RoomDatastore
}

func CreateInMemoryInstance() *ChatterboxDB {
	return &ChatterboxDB{
		M:     CreateMemberDB(),
		Rooms: CreateRoomDB(),
	}
}

// Whener all the members are stored
type MemberDatastore struct {
	members map[string]*Member
}

func (m *MemberDatastore) GetMember(name string) (*Member, error) {
	val, ok := m.members[name]
	if !ok {
		return nil, fmt.Errorf("not-found: this user doesn't exist")
	}

	return val, nil
}

func CreateMemberDB() *MemberDatastore {
	membersdb := MemberDatastore{members: make(map[string]*Member, 10)}
	return &membersdb
}

func (r *ChatterboxDB) CreateRoom(owner *Member, id, name, desc string, memberCount int) *Room {
	// messages
	messages := make([]Message, 10)

	room := Room{
		id:    id,
		name:  name,
		desc:  desc,
		owner: *owner,
		MessageBox: &MessageBox{
			count:    0,
			messages: &messages,
		},
	}

	// add members
	r.M.members[owner.Username] = owner

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
