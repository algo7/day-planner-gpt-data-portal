package utils

import "os"

// FileExists checks if a file exists and is not a directory before we
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
