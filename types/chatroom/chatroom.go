package chatroom

import (
	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/lempiy/gochat/models"
)

const Identifier = "room"

//Subscriber type represents rooms subscriber with unique ID and ws connection
type Subscriber struct {
	*models.User
	RoomIDs []int
	Conn    *websocket.Conn
}

type MessageResponse struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

//Chatroom is a general type for charooms that can be dynamically created on runtime
type Chatroom struct {
	ID          int
	Model       *models.Room
	subscribers map[int]*Subscriber
	subscribe   chan *Subscriber
	unsubscribe chan int
	publish     chan models.Event
	retrieve    chan int
	leave       chan int
}

//New is a chatroom constructor that generates room with random ID
func New(model *models.Room) *Chatroom {
	return &Chatroom{
		ID:          model.Id,
		Model:       model,
		subscribers: make(map[int]*Subscriber),
		subscribe:   make(chan *Subscriber, 10),
		unsubscribe: make(chan int, 10),
		retrieve:    make(chan int, 10),
		publish:     make(chan models.Event),
		leave:       make(chan int, 10)}
}

//Join method for joining new subs to room
func (chatroom *Chatroom) Join(sub *Subscriber) {
	chatroom.subscribe <- sub
}

//Leave method for exiting from the room
func (chatroom *Chatroom) Leave(ID int) {
	chatroom.unsubscribe <- ID
}

//Emit method triggers publish func
func (chatroom *Chatroom) Emit(e models.Event) {
	chatroom.publish <- e
}

//RetrieveEvents method retrieves publish func
func (chatroom *Chatroom) RetrieveEvents(ID int) {
	chatroom.retrieve <- ID
}

//Init func runs chatroom infinite loop in separate goroutine
func (chatroom *Chatroom) Init() {
	go chatroom.run()
}

func (chatroom *Chatroom) run() {
	for {
		select {
		case sub := <-chatroom.subscribe:
			if _, found := chatroom.subscribers[sub.Id]; !found {
				chatroom.subscribers[sub.Id] = sub
				sub.RoomIDs = append(sub.RoomIDs, chatroom.ID)
				if !sub.User.IsUserInRoom(chatroom.Model) {
					models.UpdateUserRooms(sub.User, chatroom.Model)
					beego.Info("New user: ", sub.User, "; WebSocket: ", sub.Conn != nil)
				} else {
					beego.Info("Old user: ", sub.User, "; WebSocket: ", sub.Conn != nil)
				}
			} else {
				beego.Info("Old user: ", sub.User, "; WebSocket: ", sub.Conn != nil)
			}

		case sID := <-chatroom.retrieve:
			events := chatroom.Model.GetAll()
			beego.Info(events)
			if sub, found := chatroom.subscribers[sID]; found {
				answer := MessageResponse{
					Type: "get",
					Data: events,
				}
				data, err := json.Marshal(answer)
				if err != nil {
					beego.Warning(err)
				}
				beego.Info("Write events JSON ", string(data))
				if sub.Conn.WriteMessage(websocket.TextMessage, data) != nil {
					chatroom.unsubscribe <- sub.Id
				}
			} else {
				beego.Info("User ", sID, " is not in a room.")
			}
			beego.Info("Get messages")

		case event := <-chatroom.publish:
			chatroom.Model.Add(&event)
			chatroom.broadcast(&event)
			beego.Info("New message: ", event.Content, " - ", event.User.Id)

		case unsub := <-chatroom.unsubscribe:
			if _, found := chatroom.subscribers[unsub]; found {
				beego.Info("Websocket closed: ", unsub)
				delete(chatroom.subscribers, unsub)
			} else {
				beego.Info("Cannot unsubscribe room ", chatroom.ID, " - user with ID ", unsub, " not found")
			}
		case subleave := <-chatroom.leave:
			if subscriber, found := chatroom.subscribers[subleave]; found {
				err := models.ExitUserRooms(subscriber.User, chatroom.Model)
				if err != nil {
					beego.Info("Cannot exit room ", chatroom.ID, " - user with ID ", subleave, " server error")
				}
				beego.Info("Websocket closed: ", subleave)
				delete(chatroom.subscribers, subleave)
			} else {
				beego.Info("Cannot exit room ", chatroom.ID, " - user with ID ", subleave, " not found")
			}
		}
	}
}

//broadcast func broadcasts messages to subs in room.
func (chatroom *Chatroom) broadcast(event *models.Event) {
	m := MessageResponse{
		Type: "message", //replace hardcode
		Data: event,
	}
	data, err := json.Marshal(&m)
	if err != nil {
		beego.Error("Error upon marshalling event data: ", err)
		return
	}

	for _, sub := range chatroom.subscribers {
		if sub.Conn != nil {
			if sub.Conn.WriteMessage(websocket.TextMessage, data) != nil {
				chatroom.unsubscribe <- sub.Id
			}
		}
	}
}
