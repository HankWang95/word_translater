package main

import (
	"bufio"
	"github.com/HankWang95/word_translator/word"
	"os"
	"strings"
)

func main() {
	// 多模块handler注册写在loader 中
	handler := word.NewWordHandler()
	handler.DemandWriter(os.Stdout)

	flagDict := handler.RegisterFlag()
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
			*flagDict["w"] <- splitData[0]
		}
	}
}
