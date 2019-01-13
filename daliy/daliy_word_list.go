package main

import (
	"github.com/HankWang95/word_translator/word"
	"github.com/robfig/cron"
	"os"
	"path"
	"time"
)

var (
	// todo get path from config
	projectDir = path.Join(os.Getenv("HOME"), "Documents", "Kanna")
	dailyDir = path.Join(projectDir, "daily")
)

func runDailyWork() {
	c := cron.New()
	c.AddFunc("0 0 19 * * *", dailyFunc)
	c.Start()
}

func dailyFunc()  {
	// 生成单词表
	wordList, err := word.SqlCreateWordList(30)
	if err != nil {
		word.Logger.Fatalln(err)
		return
	}

	_, err = os.Stat(dailyDir)
	if err != nil {
		err := os.Mkdir(dailyDir, os.ModeDir|0777)
		if err != nil {
			word.Logger.Fatal(err, dailyDir)
		}
	}

	filePath := path.Join(dailyDir, time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		word.Logger.Fatalln(err)
	}

	for _, word := range wordList {
		word.FormatWordList(file)
	}

	// 写入文件

}

func main() {
	runDailyWork()
	dailyFunc()
}