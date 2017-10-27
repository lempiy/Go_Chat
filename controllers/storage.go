package controllers

import (
	"github.com/lempiy/gochat/types/chatroom"
	"sync"
)

type RoomsList struct {
	locker sync.RWMutex
	store  map[int]*chatroom.Chatroom
}

var Rooms = NewRoomsList()

func NewRoomsList() *RoomsList {
	return &RoomsList{
		locker: sync.RWMutex{},
		store:  make(map[int]*chatroom.Chatroom, 0),
	}
}

func (r RoomsList) Add(cr *chatroom.Chatroom) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.store[cr.ID] = cr
}

func (r RoomsList) Get(id int) *chatroom.Chatroom {
	r.locker.Lock()
	defer r.locker.Unlock()
	cr := r.store[id]
	return cr
}

func (r RoomsList) Remove(id int) {
	r.locker.Lock()
	defer r.locker.Unlock()
	delete(r.store, id)
}

type SubsList struct {
	locker sync.RWMutex
	store  map[int]*chatroom.Subscriber
}

var Subs = NewSubsList()

func NewSubsList() *SubsList {
	return &SubsList{
		locker: sync.RWMutex{},
		store:  make(map[int]*chatroom.Subscriber, 0),
	}
}

func (r SubsList) Add(sub *chatroom.Subscriber) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.store[sub.Id] = sub
}

func (r SubsList) Get(id int) *chatroom.Subscriber {
	r.locker.Lock()
	defer r.locker.Unlock()
	sub := r.store[id]
	return sub
}

func (r SubsList) Remove(id int) {
	r.locker.Lock()
	defer r.locker.Unlock()
	delete(r.store, id)
}
