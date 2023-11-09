package go_scout

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type ChangeType byte

const (
	ChangeType_Create ChangeType = 1
	ChangeType_Del    ChangeType = 2
	ChangeType_Update ChangeType = 3
)

func (c ChangeType) String() string {
	switch c {
	case ChangeType_Create:
		return "Create"
	case ChangeType_Update:
		return "Update"
	case ChangeType_Del:
		return "Delete"
	case 0:
		return "null"
	default:
		panic("invalid format :" + strconv.Itoa(int(c)))
	}
}

type Scout struct {
	cache map[string]*FileInfo
	// 休眠时长
	sleep time.Duration
	// 侦察变化的路径
	root  string
	state bool
	lock  sync.RWMutex
}

// sleepTime /ms 每一次侦察后休眠时长 理想值 1000
// root	侦察的文件或目录
func NewScout(root string, sleep time.Duration) (*Scout, error) {
	s := new(Scout)
	s.state = true
	s.root = root
	s.sleep = sleep
	files, err := GetFiles(root)
	if err != nil {
		return nil, err
	}
	s.cache = make(map[string]*FileInfo, len(files))
	for _, item := range files {
		s.cache[item.id()] = item
	}
	return s, nil
}

// running Scout 开始侦察文件变化 入参是一个回调方法 当侦擦到变化时调用回调函数
func (s *Scout) Start(changeFunc func(info []*FileInfo)) error {
	for s.state {
		time.Sleep(s.sleep)
		files, err := GetFiles(s.root)
		if err != nil {
			return err
		}
		result := s.calculate(files)
		changeFunc(result)
	}
	return nil
}

func (s *Scout) Stop() {
	s.state = false
}

// 计算
func (s *Scout) calculate(info []*FileInfo) []*FileInfo {
	result := make([]*FileInfo, 0, len(info))
	//计算新建创建和更新
	for _, item := range info {
		id := item.id()
		v, ok := s.cache[id]
		//判断是否增加
		if !ok {
			item.ChangeType = ChangeType_Create
			s.cache[id] = item
			result = append(result, item)
			continue
		}
		//判断是否改变
		if !item.IsDir() && item.ModTime().UnixNano() != v.ModTime().UnixNano() {
			item.ChangeType = ChangeType_Update
			s.cache[id] = item
			result = append(result, item)
		}
	}
	//计算删除
	var isDel bool
	for s2, cache := range s.cache {
		isDel = true
		for _, item := range info {
			if s2 == item.id() {
				isDel = false
				break
			}
		}
		if isDel {
			cache.ChangeType = ChangeType_Del
			result = append(result, cache)
			delete(s.cache, s2)
		}
	}
	return result
}

type FileInfo struct {
	os.FileInfo
	Path string
	ChangeType
}

func (f *FileInfo) id() string {
	return f.Path + f.Name()
}

func (f *FileInfo) String() string {
	return fmt.Sprintf("name: %v, Path: %v,modTime: %v, IsDir: %v, changeType: %v", f.Name(), f.Path, f.ModTime().String(), f.IsDir(), f.ChangeType.String())
}

func GetFiles(dir string) ([]*FileInfo, error) {
	files := make([]*FileInfo, 0, 100)
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			files = append(files, &FileInfo{FileInfo: info, Path: path})
		}
		return nil
	})
	return files, err
}
