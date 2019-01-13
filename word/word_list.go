package word

import (
	"fmt"
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


// 生成单词列表
func pushWordList(n int) {
	wordList, err := SqlCreateWordList(n)
	if err != nil {
		return
	}
	for n, word := range wordList {
		fmt.Fprint(writer, n+1, " ")
		word.FormatWordList(writer)
		fmt.Fprintln(writer, "--------------------------------------")
	}
}

func SqlCreateWordList(n int) (wordList []*word, err error) {
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
