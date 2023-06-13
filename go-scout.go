package main

import (
	gf "github.com/Li-giegie/go-utils"
	"time"
)

type ChangeType byte

const (
	ChangeType_Create ChangeType = 1
	ChangeType_Del ChangeType = 2
	ChangeType_Update ChangeType = 3
)

type Scout struct {
	filePaths map[string] int64
	// 休眠时长
	SleepTime time.Duration
	// 侦察变化的路径
	Path string

}

type ScoutChange struct {
	// 改变的路径
	Path string
	// 改变的类型 增删改
	Type ChangeType

	*gf.FileInfo
}

// sleepTime /ms 每一次侦察后休眠时长 理想值 1000
//_path dirs or files	侦察的文件或目录可配置多个
// return Scout *Scout filePaths []string err error
func New(dirPath string,sleepTime time.Duration) (*Scout,[]*gf.FileInfo,error) {

	fsi,err := gf.GetDirInfo(dirPath)
	if err != nil {
		return nil,nil,err
	}
	var s Scout
	s.filePaths = make(map[string]int64)
	s.Path =  dirPath
	s.SleepTime = sleepTime

	for _, file_ := range fsi {
		s.filePaths[file_.Name] = file_.ModTime.UnixNano()
	}
	return &s,fsi,nil
}

// running Scout 开始侦察文件变化 入参是一个回调方法 当侦擦到变化时调用回调函数
func (s *Scout) Scout(changeFunc func(changePath *[]ScoutChange)) error {
	var cp []ScoutChange
	for  {
		time.Sleep(s.SleepTime)
		files,err := gf.GetDirInfo(s.Path)
		if err != nil {
			return err
		}
		cp = make([]ScoutChange, 0)

		for _, file_ := range files {
			v,ok := s.filePaths[file_.Name]

			//新建文件、文件夹事件
			if !ok {
				cp = append(cp, ScoutChange{
					Path: file_.Name,
					Type: ChangeType_Create,
					FileInfo:file_,
				})
				s.filePaths[file_.Name] = file_.ModTime.UnixNano()
				continue
			}

			//文件修改事件
			if v != file_.ModTime.UnixNano() {
				//if s.Debug == "enable" && s.RunMode == RunnMode_AllChange{ log.Println("RunMode AllChange Update") }
				cp = append(cp, ScoutChange{
					Path: file_.Name,
					Type: ChangeType_Update,
					FileInfo:file_,
				})
				s.filePaths[file_.Name] = file_.ModTime.UnixNano()
				continue
			}

		}

		//删除事件
		delPath := s.isRepetition(files)
		cp = append(cp, delPath...)
		if len(cp) < 1 {
			continue
		}
		//总回调
		changeFunc(&cp)
	}


}


// 是否重复
func (s *Scout) isRepetition (_new []*gf.FileInfo) []ScoutChange {
	var fi = make([]ScoutChange,0)
	var isDel bool
	var info *gf.FileInfo
	for k, _ := range s.filePaths {
		isDel = true
		for _, info = range _new {
			if info.Name == k {
				isDel = false
			}
			continue
		}
		if isDel {
			delete(s.filePaths,k)
			fi = append(fi, ScoutChange{
				Path: k,
				Type: ChangeType_Del,
				FileInfo:info,
			})
		}
	}
	return fi
}
