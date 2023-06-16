package main

import (
	"fmt"
	gf "github.com/Li-giegie/go-utils"
	"log"
	"sync"
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
	lock sync.Mutex
	wg sync.WaitGroup
}

type ScoutChange struct {
	// 改变的路径
	Path string
	// 改变的类型 增删改
	Type ChangeType

	*gf.FileInfo
}



// sleepTime /ms 每一次侦察后休眠时长 理想值 1000
//_path dirs or files	侦察的文件或目录
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
	log.Println("间隔侦查时间：",s.SleepTime)
	return &s,fsi,nil
}

// running Scout 开始侦察文件变化 入参是一个回调方法 当侦擦到变化时调用回调函数
func (s *Scout) Scout(changeFunc func(changePath []*ScoutChange)) error {
	var tmpSC []*ScoutChange

	for  {
		tmpSC = make([]*ScoutChange, 0)
		time.Sleep(s.SleepTime)

		files,err := gf.GetDirInfo(s.Path)
		if err != nil {
			return err
		}

		//删除事件
		for _, s2 := range findOldNotExist(s.filePaths, files) {
			tmpSC = append(tmpSC, &ScoutChange{
				Path: s2,
				Type: ChangeType_Del,
				FileInfo:nil,
			})
			delete(s.filePaths,s2)
		}

		for i:=0;i< len(files);i++{
			v,ok := s.filePaths[files[i].Name]
			//新建文件、文件夹事件
			if !ok {
				tmpSC = append(tmpSC, &ScoutChange{
					Path: files[i].Name,
					Type: ChangeType_Create,
					FileInfo:files[i],
				})
				s.filePaths[files[i].Name]=files[i].ModTime.UnixNano()
				continue
			}
			//文件修改事件
			if v != files[i].ModTime.UnixNano() {
				tmpSC = append(tmpSC, &ScoutChange{
					Path: files[i].Name,
					Type: ChangeType_Update,
					FileInfo:files[i],
				})
				s.filePaths[files[i].Name]=files[i].ModTime.UnixNano()
				continue
			}
		}

		if len(tmpSC) < 1 {
			continue
		}

		//总回调
		changeFunc(tmpSC)

	}


}


// Deprecated: 废弃
func (s *Scout) isRepetition (_new []*gf.FileInfo) []*ScoutChange {
	var fi = make([]*ScoutChange,0)
	var isDel bool
	var info *gf.FileInfo
	for k, _ := range s.filePaths {
		isDel = true
		for i:=0;i< len(_new);i++{
			if _new[i].Name == k {
				isDel = false
				break
			}
		}
		if isDel {
			delete(s.filePaths,k)
			fi = append(fi, &ScoutChange{
				Path: k,
				Type: ChangeType_Del,
				FileInfo:info,
			})
		}
	}
	return fi
}



// Deprecated: 废弃
func (s *Scout) isRepetitionV2 (_new []*gf.FileInfo) []*ScoutChange {
	var w sync.WaitGroup

	var resultSc = make([]*ScoutChange,0)
	var i uint
	var oldFileNames = make([]string,0)
	for k, _ := range s.filePaths {
		oldFileNames = append(oldFileNames, k)
		i++
		if i % 1000 == 0 {
			w.Add(1)
			go func() {
				s.lock.Lock()
				resultSc = append(resultSc, count(oldFileNames,_new)...)
				s.lock.Unlock()
				w.Done()
			}()
			oldFileNames = make([]string,0)
		}

	}
	w.Wait()

	s.lock.Lock()
	for _, change := range resultSc {
		delete(s.filePaths,change.Name)
		fmt.Println("查处删除项：",change.Name)
	}
	s.lock.Unlock()
	return resultSc
}

func findOldNotExist(_old map[string]int64,_new []*gf.FileInfo) []string  {
	var o = make([]string,0)
	var n = make([]string,0)
	for s, _ := range _old {
		o = append(o, s)
	}
	for _, info := range _new {
		n = append(n, info.Name)
	}

	return _findNotExist(&o,&n)
}

func _findNotExist(_old *[]string,_new *[]string) []string {
	existMap := make(map[string]struct{})
	for _, x := range *_new {
		existMap[x] = struct{}{}
	}
	var notExistList []string
	for _, x := range *_old {
		if _, ok := existMap[x]; !ok {
			notExistList = append(notExistList, x)
		}
	}
	return notExistList
}

func count(_old []string,_new []*gf.FileInfo) []*ScoutChange  {
	var isDel bool
	var info *gf.FileInfo
	var resultSc = make([]*ScoutChange,0)

	for _, name := range _old {
		isDel = true
		for _, info = range _new {
			if name == info.Name {
				isDel = false
				break
			}
		}
		if isDel {
			resultSc = append(resultSc, &ScoutChange{
				Path: name,
				Type: ChangeType_Del,
				FileInfo:info,
			})
		}
	}
	return resultSc
}

func (s *Scout) readFilePath(k string) (int64,bool) {
	s.lock.Lock()
	v,ok := s.filePaths[k]
	s.lock.Unlock()
	return v,ok
}

func (s *Scout) writeFilePath(k string,v int64)  {
	s.lock.Lock()
	s.filePaths[k] = v
	s.lock.Unlock()
}