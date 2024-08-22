package go_scout

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWalkFile(t *testing.T) {

	for {
		filepath.Walk("./d", func(path string, info fs.FileInfo, err error) error {
			fmt.Println(path, info.ModTime())
			return nil
		})
		time.Sleep(time.Second)
	}

	return
	s, err := NewScout(&Config{
		Paths:     []string{"D:\\"},
		Sleep:     time.Millisecond * 50,
		EnableHex: false,
		EventNum:  1024,
	})
	if err != nil {
		t.Error(err)
		return
	}
	s.WalkErrFunc = func(path string, info os.FileInfo, err error) error {
		//log.Println(path, err)
		return nil
	}
	s.FilterFunc = func(path string, info os.FileInfo) bool {
		if path[0] == '.' {
			return false
		}
		return true
	}
	var d1, d2 time.Time
	s.BeforeWalkPathFunc = func() {
		d1 = time.Now()
	}
	s.AfterWalkPathFunc = func(info *map[string]*FileInfo) {
		fmt.Println("搜索耗时", time.Since(d1), len(*info))
	}
	s.BeforeCalculateFunc = func() {
		d2 = time.Now()
	}
	s.AfterCalculateFunc = func() {
		d3 := time.Now()
		fmt.Println("计算耗时", d3.Sub(d2))
		fmt.Println("总计耗时", d3.Sub(d1))
	}
	for {
		s.Restart()
		for value := range s.EventChan {
			//log.Println(value.Type, value.FileInfo)
			_ = value
		}
		time.Sleep(time.Second * 3)
		//ssadasdasdasdasdasdasdasasdas
	}

}
