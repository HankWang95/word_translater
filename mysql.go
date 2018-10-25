package server

import (
	"github.com/smartwalle/dbs"
	"github.com/smartwalle/ini4go"
	"path"
	"os"
)

var db dbs.DB
var projectPath = path.Join(os.Getenv("HOME"), "Documents", "Kanna")


func InitMySQL() {
	var config = ini4go.New(false)
	config.SetUniqueOption(true)
	config.Load(path.Join(projectPath, "config"))

	var err error
	// todo 配置文件
	db, err = dbs.NewSQL(config.GetValue("sql", "driver"),
		config.GetValue("sql", "url"),
		config.MustInt("sql", "max_open", 10),
		config.MustInt("sql", "max_idle", 5))
	if err != nil {
		panic(err)
	}
}

func GetMySQLSession() dbs.DB {
	return db
}
