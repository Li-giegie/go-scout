package main

import (
	"fmt"
	scout "github.com/Li-giegie/go-scout"
	"log"
	"time"
)

// 4
func main() {
	s, err := scout.NewScout("./",
		scout.WithScoutSleep(time.Millisecond*1000),
		scout.WithScoutEnableHashCheck(true), scout.WithScoutFilterFunc(func(name, fullPath string) bool {
			fmt.Println(name, "-", fullPath)
			return true
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}

	err = s.Start(func(info []*scout.FileInfo) {
		for _, fileInfo := range info {

			switch fileInfo.ChangeType {
			case scout.ChangeType_Create:
			//todo:
			case scout.ChangeType_Update:
			//todo:
			case scout.ChangeType_Del:
				//todo:
			}
			fmt.Println(fileInfo.String())
		}
	})
	//stop scout.Stop()
	if err != nil {
		log.Fatalln(err)
	}
}
