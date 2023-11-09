package go_scout

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

type Option interface{}

type FilterFunc func(name, fullPath string) bool

type Scout struct {
	cache map[string]*FileInfo
	// 休眠时长
	sleep time.Duration
	// 侦察变化的路径
	root      string
	state     bool
	lock      sync.RWMutex
	hashCheck bool
	filter    FilterFunc
}

// NewScout 侦察的文件或目录 root 目录
func NewScout(root string, opt ...Option) (*Scout, error) {
	_, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	s := new(Scout)
	s.state = true
	s.sleep = time.Second
	s.filter = func(name, fullPath string) bool { return true }
	s.root = root
	s.cache = make(map[string]*FileInfo)
	for _, option := range opt {
		option.(func(scout *Scout))(s)
	}
	files, err := GetFiles(s.root, s.filter, s.hashCheck)
	if err != nil {
		return nil, err
	}
	for _, item := range files {
		s.cache[item.id()] = item
	}
	return s, nil
}

// WithScoutSleep 检测休眠时间
func WithScoutSleep(t time.Duration) Option {
	return func(s *Scout) {
		s.sleep = t
	}
}

func WithScoutEnableHashCheck(t bool) Option {
	return func(s *Scout) {
		s.hashCheck = t
	}
}

// WithScoutFilterFunc 过滤那些不需要监控的文件，name: 文件名,fullPath: 路径+文件名，返回值决定是否监控
func WithScoutFilterFunc(cb func(name, fullPath string) bool) Option {
	return func(s *Scout) {
		s.filter = cb
	}
}

func (s *Scout) FileInfoMap() map[string]*FileInfo {
	return s.cache
}

// running Scout 开始侦察文件变化 入参是一个回调方法 当侦擦到变化时调用回调函数
func (s *Scout) Start(changeFunc func(info []*FileInfo)) error {
	for s.state {
		time.Sleep(s.sleep)
		files, err := GetFiles(s.root, s.filter, s.hashCheck)
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
			if s.hashCheck {

			}
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
	md5Val string
}

func (f *FileInfo) id() string {
	return f.Path + f.Name()
}

func (f *FileInfo) String() string {
	return fmt.Sprintf("name: %v, Path: %v,modTime: %v, IsDir: %v, changeType: %v md5: %s", f.Name(), f.Path, f.ModTime().String(), f.IsDir(), f.ChangeType.String(), f.md5Val)
}

// GetFiles 获取路径内包含的所有文件、文件夹，root: 路径，filterFunc: 过滤掉那些文件有返回值决定，isCalculateMD5: 是否计算MD5值
func GetFiles(root string, filterFunc FilterFunc, isCalculateMD5 bool) ([]*FileInfo, error) {
	files := make([]*FileInfo, 0, 100)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !filterFunc(info.Name(), path) {
			return nil
		}
		if err != nil && !strings.Contains(err.Error(), "The system cannot find the file specified.") {
			return err
		}
		var md5Val string
		if !info.IsDir() && isCalculateMD5 {
			md5Val, err = calculateMD5(path)
			if err != nil {
				return err
			}
		}
		files = append(files, &FileInfo{FileInfo: info, Path: path, md5Val: md5Val})
		return nil
	})
	return files, err
}

func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("calculateMD5 err: -1 %v", err)
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("calculateMD5 err: -2 %v", err)
	}
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}
