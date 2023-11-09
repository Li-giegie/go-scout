package go_scout

import (
	"fmt"
	"testing"
	"time"
)

func TestNewScout(t *testing.T) {
	scout, err := NewScout("./", time.Millisecond*100)
	if err != nil {
		t.Error(err)
		return
	}
	err = scout.Start(func(info []*FileInfo) {
		for _, fileInfo := range info {
			fmt.Println(fileInfo.String())
		}
	})
	//stop scout.Stop()
	if err != nil {
		t.Error(err)
	}
}
