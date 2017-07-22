package controllers

import (
	"net/http"

	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/lempiy/gochat/models"
	"github.com/lempiy/gochat/types"
	"github.com/lempiy/gochat/types/chatroom"
)

var globalRoom *chatroom.Chatroom

//WebSocketCtrl deals with websocket connections and data transfer.
type WebSocketCtrl struct {
	beego.Controller
}

type RoomsList map[int]*chatroom.Chatroom

var Rooms RoomsList

func init() {
	Rooms = make(RoomsList)
	globalRoom = chatroom.New(models.NewRoom(20, models.RoomTypeGeneral, 1))
	Rooms[globalRoom.ID] = globalRoom
	globalRoom.Init()
}

func initRoomAndSubscribe(sub *chatroom.Subscriber, room *models.Room) {
	newRoom := chatroom.New(room)
	Rooms[newRoom.ID] = newRoom
	newRoom.Init()
	newRoom.Join(sub)
}

//Get methods establishes connection with users.
func (wsc *WebSocketCtrl) Get() {
	wsc.TplName = "index.html"
	id := wsc.GetString("id")
	if len(id) == 0 {
		wsc.Redirect("/", 302)
		return
	}

	ws, err := websocket.Upgrade(wsc.Ctx.ResponseWriter, wsc.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(wsc.Ctx.ResponseWriter, "Unexpected connection - expecting Websocket handshake", 400)
		return
	} else if err != nil {
		http.Error(wsc.Ctx.ResponseWriter, "Server Internal Error", 500)
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}
	u := &models.User{
		Username: id,
		Password: "q1w2e3r4",
	}

	err = models.CreateOrReadUser(u)

	if err != nil {
		fmt.Println(err)
	}

	u.Password = ""

	sub := &chatroom.Subscriber{
		User: u,
		Conn: ws}

	sub.User.LoadUsersRooms()

	joinRoomsInitial(sub)

	globalRoom.Join(sub)
	defer leaveAllRooms(sub)

	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			ws.Close()
			//TODO: Refactor to use no beego controller here. Just a custom handler,
			//to prevent http.Write calls on to hijacked connection.
			wsc.StopRun()
			return
		}

		if string(data) == "get" {
			globalRoom.RetrieveEvents(sub.Id)
		} else {
			e := models.Event{
				Type:    models.EventMessage,
				User:    sub.User,
				Content: string(data),
				Room:    globalRoom.Model}

			globalRoom.Emit(e)
		}
	}
}

/*
{
	"action": "connect"
	"data": {
		"rooms": []
	}
}
*/

type dataConnect struct {
	Rooms []int `json:"rooms"`
}

func connectToRooms(sub *chatroom.Subscriber, am types.ActionMessage) error {
	var data dataConnect
	err := json.Unmarshal([]byte(am.Data), &data)
	if err != nil {
		return err
	}
	for _, roomID := range data.Rooms {
		if room, ok := Rooms[roomID]; !ok {
			return errors.New("Unknown room " + string(roomID))
		} else {
			room.Join(sub)
		}
	}
	return nil
}

func joinRoomsInitial(sub *chatroom.Subscriber) {
	for _, room := range sub.User.Rooms {
		if r, exist := Rooms[room.Id]; exist {
			r.Join(sub)
		} else {
			initRoomAndSubscribe(sub, room)
		}
	}
}

func leaveAllRooms(sub *chatroom.Subscriber) {
	for _, roomID := range sub.RoomIDs {
		if room, ok := Rooms[roomID]; ok {
			room.Leave(sub.Id)
		}
	}
	sub.Conn.Close()
}
