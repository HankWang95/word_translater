package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/robfig/cron"
	"io"
	"os"
	"path"
	"time"
)

var (
	// todo get path from config
	projectDir = path.Join(os.Getenv("HOME"), "Documents", "Kanna")
	dailyDir   = path.Join(projectDir, "daily")
)

func runDailyWork() {
	c := cron.New()
	c.AddFunc("0 0 19 * * *", dailyFunc)
	c.Start()
}

func dailyFunc() {
	// 生成单词表
	wordList, err := SqlCreateWordList(30)
	if err != nil {
		Logger.Fatalln(err)
		return
	}

	_, err = os.Stat(dailyDir)
	if err != nil {
		err := os.Mkdir(dailyDir, os.ModeDir|0777)
		if err != nil {
			Logger.Fatal(err, dailyDir)
		}
	}

	filePath := path.Join(dailyDir, time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Logger.Fatalln(err)
	}

	for _, word := range wordList {
		word.FormatWordList(file)
	}

}

// ------------------------------------- db ------------------------------------------

type wordStruct struct {
	Id           int64      `json:"id"                sql:"id"`
	Word         string     `json:"word"              sql:"word"`
	Translations string     `json:"translations"      sql:"translations"`
	CreatedOn    *time.Time `json:"created_on"        sql:"created_on"`
	AppearTime   int        `json:"appear_time"       sql:"appear_time"`
	LastAppear   *time.Time `json:"last_appear"       sql:"last_appear"`
}

func (w *wordStruct) FormatWordList(writer io.Writer) {
	var fi interface{}
	json.Unmarshal([]byte(w.Translations), &fi)
	f := fi.(map[string]interface{})
	fmt.Fprint(writer, "---- ", w.Word, " -- ")
	if v, ok := f["translation"]; ok {
		fmt.Fprintln(writer, v.([]interface{})[0], "----")
	}
	if v, ok := f["basic"]; ok {
		basic := v.(map[string]interface{})
		if v, ok := basic["explains"]; ok {
			fmt.Fprintln(writer, "其他释义: ", v.([]interface{}))
		}
	}
}


func SqlCreateWordList(n int) (wordList []*wordStruct, err error) {
	wordList = make([]*wordStruct, 0, n)
	db := getMySQLSession()
	stmt, err := db.Prepare(`SELECT id, word, translations, appear_time, last_appear 
				FROM notebook_word 
				ORDER BY last_appear DESC 
				LIMIT ?`)
	rows, err := stmt.Query(n)
	for rows.Next() {
		var w = new(wordStruct)
		err = rows.Scan(&w.Id, &w.Word, &w.Translations, &w.AppearTime, &w.LastAppear)
		if err != nil {
			Logger.Println(err)
			continue
		}
		wordList = append(wordList, w)
	}
	if rows.Err() != nil {
		Logger.Println(rows.Err())
	}
	rows.Close()
	return wordList, nil
}
