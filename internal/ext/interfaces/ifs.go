package interfaces

import "os"

// FS is minimal file-system abstraction.
type FS interface {
	Create(name string) (*os.File, error)
	TempDir() string
	FileExists(name string) bool
	MkdirAll(path string, perm uint32) error
	WriteFile(name string, data []byte, perm uint32) error
	ReadFile(name string) ([]byte, error)
	Remove(name string) error
	FileExt(path string) string
	FileBase(path string) string
	Open(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
	IsNotExist(err error) bool
	OpenFile(name string, flag int, perm uint32) (*os.File, error)
	O_APPEND() int
	O_CREATE() int
	O_WRONLY() int
}
