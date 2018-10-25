package notebook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/HankWang95/Kanna/server"
	_ "github.com/go-sql-driver/mysql"
	"github.com/smartwalle/ini4go"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
	"path"
)

var (
	httpClient          *http.Client
	logger              *log.Logger
	key                 string
	MaxIdleConns        = 100
	MaxIdleConnsPerHost = 100
	projectPath = path.Join(os.Getenv("HOME"), "Documents", "Kanna")
	configPath = path.Join(projectPath, "config")
	speechPath = path.Join(projectPath, "speech")
)

func init() {
	_, err := os.Stat(projectPath)
	if err != nil {
		err := os.Mkdir(projectPath, os.ModeDir | 0777)
		if err != nil {
			logger.Fatal(err)
		}
		_, err = os.Stat(speechPath)
		if err != nil {
			os.Mkdir(speechPath, os.ModeDir | 0777)
		}
	}
	var config = ini4go.New(false)
	config.SetUniqueOption(true)
	config.Load(configPath)
	httpClient = createHTTPClient()
	logFile, err := os.OpenFile(path.Join(projectPath, "kanna.log"), os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	key = config.GetValue("youdao", "key")
	if err != nil {
		logger.Fatal(err)
	}
	logger = log.New(logFile, "[kanna-notebook] ", log.Ltime|log.Ldate|log.Lshortfile)
}

// ———————————————————————————————————————————— Loader -----------------------------------------------------

var queryWordChan = make(chan string, 10)
var wordListChan = make(chan string, 10)


type wordLoader struct{}

func queryWordEnter(word string) {
	w, err := queryWord(word)
	if err != nil {
		logger.Println(err)
	}else{
		w.FormatTranslations()
	}
}

func NewWordLoader() *wordLoader {
	return new(wordLoader)
}

func (this *wordLoader) LoadingFlag() (flagDict map[string]*chan string) {
	flagDict = make(map[string]*chan string)
	flagDict["w"] = &queryWordChan
	flagDict["wl"] = &wordListChan
	go flagHandler()
	go dailyWordList()
	return
}

func flagHandler() {
	for {
		select {
		case word := <-queryWordChan:
			queryWordEnter(word)
		case n:=<-wordListChan:
			wordListEnter(n)
		}
	}
}

// ———————————————————————————————————————————— Service -----------------------------------------------------

type word struct {
	Id           int64      `json:"id"                sql:"id"`
	Word         string     `json:"word"              sql:"word"`
	Translations string     `json:"translations"      sql:"translations"`
	CreatedOn    *time.Time `json:"created_on"        sql:"created_on"`
	AppearTime   int        `json:"appear_time"       sql:"appear_time"`
	LastAppear   *time.Time `json:"last_appear"       sql:"last_appear"`
}

func (w *word) FormatTranslations() {
	var fi interface{}
	json.Unmarshal([]byte(w.Translations), &fi)
	f := fi.(map[string]interface{})
	fmt.Println("---- ", w.Word, " ----")
	if v, ok := f["translation"]; ok {
		fmt.Println("基本翻译: ", v.([]interface{}))
	}
	if v, ok := f["basic"]; ok {
		basic := v.(map[string]interface{})
		if v, ok := basic["us-phonetic"]; ok {
			fmt.Println("美式发音: ", v.(string))
		}
		if v, ok := basic["uk-phonetic"]; ok {
			fmt.Println("英式发音: ", v.(string))
		}
		if v, ok := basic["explains"]; ok {
			fmt.Println("其他释义: ", v.([]interface{}))
		}
		if v, ok := basic["us-speech"]; ok {
			f, err := os.Open(fmt.Sprint(speechPath + "/", w.Word, ".mp3"))
			if err == nil {
				playMP3(f.Name())
				return
			}
			f.Close()
			usURL := fmt.Sprint(v.(string))
			go downloadMP3(w.Word, usURL)
		}
	}
}

func (w *word) FormatWordList() {
	var fi interface{}
	json.Unmarshal([]byte(w.Translations), &fi)
	f := fi.(map[string]interface{})
	fmt.Print("---- ",w.Word, " -- ")
	if v, ok := f["translation"]; ok {
		fmt.Println(v.([]interface{})[0], "----")
	}
	if v, ok := f["basic"]; ok {
		basic := v.(map[string]interface{})
		if v, ok := basic["explains"]; ok {
			fmt.Println("其他释义: ", v.([]interface{}))
		}
	}
}

func createHTTPClient() *http.Client {
	client := &http.Client{
		Timeout: time.Second * 3,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        MaxIdleConns,
			MaxIdleConnsPerHost: MaxIdleConnsPerHost,
			IdleConnTimeout:     time.Second * 90,
		},
	}
	return client
}

func youDaoTranslate(searchWord string) (result *word, err error) {
	requestUrl := fmt.Sprintf("http://fanyi.youdao.com/openapi.do?keyfrom=YouDaoCV&key=%s&type=data&doctype=json&version=1.2&q=%s", key, searchWord)
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)


	resp, err := httpClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		logger.Println("获取api出错", err)
	}

	if resp.Status != "200 OK" {
		logger.Println("调用翻译api发生错误")
		return nil, err
	}

	var wordStruct = new(word)
	wordStruct.Word = searchWord
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	wordStruct.Translations = buf.String()

	return wordStruct, nil
}

func downloadMP3(name, url string) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Println(" download MP3 err :", err)
		return
	}

	// 查看是否有存放语音的文件夹


	f, err := os.OpenFile(fmt.Sprint(speechPath, "/", name, ".mp3"), os.O_RDWR | os.O_CREATE | os.O_TRUNC, 0666)
	defer f.Close()
	if err != nil {
		logger.Fatal(err)
		return
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Println(" download MP3 err :", err)
		return
	}
	io.Copy(f, resp.Body)
	resp.Body.Close()
	playMP3(f.Name())
}

func playMP3(path string) {
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("afplay", path)
		//cmd.Start()
		// todo diff
		cmd.Run()
	}
}

func queryWord(searchWord string) (result *word, err error) {
	word, sqlErr := sqlGetWord(searchWord)
	if sqlErr == nil {
		go sqlUpdateWord(word.Id)
		return word, nil
	}

	result, err = youDaoTranslate(searchWord)
	if err != nil {
		return nil, err
	}
	err = sqlAddWord(result)
	if err != nil {
		logger.Println(err)
	}
	return result, nil
}

// ———————————————————————————————————————————— MySQL -----------------------------------------------------

func sqlAddWord(word *word) (err error) {
	db := server.GetMySQLSession()
	stmt, err := db.Prepare(`INSERT INTO notebook_word (word, translations, created_on, appear_time, last_appear) 
				VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		logger.Println(err)
		return err
	}
	now := time.Now()
	_, err = stmt.Exec(word.Word, word.Translations, now, 1, now)
	if err != nil {
		logger.Println(err)
		return err
	}
	return nil
}

func sqlGetWord(qWord string) (result *word, err error) {
	result = new(word)
	db := server.GetMySQLSession()
	stmt, err := db.Prepare(`SELECT id, word, translations 
							FROM notebook_word WHERE word = ?`)
	row := stmt.QueryRow(qWord)
	err = row.Scan(&result.Id, &result.Word, &result.Translations)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func sqlUpdateWord(id int64) (err error) {
	db := server.GetMySQLSession()
	stmt, err := db.Prepare(`UPDATE notebook_word
				SET appear_time = appear_time + 1, 
				last_appear = ?
				WHERE id = ?`)
	_, err = stmt.Exec(time.Now(), id)
	if err != nil {
		logger.Println(err)
		return err
	}
	return
}
