package go_scout

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	SkipDir = fs.SkipDir
	SkipAll = fs.SkipAll
)

type Config struct {
	Paths     []string
	Sleep     time.Duration
	EnableHex bool
	EventNum  uint
}

func (c *Config) init() error {
	if len(c.Paths) == 0 {
		return fmt.Errorf("paths can‘t be empty")
	}
	for i, path := range c.Paths {
		c.Paths[i] = filepath.Clean(path)
	}
	return nil
}

type EventValue struct {
	Type     EventType
	Error    error
	FileInfo *FileInfo
}

type Scout struct {
	cache       map[string]*FileInfo
	state       bool
	WalkErrFunc func(path string, info os.FileInfo, err error) error
	FilterFunc  func(path string, info os.FileInfo) bool
	// BeforeWalkPathFunc 开始搜索Paths中的路径
	BeforeWalkPathFunc func()
	// AfterWalkPathFunc 搜索之后的所有文件
	AfterWalkPathFunc func(*map[string]*FileInfo)
	// BeforeCalculateFunc 开始计算差异
	BeforeCalculateFunc func()
	// AfterCalculateFunc 计算结束，并已经通知到EventChan中
	AfterCalculateFunc func()
	EventChan          chan *EventValue
	lock               *sync.RWMutex
	wg                 *sync.WaitGroup
	Conf               *Config
}

// NewScout 创建一个侦查实例
func NewScout(conf *Config) (*Scout, error) {
	s := new(Scout)
	s.Conf = conf
	err := s.Conf.init()
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Scout) init() (err []error) {
	s.state = true
	s.wg = &sync.WaitGroup{}
	s.lock = &sync.RWMutex{}
	s.cache, err = s.getFileInfos()
	if len(err) > 0 {
		return err
	}
	for _, info := range s.cache {
		s.EventChan <- &EventValue{
			Type:     EventType_Init,
			FileInfo: info,
		}
	}
	return nil
}

// Start 开启
func (s *Scout) Start() {
	s.EventChan = make(chan *EventValue, s.Conf.EventNum)
	go func() {
		var errs []error
		defer func() {
			for _, err := range errs {
				s.EventChan <- &EventValue{
					Type:  EventType_Error,
					Error: err,
				}
			}
			close(s.EventChan)
		}()
		if errs = s.init(); len(errs) > 0 {
			return
		}
		var fileInfo map[string]*FileInfo
		for s.state {
			fileInfo, errs = s.getFileInfos()
			if len(errs) > 0 {
				return
			}
			s.calculate(fileInfo)
		}
	}()
}

// Stop 停止检测实例
func (s *Scout) Stop() {
	s.state = false
}

func (s *Scout) Restart() {
	s.Stop()
	time.Sleep(s.Conf.Sleep)
	s.Start()
}

func (s *Scout) calculate(newFiles map[string]*FileInfo) {
	if s.BeforeCalculateFunc != nil {
		s.BeforeCalculateFunc()
	}
	if s.AfterCalculateFunc != nil {
		s.AfterCalculateFunc()
	}
	for _, newF := range newFiles {
		v, ok := s.cache[newF.path]
		//判断是否新建
		if !ok {
			s.cache[newF.path] = newF
			s.EventChan <- &EventValue{
				Type:     EventType_Create,
				FileInfo: newF,
			}
			continue
		}
		//判断是否改变
		if !newF.IsDir() && newF.ModTime().UnixNano() != v.ModTime().UnixNano() {
			s.cache[newF.path] = newF
			if s.Conf.EnableHex && newF.hash == v.hash {
				continue
			}
			s.EventChan <- &EventValue{
				Type:     EventType_Change,
				FileInfo: newF,
			}
		}
	}
	var ok bool
	for s2, info := range s.cache {
		_, ok = newFiles[s2]
		if !ok {
			s.EventChan <- &EventValue{
				Type:     EventType_Remove,
				FileInfo: info,
			}
			delete(s.cache, s2)
		}
	}
	return
}

func (s *Scout) getFileInfos() (result map[string]*FileInfo, errs []error) {
	if s.BeforeWalkPathFunc != nil {
		s.BeforeWalkPathFunc()
	}
	if s.AfterWalkPathFunc != nil {
		defer s.AfterWalkPathFunc(&result)
	}
	result = make(map[string]*FileInfo, len(s.Conf.Paths))
	errs = make([]error, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		for _, path := range s.Conf.Paths {
			s.wg.Add(1)
			go func(ctx context.Context, path string) {
				defer func() {
					s.wg.Done()
				}()
				var err error
				var done = make(chan struct{})
				var infos []*FileInfo
				go func() {
					infos, err = s.getFileInfo(path)
					if err != nil {
						cancel()
						errs = append(errs, err)
						return
					}
					done <- struct{}{}
				}()
				select {
				case <-ctx.Done():
					return
				case <-done:
					s.lock.Lock()
					for _, info := range infos {
						result[info.path] = info
					}
					s.lock.Unlock()
				}
			}(ctx, path)
		}
		s.wg.Wait()
		cancel()
	}()
	<-ctx.Done()
	return result, errs
}

func (s *Scout) getFileInfo(path string) (fileInfo []*FileInfo, err error) {
	fileInfo = make([]*FileInfo, 0, 10)
	err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			if s.WalkErrFunc != nil {
				return s.WalkErrFunc(path, info, err)
			}
			return err
		}
		if s.FilterFunc != nil {
			if !s.FilterFunc(path, info) {
				return nil
			}
		}
		tmpInfo := new(FileInfo)
		tmpInfo.path = path
		tmpInfo.FileInfo = info
		fileInfo = append(fileInfo, tmpInfo)
		if !info.IsDir() && s.Conf.EnableHex {
			v, ok := s.cache[path]
			if ok && v.FileInfo.ModTime().UnixNano() != info.ModTime().UnixNano() {
				if err = tmpInfo.calculateHash(); err != nil {
					if s.WalkErrFunc == nil {
						return err
					}
					if err = s.WalkErrFunc(path, info, err); err != nil {
						errStr := err.Error()
						if errStr != fs.SkipAll.Error() && errStr != fs.SkipDir.Error() {
							return err
						}
					}
				}
			}
		}
		return nil
	})
	return fileInfo, err
}

type FileInfo struct {
	os.FileInfo
	path string
	hash string
}

func (f *FileInfo) Path() string {
	return f.path
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
	return fmt.Sprintf("FileInfo: %v, Path: %v,modTime: %v, IsDir: %v, hex: %s", f.Name(), f.path, f.ModTime().String(), f.IsDir(), f.hash)
}

type EventType uint8

func (e EventType) String() string {
	switch e {
	case EventType_Init:
		return "init"
	case EventType_Create:
		return "create"
	case EventType_Change:
		return "change"
	case EventType_Remove:
		return "remove"
	case EventType_Error:
		return "error"
	default:
		return "invalid event type"
	}
}

const (
	EventType_Init EventType = iota
	EventType_Create
	EventType_Change
	EventType_Remove
	EventType_Error
)
