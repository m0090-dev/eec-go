package impl

import (
	"os"
	"runtime"
)

// OSFS is a thin wrapper over os calls.
type DefaultEnv struct{}

func (DefaultEnv) Environ() []string                     { return os.Environ() }
func (DefaultEnv) Unsetenv(key string) error             { return os.Unsetenv(key) }
func (DefaultEnv) Setenv(key string, value string) error { return os.Setenv(key, value) }
func (DefaultEnv) UserHomeDir() (string, error)          { return os.UserHomeDir() }
func (DefaultEnv) GOOS() string                          { return runtime.GOOS }
func (DefaultEnv) PathListSeparator() string             { return string(os.PathListSeparator) }
func (DefaultEnv) LookupEnv(key string) (string, bool)   { return os.LookupEnv(key) }
