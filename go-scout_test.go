package main

import (
	"fmt"
	"log"
	"strconv"
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

	err = scout.Scout(func(changePath []*ScoutChange) {
		for _, change := range changePath {
			fmt.Printf("type:%v path:%v \n",change.Type,change.Path)
		}
	})

	fmt.Println(err)
}


func TestRangeMap(t *testing.T){
	var m = make(map[string] int64)
	var val = make([]int64,100000)
	t1:=time.Now()
	for i:=0;i<100000;i++{
		m[strconv.Itoa(i+10000000)] = int64(i)
		val[i] = int64(i)
	}
	fmt.Println(time.Since(t1))
	t1=time.Now()
	c := 0

	fmt.Println(c,time.Since(t1))
}

func TestDeleteCheckt(t *testing.T) {

	s,f,err := New("D:\\_project\\GO Project\\mod",time.Millisecond*10)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("打开数量：",len(f))
	t1 := time.Now()
	//res := s.isRepetition(f)
	res := s.isRepetitionV2(f)
	fmt.Println(time.Since(t1))
	for _, re := range res {
		fmt.Println(*re)
	}




}