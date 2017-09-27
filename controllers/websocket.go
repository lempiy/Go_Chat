package controllers

import (
	"net/http"

	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/lempiy/gochat/models"
	"github.com/lempiy/gochat/types/chatroom"
	"github.com/lempiy/gochat/types/system"
	"github.com/lempiy/gochat/utils/token"
)

var globalRoom *chatroom.Chatroom
var sys *system.System

//WebSocketCtrl deals with websocket connections and data transfer.
type WebSocketCtrl struct {
	beego.Controller
}

type RoomsList map[int]*chatroom.Chatroom

type Action struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Target   string `json:"target"`
	TargetId int    `json:"target_id,omitempty"`
}

var Rooms RoomsList

var StandartMap = &map[string]func(e models.Event, s *system.Session) *system.Response{
	"login": func(e models.Event, s *system.Session) *system.Response {
		r, user := login(e.Content)
		if r.Success {
			sys.Leave(s.Token)
			s.Token = r.Token
			sys.Join(s)
			upgrade2Sub(s, user)
		}
		return &system.Response{Type: "login", Data: r}
	},
	"register": func(e models.Event, s *system.Session) *system.Response {
		r, user := register(e.Content)
		if r.Success {
			sys.Leave(s.Token)
			s.Token = r.Token
			sys.Join(s)
			upgrade2Sub(s, user)
		}
		return &system.Response{Type: "register", Data: r}
	},
}

func upgrade2Sub(session *system.Session, u *models.User) error {
	session.Sub.User = u
	err := session.Sub.User.LoadUsersRooms()
	if err != nil {
		return err
	}
	joinRoomsInitial(session.Sub)
	globalRoom.Join(session.Sub)
	return nil
}

func init() {
	sys = system.New(StandartMap)
	sys.Init()

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
	t := wsc.GetString("token")
	//if len(t) == 0 {
	//	wsc.Redirect("/", 302)
	//	return
	//}

	ws, err := websocket.Upgrade(wsc.Ctx.ResponseWriter, wsc.Ctx.Request, nil, 1024, 1024)
	if _, ok := err.(websocket.HandshakeError); ok {
		http.Error(wsc.Ctx.ResponseWriter, "Unexpected connection - expecting Websocket handshake", 400)
		return
	} else if err != nil {
		http.Error(wsc.Ctx.ResponseWriter, "Server Internal Error", 500)
		beego.Error("Cannot setup WebSocket connection:", err)
		return
	}
	var session *system.Session
	var sub *chatroom.Subscriber
	fmt.Println("TOKENs", t)
	isValid, username := token.ValidateToken(t)

	if !isValid {
		session = system.NewSession(ws, "")
	} else {
		session = system.NewSession(ws, t)
		if username != "" {
			u := &models.User{
				Username: username,
				Password: "q1w2e3r4",
			}

			err = models.ReadUser(u)

			if err != nil {
				fmt.Println(err)
			}

			u.Password = ""

			sub = session.Sub
			sub.User = u

			sub.User.LoadUsersRooms()
			joinRoomsInitial(sub)
			globalRoom.Join(sub)
		}
	}

	sys.Join(session)
	defer leaveAllRooms(session.Sub)
	for {
		_, data, err := ws.ReadMessage()
		if err != nil {
			ws.Close()
			//TODO: Refactor to use no beego controller here. Just a custom handler,
			//to prevent http.Write calls on to hijacked connection.
			wsc.StopRun()
			return
		}

		message := Action{}
		json.Unmarshal(data, &message)

		if message.Target == system.Identifier {
			fmt.Println("MESSSAGE", message)
			sys.Publish(models.Event{
				SysType: message.Type,
				Content: message.Text,
				Token:   session.Token,
				Room:    globalRoom.Model})
		} else if username != "" {
			switch message.Type {
			case "get":
				globalRoom.RetrieveEvents(sub.Id)
			case "connect":
				connectToRooms(sub, message)
			default:
				if message.TargetId == 0 {
					e := models.Event{
						Type:    models.EventMessage,
						User:    sub.User,
						Token:   session.Token,
						Content: message.Text,
						Room:    globalRoom.Model}

					globalRoom.Emit(e)
				} else {
					if r, exist := Rooms[message.TargetId]; exist {
						e := models.Event{
							Type:    models.EventMessage,
							User:    sub.User,
							Content: message.Text,
							Token:   session.Token,
							Room:    r.Model}

						r.Emit(e)
					}
				}
			}
		} else {
			ws.WriteMessage(websocket.TextMessage, []byte(`{"error": "Non authorized access"}`))
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

func connectToRooms(sub *chatroom.Subscriber, am Action) error {
	var data dataConnect
	err := json.Unmarshal([]byte(am.Text), &data)
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
