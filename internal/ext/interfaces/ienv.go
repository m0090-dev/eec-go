package interfaces

// Env is minimal env-system abstraction.
type Env interface {
	Environ() []string
	Unsetenv(key string) error
	Setenv(key string, value string) error
	UserHomeDir() (string, error)
	GOOS() string
	PathListSeparator() string
	LookupEnv(key string) (string, bool)
}
