package go_scout

import (
	"fmt"
	"github.com/Li-giegie/go-utils/goruntine_manager"
	"os"
	"sync"
	"time"
)

type ScouterI interface {
	Root() string                           //目录路径
	StartEvent(info []*FileInfo)            //首次启动后触发检测目录的文件信息
	CreateEvent(info []*FileInfo)           //检测到创建行为触发
	ChangeEvent(info []*FileInfo)           //检测到有变更行为触发
	RemoveEvent(info []*FileInfo)           //检测到有删除行为触发
	ErrorEvent(err error) (isContinue bool) //isContinue：是否继续，true继续，false遇到错误终止
}

type Option func(scout *scouter)

// FilterFunc 过滤不需要的文件，回调返回false为过滤
type FilterFunc func(path string, info os.FileInfo) bool

type ScoutI interface {
	Start() error
	Stop()
}

type scouter struct {
	cache        map[string]*FileInfo
	sleep        time.Duration // 休眠时长
	state        bool
	lock         sync.RWMutex
	hashCheck    bool
	goroutineNum int
	filter       FilterFunc
	ScouterI
	goruntine_manager.GoroutineManagerI
}

// NewScout 创建一个侦查实例
func NewScout(sc ScouterI, opt ...Option) (ScoutI, error) {
	_, err := os.Stat(sc.Root())
	if err != nil {
		return nil, err
	}
	s := new(scouter)
	s.state = true
	s.sleep = DEFAULT_SLEEP
	s.filter = DEFAULT_FILTERFUNC
	s.ScouterI = sc
	s.goroutineNum = DEFAULT_GOROUTINENUM
	for _, option := range opt {
		option(s)
	}
	s.GoroutineManagerI = goruntine_manager.NewGoroutineManger(s.goroutineNum)
	if err = s.GoroutineManagerI.Start(); err != nil {
		return nil, err
	}
	s.cache, err = getFiles(s.ScouterI.Root(), s.filter, sc.ErrorEvent)
	if err != nil {
		return nil, err
	}
	s.ScouterI.StartEvent(mapToSlice(s.cache))
	return s, nil
}

// WithSleep 检测休眠时间
func WithSleep(t time.Duration) Option {
	return func(s *scouter) {
		s.sleep = t
	}
}

// WithEnableHashCheck 是否开启hash检查文件是否更改
func WithEnableHashCheck(t bool) Option {
	return func(s *scouter) {
		s.hashCheck = t
	}
}

// WithFilterFunc 过滤不需要的文件，回调返回false为过滤掉
func WithFilterFunc(cb FilterFunc) Option {
	return func(s *scouter) {
		s.filter = cb
	}
}

// WithGoroutineNum 启用的协程数量，默认值CPU核心数
func WithGoroutineNum(n int) Option {
	return func(s *scouter) {
		if n <= 0 {
			return
		}
		s.goroutineNum = n
	}
}

// Start 开启示例
func (s *scouter) Start() error {
	for s.state {
		time.Sleep(s.sleep)
		newFI, err := getFiles(s.Root(), s.filter, s.ErrorEvent)
		if err != nil {
			return err
		}
		if s.hashCheck {
			newFI = calculateHash(s.cache, newFI, s.Run)
		}
		s.calculate(newFI)
	}
	return nil
}

// Stop 停止检测实例
func (s *scouter) Stop() {
	s.state = false
}

// 计算
func (s *scouter) calculate(newFiles map[string]*FileInfo) {
	//计算新建创建和更新
	createList := make([]*FileInfo, 0, 20)
	updateList := make([]*FileInfo, 0, 20)
	deleteList := make([]*FileInfo, 0, 20)
	for _, newF := range newFiles {
		v, ok := s.cache[newF.path]
		//判断是否新建
		if !ok {
			s.cache[newF.path] = newF
			createList = append(createList, newF)
			continue
		}
		//判断是否改变
		if !newF.IsDir() && newF.modTime != v.modTime {
			s.cache[newF.path] = newF
			if s.hashCheck && newF.hash == v.hash {
				continue
			}
			updateList = append(updateList, newF)
		}
	}
	var ok bool
	for s2, info := range s.cache {
		_, ok = newFiles[s2]
		if !ok {
			deleteList = append(deleteList, info)
			delete(s.cache, s2)
		}
	}
	if len(createList) > 0 {
		s.ScouterI.CreateEvent(createList)
	}
	if len(updateList) > 0 {
		s.ScouterI.ChangeEvent(updateList)
	}
	if len(deleteList) > 0 {
		s.ScouterI.RemoveEvent(deleteList)
	}
	return
}

type FileInfo struct {
	os.FileInfo
	path    string
	hash    string
	modTime int64
}

func (f *FileInfo) GetPath() string {
	return f.path
}

func (f *FileInfo) GetHash() string {
	return f.hash
}

func (f *FileInfo) calculateHash() error {
	if f.IsDir() {
		return nil
	}
	md5Val, err := calculateMD5(f.path)
	f.hash = md5Val
	return err
}

func (f *FileInfo) String() string {
	return fmt.Sprintf("name: %v, Path: %v,modTime: %v, IsDir: %v, md5: %s", f.Name(), f.path, f.ModTime().String(), f.IsDir(), f.hash)
}
