package interfaces

type Runtime interface {
	FS() FS

	Env() Env

	Executor() Executor

	CommandLine() CommandLine

	Console() Console

	Logger() Logger
}
