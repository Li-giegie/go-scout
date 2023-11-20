package main

import (
	"fmt"
	go_scout "github.com/Li-giegie/go-scout"
	"log"
	"os"
)

//type ScoutI interface {
// Root() string                              //目录路径
// StartEvent(info []*FileInfo)               //首次启动后触发检测目录的文件信息
// CreateEvent(info []*FileInfo)              //检测到创建行为触发
// ChangeEvent(info []*FileInfo)              //检测到有变更行为触发
// RemoveEvent(info []*FileInfo)              //检测到有删除行为触发
// ErrorCallBack(err error) (isContinue bool) //isContinue：是否继续，true继续，false遇到错误终止
//}

type ScoutImpl struct {
	root string
}

func (m *ScoutImpl) Root() string {
	return m.root
}

func (*ScoutImpl) StartEvent(info []*go_scout.FileInfo) {
	fmt.Println("start event")
	for _, fileInfo := range info {
		fmt.Println(fileInfo.String())
	}
}
func (*ScoutImpl) CreateEvent(info []*go_scout.FileInfo) {
	for _, fileInfo := range info {
		fmt.Println("create event: ", fileInfo.String())
	}
}
func (*ScoutImpl) ChangeEvent(info []*go_scout.FileInfo) {
	for _, fileInfo := range info {
		fmt.Println("change event: ", fileInfo.String())
	}
}
func (*ScoutImpl) RemoveEvent(info []*go_scout.FileInfo) {
	for _, fileInfo := range info {
		fmt.Println("remove event: ", fileInfo.String())
	}
}

func (m *ScoutImpl) ErrorEvent(err error) bool {
	fmt.Println("err: ", err)
	return true
}

func main() {
	_myScout := &ScoutImpl{root: "./"}
	sc, err := go_scout.NewScout(_myScout,
		//是否启用hash检查文件变化
		go_scout.WithEnableHashCheck(false),
		//设置运行过程中启用的协程数量
		go_scout.WithGoroutineNum(go_scout.DEFAULT_GOROUTINENUM),
		//检测休眠时间，默认值1s
		go_scout.WithSleep(go_scout.DEFAULT_SLEEP),
		//回调过滤函数，返回值为false时会被过滤掉，演示过滤掉“.”开头的隐藏文件
		go_scout.WithFilterFunc(func(path string, info os.FileInfo) bool {
			if len(info.Name()) > 0 && info.Name()[0] == '.' {
				return false
			}
			return true
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
	if err = sc.Start(); err != nil {
		log.Fatalln(err)
	}
}
