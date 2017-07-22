package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"time"
	"github.com/astaxie/beego/orm"
)

//
const (
	RoomTypeGeneral = iota
	RoomTypePrivate
	RoomTypeSpecial
)

//Room is a type for events storage in memory
type Room struct {
	Id        int       `json:"id"`
	Size      int       `json:"size,omitempty"`
	Type      int       `json:"type,omitempty"`
	CreatedAt *time.Time `json:"timestamp,omitempty" orm:"auto_now_add;type(timestamp)"`
	Events    []*Event  `json:"events,omitempty" orm:"reverse(many)"`
	Users     []*User   `json:"users,omitempty" orm:"rel(m2m)"`
}

//New is a pseudo-constructor that creates an EventModel of a particular size
func NewRoom(size int, eventModelType int, roomID int) *Room {

	room := &Room{Size: size, Type: eventModelType, Id: roomID}

	if roomID == 0 {
		O.Insert(room)
		return room
	}
	CreateOrReadRoom(room)
	fmt.Println(size, eventModelType, roomID)
	return room
}

//CreateOrRead creates new room or reads existing room by supplied struct
func CreateOrReadRoom(r *Room) error {
	fmt.Println(r)
	_, _, err := O.ReadOrCreate(r, "Id")
	return err
}

//Add method add event to event storage.
func (model *Room) Add(event *Event) {
	_, err := O.Insert(event)
	if err != nil {
		beego.Warning(err)
	}
	event.User.Password = ""
	beego.Info("Successfully insert event ", event.Content)
}

//GetAll method retrieves all existing messages from model limmited by size
func (model *Room) GetAll() []*Event {
	var events []*Event

	var maps []orm.Params
	_, err := O.QueryTable("event").
		Filter("Room__Id", model.Id).
		RelatedSel("User").
		Limit(model.Size).
		Values(&maps, "id", "type", "timestamp", "content", "room__id", "user__id", "user__username")
	if err != nil {
		beego.Warning(err)
		return nil
	}
	for _, m := range maps {
		t := (m["Timestamp"]).(time.Time)
		event := &Event{
			Id: int((m["Id"]).(int64)),
			Type: int((m["Type"]).(int64)),
			Timestamp: &t,
			Content: (m["Content"]).(string),
			Room: &Room{
				Id: int((m["Room__Id"]).(int64)),
			},
			User: &User{
				Id: int((m["User__Id"]).(int64)),
				Username: (m["User__Username"]).(string),
			},
		}
		events = append(events, event)
		// There is no complicated nesting data in the map
	}
	return events
}
