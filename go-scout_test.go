package go_scout

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	s := NewScout(&Config{
		Paths:         []string{"./"},
		Sleep:         time.Millisecond * 200,
		EnableHex:     false,
		EventChanSize: 1024,
	})
	s.WalkErrFunc = func(path string, info os.FileInfo, err error) error {
		if errors.Is(err, os.ErrPermission) {
			return SkipDir
		}
		return nil
	}
	s.FilterFunc = func(path string, info os.FileInfo) SkipType {
		if info.IsDir() {
			if info.Name() != "." && info.Name()[0] == '.' {
				return SkipType_Dir
			}
			return SkipType_NoSkip
		}
		if info.Name()[0] == '.' {
			return SkipType_File
		}
		return SkipType_NoSkip
	}
	//t1 := time.Now()
	//s.BeforeCalculateFunc = func() {
	//	t1 = time.Now()
	//}
	//s.AfterCalculateFunc = func() {
	//	fmt.Println(time.Since(t1))
	//}
	if err := s.Start(); err != nil {
		t.Error(err)
		return
	}

	for value := range s.EventChan {
		switch value.Type {
		case EventType_Error:
			fmt.Println("错误", value.Type, value.Error)
		default:
			fmt.Println(value.Type, value.FileInfo.String())
		}
	}
	fmt.Println("结束")
}
