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
	"time"
)

func main() {
	s, err := scout.NewScout("../example", time.Millisecond*100)
	if err != nil {
		panic(err)
	}
	err = s.Start(func(info []*scout.FileInfo) {
		for _, fileInfo := range info {
			switch fileInfo.ChangeType {
			case scout.ChangeType_Create:
				//todo
			case scout.ChangeType_Del:
				//todo
			case scout.ChangeType_Update:
				//todo
			}
			fmt.Println(fileInfo.String())
		}
	})
}

```

