package controllers

import (
	"encoding/json"
	"errors"
	"github.com/lempiy/gochat/models"
	"github.com/lempiy/gochat/types/chatroom"
)

type RoomData struct {
	RoomID   int           `json:"room_id"`
	RoomSubs []models.User `json:"room_subs"`
}

type CreateRoomData struct {
	RoomSubIds []int `json:"room_sub_ids"`
}

func createRoomHandler(content string) {
	var crd *CreateRoomData
	json.Unmarshal([]byte(content), crd)
	//cr := createARoom(0)
	//TODO: Join room, need global Subs pool
}

func createARoom(id int) *chatroom.Chatroom {
	newRoom := chatroom.New(models.NewRoom(20, models.RoomTypeSpecial, id))
	Rooms.Add(newRoom)
	newRoom.Init()
	return newRoom
}

func initRoomAndSubscribe(sub *chatroom.Subscriber, room *models.Room) {
	newRoom := chatroom.New(room)
	Rooms.Add(newRoom)
	newRoom.Init()
	newRoom.Join(sub)
}

func connectToRooms(sub *chatroom.Subscriber, am Action) error {
	var data dataConnect
	err := json.Unmarshal([]byte(am.Text), &data)
	if err != nil {
		return err
	}
	for _, roomID := range data.Rooms {
		if room := Rooms.Get(roomID); room != nil {
			return errors.New("Unknown room " + string(roomID))
		} else {
			room.Join(sub)
		}
	}
	return nil
}

func joinRoomsInitial(sub *chatroom.Subscriber) {
	for _, room := range sub.User.Rooms {
		if r := Rooms.Get(room.Id); r != nil {
			r.Join(sub)
		} else {
			initRoomAndSubscribe(sub, room)
		}
	}
}

func leaveAllRooms(sub *chatroom.Subscriber) {
	for _, roomID := range sub.RoomIDs {
		if room := Rooms.Get(roomID); room != nil {
			room.Leave(sub.Id)
		}
	}
}
