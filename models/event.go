package models

import "time"

//Event is a type for every events in chat
type Event struct {
	Id        int        `json:"id"`
	Type      int        `json:"type,omitempty"`
	SysType   string     `json:"sys_type,omitempty"`
	Timestamp *time.Time `json:"timestamp,omitempty" orm:"auto_now_add;type(timestamp)"`
	Content   string     `json:"content,omitempty" orm:"type(json)"`
	Room      *Room      `json:"room" orm:"rel(fk)"`
	User      *User      `json:"user,omitempty" orm:"rel(fk);null;on_delete(set_null)"`
	Token     string     `json:"-" orm:"-"`
	NoSave    bool       `json:"-" orm:"-"`
}

//List of constants for methods types
const (
	EventJoin = iota
	EventLeave
	EventMessage
)
