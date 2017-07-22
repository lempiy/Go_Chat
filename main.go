package main

import (
	"github.com/astaxie/beego"
	"github.com/lempiy/gochat/controllers"
	_ "github.com/lempiy/gochat/models"
)

func main() {
	beego.Router("/", &controllers.AppController{})
	beego.Router("/ws/join/", &controllers.WebSocketCtrl{})
	beego.Run()
}
