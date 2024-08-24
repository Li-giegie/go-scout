package go_scout

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Scout struct {
	state       bool
	WalkErrFunc func(path string, info os.FileInfo, err error) error
	FilterFunc  func(path string, info os.FileInfo) SkipType
	// BeforeCalculateFunc 开始计算差异
	BeforeCalculateFunc func()
	// AfterCalculateFunc 计算结束
	AfterCalculateFunc func()
	EventChan          chan *EventValue
	Conf               *Config
}

// NewScout 创建一个侦查实例
func NewScout(conf *Config) *Scout {
	s := new(Scout)
	s.Conf = conf
	return s
}

// Stop 停止检测实例
func (s *Scout) Stop() {
	s.state = false
}

func (s *Scout) init() error {
	s.EventChan = make(chan *EventValue, s.Conf.EventChanSize)
	s.state = true
	if s.WalkErrFunc == nil {
		s.WalkErrFunc = func(path string, info os.FileInfo, err error) error {
			return err
		}
	}
	if s.FilterFunc == nil {
		s.FilterFunc = func(path string, info os.FileInfo) SkipType {
			return SkipType_NoSkip
		}
	}
	if s.BeforeCalculateFunc == nil {
		s.BeforeCalculateFunc = func() {}
	}
	if s.AfterCalculateFunc == nil {
		s.AfterCalculateFunc = func() {}
	}

	if len(s.Conf.Paths) == 0 {
		return fmt.Errorf("paths can‘t be empty")
	}
	for i, path := range s.Conf.Paths {
		s.Conf.Paths[i] = filepath.Clean(path)
		_, err := os.Stat(s.Conf.Paths[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scout) Start() (err error) {
	if err = s.init(); err != nil {
		return err
	}
	go func() {
		var sErr error
		wg := sync.WaitGroup{}
		defer func() {
			wg.Wait()
			if sErr != nil {
				s.EventChan <- &EventValue{
					Type:  EventType_Error,
					Error: sErr,
				}
			}
			close(s.EventChan)
		}()
		fileList := make([]*FileInfo, 0, len(s.Conf.Paths))
		var ps []*FileInfo
		for _, path := range s.Conf.Paths {
			ps, sErr = s.loadPath(path)
			if sErr != nil {
				return
			}
			if len(ps) == 1 && !ps[0].IsDir() {
				fileList = append(fileList, ps[0])
				continue
			}
			wg.Add(1)
			go s.handleDir(path, &wg, ps)
		}
		if len(fileList) > 0 {
			wg.Add(1)
			go s.handleFile(&wg, fileList...)
		}
	}()
	return nil
}

func (s *Scout) loadPath(path string) (paths []*FileInfo, err error) {
	paths = make([]*FileInfo, 0, 10)
	err = s.walk(path, func(path string, info fs.FileInfo) error {
		fInfo := new(FileInfo)
		fInfo.FileInfo = info
		fInfo.path = path
		paths = append(paths, fInfo)
		if !s.Conf.EnableHex || info.IsDir() {
			s.EventChan <- &EventValue{Type: EventType_Init, Error: nil, FileInfo: fInfo}
			return nil
		}
		if err = fInfo.updateHash(); err != nil {
			s.EventChan <- &EventValue{Type: EventType_Error, Error: err, FileInfo: fInfo}
			return nil
		}
		s.EventChan <- &EventValue{Type: EventType_Init, Error: nil, FileInfo: fInfo}
		return nil
	})
	return
}

func (s *Scout) handleFile(wg *sync.WaitGroup, infos ...*FileInfo) {
	m := make(map[string]*FileInfo, len(infos))
	for _, info := range infos {
		m[info.path] = info
	}
	var err error
	ticker := time.NewTicker(s.Conf.Sleep)
	defer func() {
		ticker.Stop()
		if err != nil {
			s.EventChan <- &EventValue{
				Type:  EventType_Error,
				Error: err,
			}
		}
		wg.Done()
	}()
	for _ = range ticker.C {
		if !s.state {
			return
		}
		s.BeforeCalculateFunc()
		err = s.walkFile(m, func(path string, info fs.FileInfo, exist bool) error {
			var wErr error
			fInfo := m[path]
			if !exist {
				s.EventChan <- &EventValue{Type: EventType_Remove, FileInfo: fInfo}
				delete(m, path)
				return nil
			}
			if fInfo.ModTime().UnixNano() == info.ModTime().UnixNano() {
				return nil
			}
			if s.Conf.EnableHex {
				old := fInfo.hash
				if wErr = fInfo.updateHash(); wErr != nil {
					s.EventChan <- &EventValue{Type: EventType_Error, Error: wErr, FileInfo: fInfo}
					return nil
				}
				if bytesEqual(old, fInfo.hash) {
					return nil
				}
			}
			s.EventChan <- &EventValue{Type: EventType_Change, Error: nil, FileInfo: fInfo}
			return nil
		})
		s.AfterCalculateFunc()
		if err != nil {
			if err.Error() == fs.SkipAll.Error() {
				err = nil
			}
			return
		}
	}
}

func (s *Scout) handleDir(root string, w *sync.WaitGroup, infos []*FileInfo) {
	cache := make(map[string]*fileInfoX)
	for _, info := range infos {
		cache[info.path] = &fileInfoX{info: info}
	}
	var err error
	tick := time.NewTicker(s.Conf.Sleep)
	defer func() {
		tick.Stop()
		if err != nil {
			s.EventChan <- &EventValue{
				Type:  EventType_Error,
				Error: err,
			}
		}
		w.Done()
	}()
	for t2 := range tick.C {
		if !s.state {
			return
		}
		updateAt := t2.UnixNano()
		s.BeforeCalculateFunc()
		err = s.walk(root, func(path string, info fs.FileInfo) error {
			v, ok := cache[path]
			if !ok {
				fInfo := &FileInfo{FileInfo: info, path: path}
				cache[path] = &fileInfoX{info: fInfo, updateAt: updateAt}
				s.EventChan <- &EventValue{Type: EventType_Create, FileInfo: fInfo}
				return nil
			}
			v.updateAt = updateAt
			if info.IsDir() || v.info.ModTime().UnixNano() == info.ModTime().UnixNano() {
				return nil
			}
			v.info.FileInfo = info
			if s.Conf.EnableHex {
				old := v.info.hash
				if err = v.info.updateHash(); err != nil {
					s.EventChan <- &EventValue{Type: EventType_Error, Error: err, FileInfo: v.info}
					return nil
				}
				if bytesEqual(old, v.info.hash) {
					return nil
				}
			}
			s.EventChan <- &EventValue{Type: EventType_Change, Error: err, FileInfo: v.info}
			return nil
		})
		for path, x := range cache {
			if x.updateAt != updateAt {
				s.EventChan <- &EventValue{Type: EventType_Remove, Error: err, FileInfo: x.info}
				delete(cache, path)
			}
		}
		s.AfterCalculateFunc()
	}
}

func (s *Scout) walk(root string, fn func(path string, info fs.FileInfo) error) error {
	return filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return s.WalkErrFunc(path, info, err)
		}
		t := s.FilterFunc(path, info)
		switch t {
		case SkipType_Dir:
			return fs.SkipDir
		case SkipType_All:
			return fs.SkipAll
		case SkipType_File:
			return nil
		default:
			return fn(path, info)
		}
	})
}

func (s *Scout) walkFile(cache map[string]*FileInfo, fn func(path string, info fs.FileInfo, exist bool) error) error {
	for path, _ := range cache {
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				if err = fn(path, info, false); err == nil {
					continue
				}
			}
			if err = s.WalkErrFunc(path, info, err); err == nil {
				continue
			}
			if errStr := err.Error(); errStr == SkipDir.Error() || errStr == SkipAll.Error() {
				continue
			}
			return err
		}
		t := s.FilterFunc(path, info)
		switch t {
		case SkipType_File:
			continue
		case SkipType_All:
			return fs.SkipAll
		}
		if err = fn(path, info, true); err == nil {
			continue
		}
		if errStr := err.Error(); errStr == SkipDir.Error() || errStr == SkipAll.Error() {
			continue
		}
		return err
	}
	return nil
}

type FileInfo struct {
	os.FileInfo
	path string
	hash []byte
}

func (f *FileInfo) Path() string {
	return f.path
}

func (f *FileInfo) updateHash() error {
	if f.IsDir() {
		return nil
	}
	file, err := os.Open(f.path)
	if err != nil {
		return &ErrHash{path: f.path, err: err}
	}
	defer file.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return &ErrHash{path: f.path, err: err}
	}
	f.hash = hash.Sum(nil)
	return nil
}

func (f *FileInfo) String() string {
	return fmt.Sprintf("FileInfo: %v, Path: %v,modTime: %v, IsDir: %v", f.Name(), f.path, f.ModTime().String(), f.IsDir())
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
