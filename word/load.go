package word

import (
	"errors"
	"log"
)

type Loader interface {
	LoadingFlag() (flagDict map[string]*chan string)
}

func NewLoaders() (flagDict map[string]*chan string) {
	loaders := make([]Loader, 0)
	loaders = append(loaders, NewWordLoader())
	// todo 加入其他 loader
	flagDict = loading(loaders)
	return
}

func loading(loaders []Loader) (flagDict map[string]*chan string) {
	flagDictList := make([]map[string]*chan string, 0)
	for _, loader := range loaders {
		flagDictList = append(flagDictList, loader.LoadingFlag())
	}

	flagDict, err := mergeMap(flagDictList)
	if err != nil {
		log.Fatal(err)
	}
	return flagDict
}

func mergeMap(maps []map[string]*chan string) (mergedMap map[string]*chan string, err error) {
	var flagDict = make(map[string]*chan string)
	flagDict = maps[0]
	for i := 1; i < len(maps); i++ {
		for key, value := range maps[i] {
			if _, ok := flagDict[key]; ok {
				return nil, errors.New("flag key 冲突")
			} else {
				flagDict[key] = value
			}
		}
	}
	return flagDict, nil
}
