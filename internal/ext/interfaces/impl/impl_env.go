package impl

import (
	"os"
	"runtime"
)

// OSFS is a thin wrapper over os calls.
type OSEnv struct{}

func (OSEnv) Environ() []string                     { return os.Environ() }
func (OSEnv) Unsetenv(key string) error             { return os.Unsetenv(key) }
func (OSEnv) Setenv(key string, value string) error { return os.Setenv(key, value) }
func (OSEnv) UserHomeDir() (string, error)          { return os.UserHomeDir() }
func (OSEnv) GOOS() string                          { return runtime.GOOS }
func (OSEnv) PathListSeparator() string             { return string(os.PathListSeparator) }
func (OSEnv) LookupEnv(key string) (string, bool)   { return os.LookupEnv(key) }
