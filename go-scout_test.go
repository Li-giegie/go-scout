package go_scout

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestNewScout(t *testing.T) {
	s, err := NewScout("./",
		WithScoutSleep(time.Millisecond*1000),
		WithScoutEnableHashCheck(true), WithScoutFilterFunc(func(name, fullPath string) bool {
			if fullPath[0] == '.' {
				return false
			}
			return true
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	for _, info := range s.FileInfoMap() {
		fmt.Println(info.String())
	}
	err = s.Start(func(info []*FileInfo) {
		for _, fileInfo := range info {

			switch fileInfo.ChangeType {
			case ChangeType_Create:
			//todo:
			case ChangeType_Update:
			//todo:
			case ChangeType_Del:
				//todo:
			}
			fmt.Println(fileInfo.String())
		}
	})
	//stop Stop()
	if err != nil {
		log.Fatalln(err)
	}
}

// a
func TestCalculateMD5(t *testing.T) {
	mdgStr, err := calculateMD5("./go.mod")
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(mdgStr, len(mdgStr))
}
