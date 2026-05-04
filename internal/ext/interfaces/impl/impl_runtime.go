package impl

import (
	"github.com/m0090-dev/eec/internal/ext/interfaces"
)

type DefaultRuntime struct{}

func (DefaultRuntime) FS() interfaces.FS { return DefaultFS{} }

func (DefaultRuntime) Env() interfaces.Env { return DefaultEnv{} }

func (DefaultRuntime) Executor() interfaces.Executor { return DefaultExecutor{} }

func (DefaultRuntime) CommandLine() interfaces.CommandLine { return DefaultCommandLine{} }

func (DefaultRuntime) Console() interfaces.Console { return DefaultConsole{} }

func (DefaultRuntime) Logger() interfaces.Logger { return NewDefaultLogger() }
