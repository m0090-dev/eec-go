package utils

import (
	"os"
	"path/filepath"
)

/*
func FileExists(path string) bool {
    _, err := os.Stat(path)
    return err == nil || !os.IsNotExist(err)
}
*/

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
func FileExt(path string) string {
	return filepath.Ext(path)
}
