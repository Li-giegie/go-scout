## [go-scout是一个侦察文件、目录发生变化的侦察服务](#)
![golang](https://img.shields.io/badge/golang-v1.19-blue)
![simple](https://img.shields.io/badge/simple-extend-green)

### 获取
```
go get -u github.com/Li-giegie/go-scout
```

### 使用

```go
package main

import (
	"fmt"
	scout "github.com/Li-giegie/go-scout"
	"log"
	"time"
)

func main() {
	//创建侦查对象
	s, err := scout.NewScout("./",
		//设置休眠时间
		scout.WithScoutSleep(time.Millisecond*1000),
		//设置是否开启MD5检测文件是否变更
		scout.WithScoutEnableHashCheck(true),
		//设置过滤掉那些文件
		scout.WithScoutFilterFunc(func(name, fullPath string) bool {
			//过滤掉"."开头的任何文件, .git
			if fullPath[0] == '.' {
				return false
			}
			return true
		}),
	)
	if err != nil {
		log.Fatalln(err)
	}
    
	//启动 入参侦查的文件、目录发生变动了回调入参的方法，
	err = s.Start(func(info []*scout.FileInfo) {
		for _, fileInfo := range info {

			switch fileInfo.ChangeType {
			case scout.ChangeType_Create:
			//todo:
			case scout.ChangeType_Update:
			//todo:
			case scout.ChangeType_Del:
				//todo:
			}
			fmt.Println(fileInfo.String())
		}
	})
	//stop scout.Stop() 关闭
	if err != nil {
		log.Fatalln(err)
	}
}

```