package go_scout

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var DEFAULT_FILTERFUNC = func(path string, info os.FileInfo) bool { return true }
var DEFAULT_SLEEP = time.Second
var DEFAULT_GOROUTINENUM = runtime.NumCPU()

// GetFiles 获取路径内包含的所有文件、文件夹，root: 路径，filterFunc: 过滤掉那些文件有返回值决定，isCalculateMD5: 是否计算MD5值
func GetFiles(root string, filterFunc FilterFunc) (map[string]*FileInfo, error) {
	t := time.Now()
	files := make(map[string]*FileInfo, 100)
	defer func() {
		fmt.Println("GetFiles: ", time.Since(t), len(files))
	}()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if errors.Is(err, filepath.SkipDir) {
				log.Println("[debug] -1 ", err)
				return nil
			}
			if errors.Is(err, os.ErrNotExist) {
				log.Println("[debug] -2 ", err)
				return nil
			}
			if errors.Is(err, os.ErrPermission) {
				log.Println("[waring] -3 ", err)
				return nil
			}
			return err
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

func GetFilesV2(dirname string, filterFunc FilterFunc) (map[string]*FileInfo, error) {
	dirname = strings.TrimSuffix(dirname, string(os.PathSeparator))
	infos, err := ioutil.ReadDir(dirname)
	if err != nil {
		if errors.Is(err, filepath.SkipDir) {
			log.Println("[debug] -1 ", err)
			return nil, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			log.Println("[debug] -2 ", err)
			return nil, nil
		}
		if errors.Is(err, os.ErrPermission) {
			log.Println("[waring] -3 ", err)
			return nil, nil
		}
		return nil, err
	}

	paths := make(map[string]*FileInfo, len(infos))
	info, err := os.Stat(dirname)
	if err != nil {
		if errors.Is(err, filepath.SkipDir) {
			log.Println("[debug] -3 ", err)
			return nil, nil
		}
		if errors.Is(err, os.ErrNotExist) {
			log.Println("[debug] -4 ", err)
			return nil, nil
		}
		if errors.Is(err, os.ErrPermission) {
			log.Println("[waring] -5 ", err)
			return nil, nil
		}
		return nil, err
	}
	if filterFunc == nil || filterFunc != nil && filterFunc(dirname, info) {
		paths[dirname] = &FileInfo{
			FileInfo: info,
			path:     dirname,
			hash:     dirname,
			modTime:  time.Now().UnixNano(),
		}
	}

	for _, info := range infos {
		path := dirname + string(os.PathSeparator) + info.Name()
		realInfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if realInfo.IsDir() {
			tmp, err := GetFilesV2(path, filterFunc)
			if err != nil {
				return nil, err
			}
			for s, fileInfo := range tmp {
				paths[s] = fileInfo
			}
			continue
		}
		if filterFunc != nil && !filterFunc(dirname, info) {
			continue
		}
		paths[path] = &FileInfo{
			FileInfo: realInfo,
			path:     path,
			hash:     "",
			modTime:  time.Now().UnixNano(),
		}
	}
	return paths, nil
}

// 计算新的FileInfo hash,old 提供时间依据
func calculateHash(oldFI, newFI map[string]*FileInfo, goroutine func(func()) error) map[string]*FileInfo {
	var w sync.WaitGroup
	var lock sync.RWMutex
	t := time.Now()
	defer func() {
		fmt.Println("calculateHash: ", time.Since(t))
	}()
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
					log.Printf("[warning] calculate MD5 fail -1 %v %v \n", err, tmp.path)
					return
				}
				log.Printf("[warning] calculate MD5 fail -2 %v %v \n", err, tmp.path)
				return
			}
		})
	}
	w.Wait()
	for _, s := range delList {
		delete(newFI, s)
		fmt.Println("del key: ", s)
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
