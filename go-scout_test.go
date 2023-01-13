package go_scout

import (
	"fmt"
	"log"
	"testing"
)

func TestNew(t *testing.T) {

	scout,fileInfos,err := New(600,"./")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Print("变化项：名称	|	路径|	 时间|	 类型|	 是否文件	\n")
	for _, info := range fileInfos {
		fmt.Println(info.Name,"|",info.Path,"|",info.ModTime.Format("2006-01-02 15:04:05"),"|",ctm[info.Type],"|",info.IsDir)
	}
	fmt.Print("\n---------------------------------------------------------------\n")
	//scout.SetDebug()
	fmt.Print("变化项：名称	|	路径|	 时间|	 类型|	 是否文件	\n")
	err = scout.Scout(func(changePath []*fileInfo) {
		for _, info := range changePath {
			fmt.Println(info.Name,info.Path,info.ModTime.Format("2006-01-02 15:04:05"),ctm[info.Type],info.IsDir)
		}
	})
	fmt.Println(err)
}


