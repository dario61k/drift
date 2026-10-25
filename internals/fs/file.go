package fs

import (
	"path/filepath"
	"os"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"time"
)


type FileInfo struct {
	Route   string
	ModTime time.Time
}

func CreateDir(path, dirName string) (string, error) {

	dirPath := filepath.Join(path, dirName)
	f, err := os.Stat(dirPath)

	if err != nil && !os.IsNotExist(err){
		return "", err
	}

	if f == nil {
		err = os.Mkdir(dirPath, 0755)
		if err != nil {
			return "", err
		}
	}
	
	newDirPath := filepath.Join(path, dirName)
	return newDirPath, nil
}

func GetFileInfo(path string) (os.FileInfo, error) {

	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return fi, nil
}

func ScanDir(path string) ([]FileInfo, error) {
	
	var files []FileInfo
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, e := range entries {
		if e.IsDir() {
			dirPath := filepath.Join(path, e.Name())
			subFiles, err := ScanDir(dirPath)
			if err != nil {
				return nil, err
			}
			files = append(files, subFiles...)
		} else {
			filePath := filepath.Join(path, e.Name())
			hash, err := GetFileInfo(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil, nil
				}
				return nil, err
			}
			files = append(files, FileInfo{Route: filePath, ModTime: hash.ModTime()})
		}
	}

	return files, nil
}

func OpenFile(path string) ([]byte, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func Hashsha256(path string) (string, error) {

	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
    _, err = io.Copy(h, f)
	if err != nil {
		return "", err
	}	

	return hex.EncodeToString(h.Sum(nil)), nil
}