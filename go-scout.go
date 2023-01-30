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
	ChangeType_init ChangeType = 0
	ChangeType_Create ChangeType = 1
	ChangeType_Del ChangeType = 2
	ChangeType_Update ChangeType = 3
)

var ctm  map[ChangeType] string

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
	// 是否侦察空目录
	ScoutMnt bool
	// 调试模式
	Debug string
}

//type ScoutChange struct {
//	// 改变的路径
//	Path string
//	// 改变的类型 增删改
//	Type ChangeType
//}

type FileInfo struct {
	Name string
	Size int64
	Mode os.FileMode
	ModTime time.Time
	IsDir bool
	Type ChangeType
	Path string
}

func init()  {
	ctm = map[ChangeType]string{
		ChangeType_Create:"新建",
		ChangeType_Update: "更新",
		ChangeType_Del :"删除",
	}
}

// sleepTime /ms 每一次侦察后休眠时长 理想值 1000
//_path dirs or files	侦察的文件或目录可配置多个
// return Scout *Scout filePaths []string err error
func New(sleepTime int64,_path ...string) (*Scout,[]*FileInfo,error) {
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
	var infos = make([]*FileInfo,0)
	for _, file_ := range files {
		info,err := getFileInfo(file_)
		if err != nil {
			return nil,nil,errors.New("get file modTime err -New")
		}
		info.Path = file_
		socut.filePaths.Store(file_,info.ModTime.UnixNano())
		infos = append(infos,info )
	}
	return &socut,infos,nil
}

// running Scout 开始侦察文件变化 入参是一个回调方法 当侦擦到变化时调用回调函数
func (s *Scout) Scout(changeFunc func(changePath []*FileInfo)) error {

	var st = time.Millisecond * time.Duration(s.SleepTime)
	var cp []*FileInfo
	//var modTime int64
	var isRunnMode_ChangeOnce_ok bool

	for  {
		time.Sleep(st)
		files,err := getFilePaths(s.Path...)
		if err != nil {
			return err
		}
		cp = make([]*FileInfo, 0)
		isRunnMode_ChangeOnce_ok = false
		for _, file_ := range files {
			info,err := getFileInfo(file_)
			// 2023/1/13 编辑
			// modTime = getFileMod(file_)
			if err != nil { return appendError("get file modTime err -Scout",file_) }
			info.Path = file_
			v,ok := s.filePaths.Load(file_)
			// 2023/1/13 添加
			// 情况一、新增文件
			if !ok {
				if s.Debug == "enable" && s.RunMode == RunnMode_AllChange{ log.Println("RunMode AllChange Create") }
				info.Type = ChangeType_Create

				cp = append(cp,info)

				s.filePaths.Store(file_,info.ModTime.UnixNano())
				if s.RunMode == RunnMode_ChangeOnce {
					isRunnMode_ChangeOnce_ok = true
					if s.Debug == "enable"{ log.Println("RunMode ChangeOnce create ") }
					break
				}
				continue
			}
			// 情况二、修改文件
			if v != info.ModTime.UnixNano() {
				if s.Debug == "enable" && s.RunMode == RunnMode_AllChange{ log.Println("RunMode AllChange Update") }
				info.Type = ChangeType_Update
				cp = append(cp, info)
				s.filePaths.Store(file_,info.ModTime.UnixNano())
				if s.RunMode == RunnMode_ChangeOnce {
					isRunnMode_ChangeOnce_ok = true
					if s.Debug == "enable"{ log.Println("RunMode ChangeOnce Update ") }
					break
				}
				continue
			}

		}
		// 情况三、删除文件
		s.filePaths.Range(func(key, value any) bool {

			if s.RunMode == RunnMode_ChangeOnce && isRunnMode_ChangeOnce_ok {
				if s.Debug == "enable"{
					log.Println("RunMode ChangeOnce delete no Scout")
				}
				return false
			}
			fn := key.(string)

			if !isRepetition(files,fn) {
				cp = append(cp, &FileInfo{Name: fn,Type: ChangeType_Del})
				s.filePaths.Delete(key)
				isRunnMode_ChangeOnce_ok = true
				if s.Debug == "enable" && s.RunMode == RunnMode_AllChange{
					log.Println("RunMode AllChange delete")
				}
			}
			return true
		})

		if isRunnMode_ChangeOnce_ok {
			changeFunc(cp)
		}

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

// 获取文件信息 替换 getFileMod
func getFileInfo(_path string) (*FileInfo,error) {
	info ,err := os.Stat(_path)
	if err != nil {
		log.Println(err)
		return nil,err
	}

	return &FileInfo{
		Name:    info.Name(),
		Size:    info.Size(),
		Mode:    info.Mode(),
		ModTime: info.ModTime(),
		IsDir:   info.IsDir(),
	},nil
}


