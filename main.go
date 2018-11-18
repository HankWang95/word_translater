package main

import (
	"bufio"
	"github.com/HankWang95/word_translator/word"
	"github.com/smartwalle/ini4go"
	"log"
	"net"
	"os"
	"path"
	"strings"
)

var flagDict map[string]*chan string

func main() {
	// 多模块handler注册写在loader 中
	handler := word.NewWordHandler()
	handler.DemandWriter(os.Stdout)

	flagDict = handler.RegisterFlag()
	//go scan(flagDict)
	select {}
}

func scan(input net.Conn) {
	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		data := scanner.Text()
		splitData := strings.Split(data, " ")
		if flagChan, ok := flagDict[splitData[0]]; ok {
			*flagChan <- strings.Join(splitData[1:], "")
		} else {
			*flagDict["w"] <- splitData[0]
		}
	}
}

func listener() {
	var config = ini4go.New(false)
	config.SetUniqueOption(true)
	config.Load(path.Join(path.Join(os.Getenv("HOME"), "MyProject", "Kanna"), "config"))
	l, err := net.Listen("tcp", config.GetValue("server", "address"))
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			continue
		}
		go scan(conn)

	}

}
