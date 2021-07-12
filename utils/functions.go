package utils

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
)

func FileExists(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return false
	}
	if fileInfo.IsDir() { //equals fileInfo.Mode().IsDir()
		// file is a directory
		return false
	}
	return true
}

func FilePutContents(data []byte, path string) {
	fmt.Println(path)
	ioutil.WriteFile(path, data, 0644)

}

// CacheDir creates and returns a tempory cert directory under current path
func CacheDir(prefix string) (dir string) {
	if u, _ := user.Current(); u != nil {
		dir = filepath.Join(".", prefix+"-"+u.Username)
		log.Printf("Cache directory : %v \n", dir)
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	return ""
}
