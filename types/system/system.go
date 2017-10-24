package system

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/lempiy/gochat/models"
	"github.com/lempiy/gochat/types/chatroom"
	"github.com/lempiy/gochat/utils/token"
)

const Identifier = "system"

//Session type represents anonymous with unique uuid v4 and ws connection
type Session struct {
	Token string               `json:"token"`
	Conn  *websocket.Conn      `json:"-"`
	Sub   *chatroom.Subscriber `json:"-"` //nil if anonymous
}

//NewSession creates session with unique Token
func NewSession(con *websocket.Conn, tkn string) *Session {
	if tkn == "" {
		tkn, _ = token.GetAnonToken()
	}
	return &Session{
		Token: tkn,
		Conn:  con,
		Sub: &chatroom.Subscriber{
			Conn: con,
		},
	}
}

//Response is a basic struct for system reponse messages
type Response struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

//System is a global event broker for non-room events
type System struct {
	subscribers map[string]*Session
	subscribe   chan *Session
	unsubscribe chan string
	publish     chan models.Event
	retrieve    chan string
	eventMap    map[string]func(e models.Event, s *Session) *Response
	leave       chan string
}

//New is a system broker constructor
func New(eventMap *map[string]func(e models.Event, s *Session) *Response) *System {
	return &System{
		subscribers: make(map[string]*Session),
		subscribe:   make(chan *Session, 10),
		unsubscribe: make(chan string, 10),
		retrieve:    make(chan string, 10),
		publish:     make(chan models.Event),
		leave:       make(chan string, 10),
		eventMap:    *eventMap,
	}
}

//Join method for joining new actor to system
func (sys *System) Join(sub *Session) {
	sys.subscribe <- sub
}

//Leave method for exiting from the system
func (sys *System) Leave(Token string) {
	sys.unsubscribe <- Token
}

//Emit method triggers publish func
func (sys *System) Publish(e models.Event) {
	sys.publish <- e
}

//RetrieveEvents method retrieves publish func
func (sys *System) RetrieveEvents(Token string) {
	sys.retrieve <- Token
}

//Init func runs sys infinite loop in separate goroutine
func (sys *System) Init() {
	go sys.run()
}

func (sys *System) run() {
	for {
		select {
		case sub := <-sys.subscribe:
			if _, found := sys.subscribers[sub.Token]; !found {
				sys.subscribers[sub.Token] = sub
				r, _ := json.Marshal(struct {
					Type string      `json:"type"`
					Data interface{} `json:"data"`
				}{
					Type: "session",
					Data: sub,
				})
				if sub.Conn.WriteMessage(websocket.TextMessage, r) != nil {
					sys.unsubscribe <- sub.Token
				}
				beego.Info("NEW SESSION: ", sub.Token)
			} else {
				beego.Info("Old session: ", sub.Token, "; WebSocket: ", sub.Conn != nil)
			}

		case sToken := <-sys.retrieve:
			if sub, found := sys.subscribers[sToken]; found {
				if sub.Conn.WriteMessage(websocket.TextMessage, []byte("null")) != nil {
					sys.unsubscribe <- sub.Token
				}
			} else {
				beego.Info("User ", sToken, " is not in a system.")
			}
			beego.Info("Get messages")

		case event := <-sys.publish:
			beego.Info("New system event: ", event.Token)
			if session, exist := sys.subscribers[event.Token]; exist {
				if h, found := sys.eventMap[event.SysType]; found {
					res := h(event, session)
					r, _ := json.Marshal(res)
					if session.Conn.WriteMessage(websocket.TextMessage, r) != nil {
						sys.unsubscribe <- session.Token
					}
				}
			}
		case unsub := <-sys.unsubscribe:
			if _, found := sys.subscribers[unsub]; found {
				beego.Info("Websocket closed: ", unsub)
				delete(sys.subscribers, unsub)
			} else {
				beego.Info("Cannot unsubscribe from system - user with Token ", unsub, " not found")
			}
		case subleave := <-sys.leave:
			if _, found := sys.subscribers[subleave]; found {
				delete(sys.subscribers, subleave)
			} else {
				beego.Info("Cannot exit room - user with Token ", subleave, " not found")
			}
		}
	}
}
