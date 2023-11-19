package go_scout

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

type Handler struct {
}

func (h *Handler) StartEvent(info []*FileInfo) {
	log.Println("start success", len(info))
	//for _, fileInfo := range info {
	//	fmt.Println(fileInfo.String())
	//}
}

func (h *Handler) CreateEvent(info []*FileInfo) {
	for _, fileInfo := range info {
		fmt.Println("CreateEvent ", fileInfo.String())
	}

}
func (h *Handler) ChangeEvent(info []*FileInfo) {
	for _, fileInfo := range info {
		fmt.Println("ChangeEvent ", fileInfo.String())
	}
}
func (h *Handler) RemoveEvent(info []*FileInfo) {
	for _, fileInfo := range info {
		fmt.Println("RemoveEvent ", fileInfo.String())
	}
}

const testDir = "D:/"

func TestNewScout(t *testing.T) {
	s, err := NewScout(testDir, &Handler{},
		WithSleep(time.Millisecond*10),
		WithEnableHashCheck(true),
		WithFilterFunc(func(path string, info os.FileInfo) bool {
			if len(info.Name()) > 0 && info.Name()[0] == '.' {
				return false
			} else if len(path) > 0 && path[0] == '.' {
				return false
			} else {
				return true
			}
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	err = s.Start()
	//stop Stop()
	if err != nil {
		log.Fatalln(err)
	}
}

func TestGetFiles(t *testing.T) {
	f, err := GetFiles("d:/test", nil)
	if err != nil {
		t.Error(err)
		return
	}
	//gm := goruntine_manager.NewGoroutineManger(12)
	//gm.Start()
	//calculateHash(nil, f, gm.Run)
	//for _, info := range f {
	//	fmt.Println(info.String())
	//}
	fmt.Println(len(f))
}

func TestGetFilesV2(t *testing.T) {
	a, err := GetFilesV2("d:/test", nil)
	if err != nil {
		t.Error(err)
		return
	}
	//for s, info := range a {
	//	fmt.Println(s, info.String())
	//}
	fmt.Println(len(a))
}

func TestCreateFile(t *testing.T) {
	f, err := os.Create("./a.txt")
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	if _, err = f.Write([]byte("bb")); err != nil {
		t.Error(err)
		return
	}

}
