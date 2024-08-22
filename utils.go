package go_scout

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

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
