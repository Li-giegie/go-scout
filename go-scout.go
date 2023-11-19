package go_scout

import (
	"fmt"
	"github.com/Li-giegie/go-utils/goruntine_manager"
	"os"
	"sync"
	"time"
)

type ScoutI interface {
	StartEvent(info []*FileInfo)
	CreateEvent(info []*FileInfo)
	ChangeEvent(info []*FileInfo)
	RemoveEvent(info []*FileInfo)
}

type Option interface{}

type FilterFunc func(path string, info os.FileInfo) bool

type Scout struct {
	cache        map[string]*FileInfo
	sleep        time.Duration // 休眠时长
	root         string        // 侦察变化的路径
	state        bool
	lock         sync.RWMutex
	hashCheck    bool
	goroutineNum int
	filter       FilterFunc
	ScoutI
	goruntine_manager.GoroutineManagerI
}

// sleepTime /ms 每一次侦察后休眠时长 理想值 1000
// root	侦察的文件或目录
func NewScout(root string, sc ScoutI, opt ...Option) (*Scout, error) {
	_, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	s := new(Scout)
	s.root = root
	s.state = true
	s.sleep = DEFAULT_SLEEP
	s.filter = DEFAULT_FILTERFUNC
	s.ScoutI = sc
	s.goroutineNum = DEFAULT_GOROUTINENUM
	for _, option := range opt {
		option.(func(scout *Scout))(s)
	}
	s.GoroutineManagerI = goruntine_manager.NewGoroutineManger(s.goroutineNum)
	if err = s.GoroutineManagerI.Start(); err != nil {
		return nil, err
	}
	s.cache, err = GetFiles(s.root, s.filter)
	if err != nil {
		return nil, err
	}
	s.ScoutI.StartEvent(mapToSlice(s.cache))
	return s, nil
}

// WithScoutSleep 检测休眠时间
func WithSleep(t time.Duration) Option {
	return func(s *Scout) {
		s.sleep = t
	}
}

func WithEnableHashCheck(t bool) Option {
	return func(s *Scout) {
		s.hashCheck = t
	}
}

// WithScoutFilterFunc 过滤那些不需要监控的文件，name: 文件名,fullPath: 路径+文件名，返回值决定是否监控
func WithFilterFunc(cb FilterFunc) Option {
	return func(s *Scout) {
		s.filter = cb
	}
}

// WithScoutFilterFunc 过滤那些不需要监控的文件，name: 文件名,fullPath: 路径+文件名，返回值决定是否监控
func WithGoroutineNum(n int) Option {
	return func(s *Scout) {
		if n <= 0 {
			return
		}
		s.goroutineNum = n
	}
}

// running Scout 开始侦察文件变化 入参是一个回调方法 当侦擦到变化时调用回调函数
func (s *Scout) Start() error {
	for s.state {
		time.Sleep(s.sleep)
		newFI, err := GetFiles(s.root, s.filter)
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

func (s *Scout) Stop() {
	s.state = false
}

// 计算
func (s *Scout) calculate(newFiles map[string]*FileInfo) {
	//计算新建创建和更新
	createList := make([]*FileInfo, 0, 20)
	updateList := make([]*FileInfo, 0, 20)
	deleteList := make([]*FileInfo, 0, 20)
	t := time.Now()
	defer func() {
		fmt.Printf("calculate sum: %v\n\n", time.Since(t))
	}()
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
	if createList != nil {
		s.ScoutI.CreateEvent(createList)
	}
	if updateList != nil {
		s.ScoutI.ChangeEvent(updateList)
	}
	if deleteList != nil {
		s.ScoutI.RemoveEvent(deleteList)
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
	return f.hash
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
