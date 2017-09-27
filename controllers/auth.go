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
	Success bool `json:"success"`
	Token string `json:"token,omitempty"`
}

func login(l string) loginResponse {
	ld := loginData{}
	err := json.Unmarshal([]byte(l), &ld)
	if err != nil {
		return loginResponse{}
	}
	u := models.User{
		Username: ld.Username,
		Password: ld.Password,
	}
	err = models.ReadUser(&u)
	if err != nil || u.Id == 0 {
		return loginResponse{}
	}
	t, _ := token.GetToken(u.Username, u.Id)
	return loginResponse{
		Success: true,
		Token: t,
	}
}
