package impl

import "os"
import "github.com/m0090-dev/eec/internal/ext/utils/general"

// DefaultFS is a thin wrapper over os calls.
type DefaultFS struct{}

func (DefaultFS) Create(name string) (*os.File, error) { return os.Create(name) }
func (DefaultFS) TempDir() string                      { return os.TempDir() }
func (DefaultFS) FileExists(name string) bool          { return general.FileExists(name) }
func (DefaultFS) MkdirAll(path string, perm uint32) error {
	return os.MkdirAll(path, os.FileMode(perm))
}
func (DefaultFS) WriteFile(name string, data []byte, perm uint32) error {
	return os.WriteFile(name, data, os.FileMode(perm))
}
func (DefaultFS) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }
func (DefaultFS) Remove(name string) error             { return os.Remove(name) }
func (DefaultFS) FileExt(path string) string           { return general.FileExt(path) }
func (DefaultFS) Open(name string) (*os.File, error)   { return os.Open(name) }
func (DefaultFS) OpenFile(name string, flag int, perm uint32) (*os.File, error) {
	return os.OpenFile(name, flag, os.FileMode(perm))
}
func (DefaultFS) FileBase(path string) string           { return general.FileBase(path) }
func (DefaultFS) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (DefaultFS) IsNotExist(err error) bool             { return os.IsNotExist(err) }
func (DefaultFS) O_APPEND() int                         { return os.O_APPEND }
func (DefaultFS) O_CREATE() int                         { return os.O_CREATE }
func (DefaultFS) O_WRONLY() int                         { return os.O_WRONLY }
