package main

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestNewA(t *testing.T) {

	scout,Paths,err := New("./test",time.Second*3)
	if err != nil {
		log.Fatalln(err)
	}

	for _, path := range Paths {
		fmt.Println("Paths：",path.Name)
	}

	err = scout.Scout(func(changePath *[]ScoutChange) {
		for _, change := range *changePath {
			fmt.Printf("type:%v path:%v \n",change.Type,change.Path)
		}
	})

	fmt.Println(err)
}


