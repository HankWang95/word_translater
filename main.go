package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	flagDict := NewLoaders()
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