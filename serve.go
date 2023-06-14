package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"time"
)

type conf struct {
	SleepTime time.Duration
	Path string
	Api string
}

var _conf conf
func main()  {
	parseFlag()

	scout,Paths,err := New(_conf.Path,_conf.SleepTime)
	if err != nil {
		log.Fatalln(err)
	}
	for _, path := range Paths {
		fmt.Println("管理的目录：",path.Name)
	}
	log.Println("dir:",_conf.Path)
	log.Println("共计管理目录：",len(Paths))
	err = scout.Scout(func(changePath []*ScoutChange) {
		buf,err := json.Marshal(&changePath)
		if err != nil {
			log.Fatalln("main.json.Marshal(&changePath) err: ",err)
		}
		fmt.Println(string(buf))
		_,err = http.Post(_conf.Api,"application/json",bytes.NewReader(buf))
		if err != nil {
			log.Println("http request err: ",err)
		}
	})

	fmt.Println(err)
}

func parseFlag()  {
	confF := flag.String("conf","./conf.yaml","指定配置文件启动")
	createConf := flag.String("createconf","","conf.yaml 创建一个配置文件模板 输入文件名")
	flag.Parse()
	if *createConf != "" {
		buf,err := yaml.Marshal(
			&conf{
				SleepTime: time.Second*3,
				Path:      "./",
				Api: "https://",
			},
		)
		if err != nil {
			fmt.Println("创建失败：",err)
		}else {
			if err = os.WriteFile(*createConf,buf,0666); err != nil {
				fmt.Println("创建失败：",err)
			}
		}
		os.Exit(0)
	}

	buf,err := os.ReadFile(*confF)
	if err != nil {
		fmt.Println("配置文件打开失败：",err)
		os.Exit(1)
	}
	err = yaml.Unmarshal(buf,&_conf)
	if err != nil {
		fmt.Println("配置文件序列化失败：",err)
		os.Exit(1)
	}
}
