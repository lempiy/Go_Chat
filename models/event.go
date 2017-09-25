package models

import "time"

//Event is a type for every events in chat
type Event struct {
	Id        int        `json:"id"`
	Type      int        `json:"type, omitempty"`
	SysType   string     `json:"sys_type, omitempty"`
	Timestamp *time.Time `json:"timestamp, omitempty" orm:"auto_now_add;type(timestamp)"`
	Content   string     `json:"content, omitempty"`
	Room      *Room      `json:"room" orm:"rel(fk)"`
	User      *User      `json:"user" orm:"rel(fk);null;on_delete(set_null)"`
}

//List of constants for methods types
const (
	EventJoin = iota
	EventLeave
	EventMessage
)
