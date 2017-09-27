package controllers

import (
	"encoding/json"
	"github.com/lempiy/gochat/models"
	"github.com/lempiy/gochat/utils/token"
)

type loginData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
}

func login(l string) (loginResponse, *models.User) {
	ld := loginData{}
	err := json.Unmarshal([]byte(l), &ld)
	if err != nil {
		return loginResponse{}, nil
	}
	u := models.User{
		Username: ld.Username,
		Password: ld.Password,
	}
	err = models.ReadUser(&u, "Username", "Password")
	if err != nil || u.Id == 0 {
		return loginResponse{}, nil
	}
	t, _ := token.GetToken(u.Username, u.Id)

	return loginResponse{
		Success: true,
		Token:   t,
	}, &u
}

func register(l string) (loginResponse, *models.User) {
	ld := loginData{}
	err := json.Unmarshal([]byte(l), &ld)
	if err != nil {
		return loginResponse{}, nil
	}
	u, err := models.Create(ld.Username, ld.Password)
	if err != nil || u.Id == 0 {
		return loginResponse{}, nil
	}
	t, _ := token.GetToken(u.Username, u.Id)
	return loginResponse{
		Success: true,
		Token:   t,
	}, u
}
