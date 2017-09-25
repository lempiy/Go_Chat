package system

import (
	"github.com/lempiy/gochat/models"
)

var StandartMap = &map[string]func(e models.Event, s *Session) *Response{}
