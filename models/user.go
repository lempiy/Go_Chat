package models

import (
	"time"
)

type User struct {
	Id        int        `json:"id"`
	Username  string     `json:"username" orm:"unique"`
	Password  string     `json:"password,omitempty"`
	Rooms     []*Room    `json:"rooms,omitempty" orm:"reverse(many)"`
	CreatedAt *time.Time `json:"created_at,string,omitempty" orm:"auto_now_add;type(timestamp)"`
}

func Create(username string, password string) (*User, error) {
	u := new(User)
	u.Username = username
	u.Password = password
	_, err := O.Insert(u)
	if err != nil {
		return nil, err
	}
	// Auto join global room
	err = UpdateUserRooms(u, &Room{
		Id: 1,
	})
	u.Password = ""
	return u, nil
}

func (u *User) IsUserInRoom(r *Room) bool {
	m2m := O.QueryM2M(u, "Rooms")
	return m2m.Exist(r)
}

func (u *User) LoadUsersRooms() error {
	_, err := O.LoadRelated(u, "Rooms")
	return err
}

func UpdateUserRooms(u *User, r *Room) error {
	m2m := O.QueryM2M(u, "Rooms")
	_, err := m2m.Add(r)
	return err
}

func ExitUserRooms(u *User, r *Room) error {
	m2m := O.QueryM2M(u, "Rooms")
	_, err := m2m.Remove(r)
	return err
}

//CreateOrRead creates new user or reads existing user by supplied struct
func CreateOrReadUser(u *User) error {
	_, _, err := O.ReadOrCreate(u, "Username", "Password")
	return err
}

func ReadUser(u *User, f ...string) error {
	return O.Read(u, f...)
}
