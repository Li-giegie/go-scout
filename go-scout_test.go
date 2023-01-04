package go_scout

import (
	"fmt"
	"log"
	"testing"
)

func TestNew(t *testing.T) {

	scout,Paths,err := New(1000,"./")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Paths：",Paths)

	scout.SetDebug()
	err = scout.Scout(func(changePath *[]ScoutChange) {
		for _, change := range *changePath {
			fmt.Printf("type:%v path:%v \n",change.Type,change.Path)
		}
	})

	fmt.Println(err)
}


