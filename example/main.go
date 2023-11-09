package main

import (
	"fmt"
	scout "github.com/Li-giegie/go-scout"
	"time"
)

func main() {
	s, err := scout.NewScout("../example", time.Millisecond*100)
	if err != nil {
		panic(err)
	}
	err = s.Start(func(info []*scout.FileInfo) {
		for _, fileInfo := range info {
			switch fileInfo.ChangeType {
			case scout.ChangeType_Create:
				//todo
			case scout.ChangeType_Del:
				//todo
			case scout.ChangeType_Update:
				//todo
			}
			fmt.Println(fileInfo.String())
		}
	})
}
