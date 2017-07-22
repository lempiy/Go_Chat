package controllers

import (
	"github.com/astaxie/beego"
)

type AppController struct {
	beego.Controller
}

func (appCtrl *AppController) Get() {
	appCtrl.TplName = "index.html"
}
