package utils

import "os"

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
