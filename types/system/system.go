package system

import (
	"encoding/json"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/lempiy/gochat/models"
	"github.com/satori/go.uuid"
)

//Session type represents anonymous with unique uuid v4 and ws connection
type Session struct {
	UUID string          `json:"uuid"`
	Conn *websocket.Conn `json:"-"`
	Sub  *models.User    `json:"-"` //nil if anonymous
}

//NewSession creates session with unique UUID
func NewSession(con *websocket.Conn) *Session {
	return &Session{
		UUID: uuid.NewV4().String(),
		Conn: con,
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
func (sys *System) Leave(UUID string) {
	sys.unsubscribe <- UUID
}

//Emit method triggers publish func
func (sys *System) Emit(e models.Event) {
	sys.publish <- e
}

//RetrieveEvents method retrieves publish func
func (sys *System) RetrieveEvents(UUID string) {
	sys.retrieve <- UUID
}

//Init func runs sys infinite loop in separate goroutine
func (sys *System) Init() {
	go sys.run()
}

func (sys *System) run() {
	for {
		select {
		case sub := <-sys.subscribe:
			if _, found := sys.subscribers[sub.UUID]; !found {
				sys.subscribers[sub.UUID] = sub
				r, _ := json.Marshal(sub)
				if sub.Conn.WriteMessage(websocket.TextMessage, r) != nil {
					sys.unsubscribe <- sub.UUID
				}
				beego.Info("NEW SESSION: ", sub.UUID)
			} else {
				beego.Info("Old session: ", sub.UUID, "; WebSocket: ", sub.Conn != nil)
			}

		case sUUID := <-sys.retrieve:
			if sub, found := sys.subscribers[sUUID]; found {
				if sub.Conn.WriteMessage(websocket.TextMessage, []byte("null")) != nil {
					sys.unsubscribe <- sub.UUID
				}
			} else {
				beego.Info("User ", sUUID, " is not in a system.")
			}
			beego.Info("Get messages")

		case event := <-sys.publish:
			var sub *Session
			var exist bool
			json.Unmarshal([]byte(event.Content), sub)
			sub, exist = sys.subscribers[sub.UUID]
			if h, found := sys.eventMap[event.SysType]; found && exist {
				h(event, sub)
			}
			beego.Info("New system event: ", event.Content)

		case unsub := <-sys.unsubscribe:
			if _, found := sys.subscribers[unsub]; found {
				beego.Info("Websocket closed: ", unsub)
				delete(sys.subscribers, unsub)
			} else {
				beego.Info("Cannot unsubscribe from system - user with UUID ", unsub, " not found")
			}
		case subleave := <-sys.leave:
			if _, found := sys.subscribers[subleave]; found {
				delete(sys.subscribers, subleave)
			} else {
				beego.Info("Cannot exit room - user with UUID ", subleave, " not found")
			}
		}
	}
}
