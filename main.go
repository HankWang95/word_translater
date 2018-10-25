package main

import (
	"bufio"
	"fmt"
	"github.com/HankWang95/Kanna/load"
	"github.com/HankWang95/Kanna/server"
	"log"
	"os"
	"strings"
)

func main() {
	server.InitMySQL()
	flagDict := load.NewLoaders()
	go scan(flagDict)
	select {}
}

func scan(flagDict map[string]*chan string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		data := scanner.Text()
		splitData := strings.Split(data, " ")
		if flagChan, ok := flagDict[splitData[0]]; ok {
			*flagChan <- strings.Join(splitData[1:], "")
		} else {
			fmt.Println("flag not exist, read help")
			log.SetFlags(log.Ldate)
		}
	}
}

// todo 非对称加密
//func chackKeyForHeaven()  {
//}
