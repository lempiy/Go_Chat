package models

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	_ "github.com/lib/pq"
	"os"
	"time"
)

//O global ORM struct
var O orm.Ormer

func init() {
	orm.RegisterDriver("postgres", orm.DRPostgres)
	orm.RegisterModel(new(Event), new(Room), new(User))
	orm.Debug = true
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s host=postgres port=5432 sslmode=disable",
		os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_DB"))
	orm.DefaultTimeLoc = time.UTC
	err := orm.RegisterDataBase("default", "postgres", dbinfo)
	newORM()
	//err = orm.RunSyncdb("default", true, true)
	beego.Info("DB errors", err)
}

//NewORM creates new ORM instance
func newORM() {
	O = orm.NewOrm()
	O.Using("default")
}
