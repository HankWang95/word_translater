package main

import (
	"github.com/smartwalle/dbs"
	"github.com/smartwalle/ini4go"
	"log"
	"os"
	"path"
)

func main() {
	runDailyWork()
	select {

	}
}


var (
	Logger      *log.Logger
	db          dbs.DB
	projectPath = path.Join(os.Getenv("HOME"), "Documents", "Kanna")
	configPath  = path.Join(projectPath, "config")
	speechPath  = path.Join(projectPath, "speech")
)

func init() {
	// 初始化log文件
	logFile, err := os.OpenFile(path.Join(projectPath, "kanna.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	Logger = log.New(logFile, "[word_translator] ", log.Ltime|log.Ldate|log.Lshortfile)

	// 读取config
	var config = ini4go.New(false)
	config.SetUniqueOption(true)
	err = config.Load(configPath)
	if err != nil {
		panic(err)
	}

	// 生成文件及目录
	_, err = os.Stat(projectPath)
	if err != nil {
		err := os.Mkdir(projectPath, os.ModeDir|0777)
		if err != nil {
			Logger.Fatal(err, projectPath)
		}
		_, err = os.Stat(speechPath)
		if err != nil {
			err = os.Mkdir(speechPath, os.ModeDir|0777)
			if err != nil {
				panic(err)
			}
		}
	}

	// 初始化mysql
	db, err = dbs.NewSQL(config.GetValue("sql", "driver"),
		config.GetValue("sql", "url"),
		config.MustInt("sql", "max_open", 10),
		config.MustInt("sql", "max_idle", 5))
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

}

func getMySQLSession() dbs.DB {
	return db
}
