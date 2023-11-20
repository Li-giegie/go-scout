package go_scout

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

var DEFAULT_FILTERFUNC = func(path string, info os.FileInfo) bool { return true }
var DEFAULT_SLEEP = time.Second
var DEFAULT_GOROUTINENUM = runtime.NumCPU()

// GetFiles 获取路径内包含的所有文件、文件夹，root: 路径，filterFunc: 过滤掉那些文件有返回值决定，isCalculateMD5: 是否计算MD5值
func getFiles(root string, filterFunc FilterFunc, errCb func(err error) bool) (map[string]*FileInfo, error) {
	files := make(map[string]*FileInfo, 100)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errCb != nil && !errCb(err) {
				return err
			}
			return filepath.SkipAll
		}
		path = filepath.ToSlash(path)
		if filterFunc != nil && !filterFunc(path, info) {
			return nil
		}
		files[path] = &FileInfo{FileInfo: info, path: path, modTime: info.ModTime().UnixNano()}
		return nil
	})
	return files, err
}

// 计算新的FileInfo hash,old 提供时间依据
func calculateHash(oldFI, newFI map[string]*FileInfo, goroutine func(func()) error) map[string]*FileInfo {
	var w sync.WaitGroup
	var lock sync.RWMutex
	var delList = make([]string, 0, 10)
	for s, info := range newFI {
		v, ok := oldFI[s]
		if ok && v.modTime == info.modTime {
			continue
		}
		w.Add(1)
		tmp := info
		_ = goroutine(func() {
			defer w.Done()
			err := tmp.calculateHash()
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					lock.Lock()
					delList = append(delList, tmp.path)
					lock.Unlock()
					return
				}
				return
			}
		})
	}
	w.Wait()
	for _, s := range delList {
		delete(newFI, s)
	}
	return newFI
}

func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := md5.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("calculateMD5 err: -2 %v", err)
	}
	hashInBytes := hash.Sum(nil)
	return hex.EncodeToString(hashInBytes), nil
}

func mapToSlice(m map[string]*FileInfo) []*FileInfo {
	fs := make([]*FileInfo, 0, len(m))
	for _, info := range m {
		fs = append(fs, info)
	}
	return fs
}
