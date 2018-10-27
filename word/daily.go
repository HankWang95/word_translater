package word

import (
	"fmt"
	"github.com/robfig/cron"
	"strconv"
)

// ---------------------- service --------------------------

func wordListEnter(sn string) {
	n, err := strconv.Atoi(sn)
	if err != nil {
		n = 5
	}
	pushWordList(n)
}

func (this *wordHandler) runDailyWork() {
	c := cron.New()
	c.AddFunc("0 0 19 * * *", dailyFunc)
	c.Start()
}

func dailyFunc() { pushWordList(10) }

// 生成单词列表
func pushWordList(n int) {
	wordList, err := sqlCreateWordList(n)
	if err != nil {
		return
	}
	for n, word := range wordList {
		fmt.Fprint(writer, n+1, " ")
		word.FormatWordList()
		fmt.Fprintln(writer, "--------------------------------------")
	}
}

func sqlCreateWordList(n int) (wordList []*word, err error) {
	wordList = make([]*word, 0, n)
	db := getMySQLSession()
	stmt, err := db.Prepare(`SELECT id, word, translations, appear_time, last_appear 
				FROM notebook_word 
				ORDER BY last_appear DESC 
				LIMIT ?`)
	rows, err := stmt.Query(n)
	for rows.Next() {
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
