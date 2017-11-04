package controllers

import (
	"net/http"

	"encoding/json"
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

type Action struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Target   string `json:"target"`
	TargetId int    `json:"target_id,omitempty"`
}

var StandartMap = &map[string]func(e models.Event, s *system.Session) *system.Response{
	"login": func(e models.Event, s *system.Session) *system.Response {
		if s.Sub.User != nil {
			return &system.Response{Type: "login", Data: loginResponse{
				Success: true,
				Token:   s.Token,
			}}
		}
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
		if s.Sub.User != nil {
			return &system.Response{Type: "register", Data: loginResponse{
				Success: false,
			}}
		}
		r, user := register(e.Content)
		if r.Success {
			sys.Leave(s.Token)
			s.Token = r.Token
			sys.Join(s)
			upgrade2Sub(s, user)
		}
		return &system.Response{Type: "register", Data: r}
	},
	"logout": func(e models.Event, s *system.Session) *system.Response {
		if s.Sub.User == nil {
			return &system.Response{Type: "logout", Data: false}
		}
		downgrade2Ses(s)
		return &system.Response{Type: "logout", Data: s.Token}
	},
	"create_room": func(e models.Event, s *system.Session) *system.Response {
		if s.Sub.User != nil {
			var cd CreateRoomData
			var rData RoomData
			r := createARoom(0)
			json.Unmarshal([]byte(e.Content), &cd)
			rData.RoomID = r.ID
			for _, id := range cd.RoomSubIds {
				if sub := Subs.Get(id); sub != nil {
					r.Join(sub)
					rData.RoomSubs = append(rData.RoomSubs, *sub.User)
					rData.RoomMembers = append(rData.RoomMembers, *sub.User)
				} else {
					var usr *models.User
					usr.Id = id
					models.ReadUser(usr)
					if usr.Username != "" {
						models.UpdateUserRooms(usr, r.Model)
						rData.RoomSubs = append(rData.RoomSubs, *sub.User)
						rData.RoomMembers = append(rData.RoomMembers, *sub.User)
					}
				}
			}
			bts, err := json.Marshal(rData)
			if err != nil {
				return &system.Response{Type: "create_room", Data: false}
			}
			joinedEvent := models.Event{
				Type:    models.EventJoin,
				Content: string(bts),
				Room:    r.Model,
				NoSave:  true,
			}
			r.Emit(joinedEvent)
			return &system.Response{Type: "create_room", Data: rData}
		}
		return &system.Response{Type: "create_room", Data: false}
	},
}

func downgrade2Ses(session *system.Session) error {
	leaveAllRooms(session.Sub)
	session.Sub = &chatroom.Subscriber{
		Conn: session.Conn,
	}
	session.Token, _ = token.GetAnonToken()
	return nil
}

func upgrade2Sub(session *system.Session, u *models.User) error {
	session.Sub.User = u
	err := session.Sub.User.LoadUsersRooms()
	if err != nil {
		return err
	}
	joinRoomsInitial(session.Sub)
	globalRoom.Join(session.Sub)
	Subs.Add(session.Sub)
	return nil
}

func init() {
	sys = system.New(StandartMap)
	sys.Init()

	globalRoom = chatroom.New(models.NewRoom(20, models.RoomTypeGeneral, 1))
	Rooms.Add(globalRoom)
	globalRoom.Init()

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
	fmt.Println("TOKEN", isValid, username)
	if !isValid {
		session = system.NewSession(ws, "")
	} else {
		session = system.NewSession(ws, t)
		if username != "" {
			u := &models.User{
				Username: username,
			}

			err = models.ReadUser(u, "username")

			if err != nil {
				fmt.Println(err)
			}

			u.Password = ""

			sub = session.Sub
			sub.User = u
			Subs.Add(sub)

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
		} else if session.Sub.User != nil {
			sub = session.Sub
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
					if r := Rooms.Get(message.TargetId); r != nil {
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
