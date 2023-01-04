package go_scout

import (
	"bytes"
	"errors"
	_file "github.com/dablelv/go-huge-util/file"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

type ChangeType byte

const (
	ChangeType_Create ChangeType = 1
	ChangeType_Del ChangeType = 2
	ChangeType_Update ChangeType = 3
)

type RunningMode byte

const (
	// 一次可能侦察多个路径 当侦察到一个路径变化后不再侦察剩下来的路径进入休眠 休眠超时后继续侦察
	RunnMode_ChangeOnce RunningMode = 10
	// 侦察所有变化的文件、目录
	RunnMode_AllChange RunningMode = 11
)

type Scout struct {
	filePaths sync.Map
	// 休眠时长
	SleepTime int64
	// 侦察变化的路径
	Path []string
	// 运行侦察的模式
	RunMode RunningMode
	// 调试模式
	Debug string
}

type ScoutChange struct {
	// 改变的路径
	Path string
	// 改变的类型 增删改
	Type ChangeType
}

// sleepTime /ms 每一次侦察后休眠时长 理想值 1000
//_path dirs or files	侦察的文件或目录可配置多个
// return Scout *Scout filePaths []string err error
func New(sleepTime int64,_path ...string) (*Scout,[]string,error) {
	var socut = Scout{
		filePaths: sync.Map{},
		SleepTime: sleepTime,
		Path: _path,
		RunMode: RunnMode_AllChange,
		Debug: "disable",
	}

	files,err := getFilePaths(_path...)
	if err != nil {
		return nil,nil,err
	}
	var mod int64
	for _, file_ := range files {
		mod = getFileMod(file_)
		if mod == -1 {
			return nil,nil,errors.New("get file modTime err -New")
		}
		socut.filePaths.Store(file_,mod)
	}
	return &socut,files,nil
}

// running Scout 开始侦察文件变化 入参是一个回调方法 当侦擦到变化时调用回调函数
func (s *Scout) Scout(changeFunc func(changePath *[]ScoutChange)) error {

	var st = time.Millisecond * time.Duration(s.SleepTime)
	var cp []ScoutChange
	var modTime int64
	var isRunnMode_ChangeOnce_ok bool
	for  {
		time.Sleep(st)
		files,err := getFilePaths(s.Path...)
		if err != nil {
			return err
		}
		cp = make([]ScoutChange, 0)
		isRunnMode_ChangeOnce_ok = false
		for _, file_ := range files {

			modTime = getFileMod(file_)
			if modTime == -1 { return appendError("get file modTime err -Scout",file_) }

			v,ok := s.filePaths.Load(file_)
			if !ok {
				if s.Debug == "enable" && s.RunMode == RunnMode_AllChange{ log.Println("RunMode AllChange Create") }

				cp = append(cp, ScoutChange{
					Path: file_,
					Type: ChangeType_Create,
				})
				s.filePaths.Store(file_,modTime)
				if s.RunMode == RunnMode_ChangeOnce {
					isRunnMode_ChangeOnce_ok = true
					if s.Debug == "enable"{ log.Println("RunMode ChangeOnce create ") }
					break
				}
				continue
			}

			if v != modTime {
				if s.Debug == "enable" && s.RunMode == RunnMode_AllChange{ log.Println("RunMode AllChange Update") }
				cp = append(cp, ScoutChange{
					Path: file_,
					Type: ChangeType_Update,
				})
				s.filePaths.Store(file_,modTime)
				if s.RunMode == RunnMode_ChangeOnce {
					isRunnMode_ChangeOnce_ok = true
					if s.Debug == "enable"{ log.Println("RunMode ChangeOnce Update ") }
					break
				}
				continue
			}

		}

		s.filePaths.Range(func(key, value any) bool {

			if s.RunMode == RunnMode_ChangeOnce && isRunnMode_ChangeOnce_ok {
				if s.Debug == "enable"{
					log.Println("RunMode ChangeOnce delete no Scout")
				}
				return false
			}
			fn := key.(string)
			if !isRepetition(files,fn) {
				cp = append(cp, ScoutChange{
					Path: fn,
					Type: ChangeType_Del,
				})
				s.filePaths.Delete(key)
				if s.Debug == "enable" && s.RunMode == RunnMode_AllChange{
					log.Println("RunMode AllChange delete")
				}
			}
			return true
		})

		changeFunc(&cp)
	}


}

// 设置运行模式
func (s *Scout) SetRunMode(mode RunningMode)  {
	s.RunMode = mode
}

// [enable | disable ] 是否开启调试 入参数为空 执行取反
func (s *Scout) SetDebug(arg ...bool)  {
	if arg == nil {
		if s.Debug == "enable" { s.Debug = "disable"
		}else { s.Debug = "enable" }
		return
	}
	if arg[0] {
		s.Debug = "enable"
		return
	}
	s.Debug = "disable"
}

// 获取路径
func getFilePaths(_path ...string) ([]string,error) {
	var result =make([]string,0)
	for _, s := range _path {
		s = strings.ReplaceAll(strings.ReplaceAll(s,`\`,"/"),"//","/")
		exist,err := _file.IsPathExist(s)
		if !exist || err != nil {
			return nil, appendError("not a path err:",err)
		}

		if !_file.IsDir(s) {
			if isRepetition(result,s) {
				continue
			}
			result = append(result, s)
			continue
		}
		files,err := _file.GetDirAllEntryPaths(s,true)
		if err != nil {
			return nil, appendError("open dir [",s,"] err",err)
		}
		for _, file_ := range files {
			file_ = strings.ReplaceAll(strings.ReplaceAll(file_,`\`,"/"),"//","/")
			if isRepetition(result,file_) {
				continue
			}
			result = append(result, file_)
		}

	}

	return result,nil
}

// 是否重复
func isRepetition(src []string,dist string) bool {
	for _, s := range src {
		if s == dist {
			return true
		}
	}
	return false
}

// 入参 error、string 类型 生成新的 error
func appendError(errs ...interface{}) error {
	var errBuf = new(bytes.Buffer)
	for _, err := range errs {
		switch val := err.(type) {
		case error:
			errBuf.WriteString(val.Error())
		case string:
			errBuf.WriteString(val)
		}
	}

	return errors.New(errBuf.String())
}

// 获取文件、目录变成时间
func getFileMod(_path string) int64 {
	info ,err := os.Stat(_path)
	if err != nil {
		log.Println(err)
		return -1
	}

	return info.ModTime().UnixNano()
}


