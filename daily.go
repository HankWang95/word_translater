package notebook

import (
	"github.com/HankWang95/Kanna/server"
	"fmt"
	"strconv"
	"github.com/robfig/cron"
)

// ---------------------- service --------------------------

func wordListEnter(sn string) {
	n,err := strconv.Atoi(sn)
	if err != nil {
		n = 5
	}
	pushWordList(n)
}

func dailyWordList() {
	c := cron.New()
	c.AddFunc("0 0 19 * * *", dailyFunc)
	c.Start()
}

func dailyFunc() {	pushWordList(10) }

// 生成单词列表
func pushWordList(n int) {
	wordList, err := sqlCreateWordList(n)
	if err != nil {
		return
	}
	for n, word := range wordList{
		fmt.Print(n+1, " ")
		word.FormatWordList()
		fmt.Println("--------------------------------------")
	}
}

func sqlCreateWordList(n int) (wordList []*word, err error) {
	wordList = make([]*word, 0, n)
	db := server.GetMySQLSession()
	stmt, err := db.Prepare(`SELECT id, word, translations, appear_time, last_appear 
				FROM notebook_word 
				ORDER BY last_appear DESC 
				LIMIT ?`)
	rows,err := stmt.Query(n)
	for rows.Next(){
		var w = new(word)
		err = rows.Scan(&w.Id, &w.Word, &w.Translations, &w.AppearTime, &w.LastAppear)
		if err != nil {
			logger.Println(err)
			continue
		}
		wordList = append(wordList, w)
	}
	if rows.Err() != nil {
		logger.Println(rows.Err())
	}
	rows.Close()
	return wordList, nil
}
