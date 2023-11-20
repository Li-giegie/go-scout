package go_scout

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"
)

type Handler struct {
	name string
}

func (h *Handler) Root() string {
	return h.name
}

func (h *Handler) StartEvent(info []*FileInfo) {
	log.Println("start success", len(info))
	for _, fileInfo := range info {
		fmt.Println(fileInfo.String())
	}
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

func (h *Handler) ErrorEvent(err error) (isContinue bool) {
	fmt.Println("ErrorEvent ", err)
	if errors.Is(err, os.ErrPermission) {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return true
	} else if errors.Is(err, filepath.SkipDir) {
		return true
	}
	return false
}

const testDir = "D:/"

func TestNewScout(t *testing.T) {
	s, err := NewScout(&Handler{name: "./"},
		WithSleep(DEFAULT_SLEEP),
		WithEnableHashCheck(true),
		WithGoroutineNum(DEFAULT_GOROUTINENUM),
		WithFilterFunc(func(path string, info os.FileInfo) bool {
			if len(info.Name()) > 0 && info.Name()[0] == '.' {
				return false
			}
			return true
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
	f, err := getFiles("d:/", nil, func(err error) bool {
		fmt.Println(err)
		return true
	})
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(len(f))
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
