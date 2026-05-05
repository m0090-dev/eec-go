package impl

import (
	"github.com/m0090-dev/eec/internal/ext/interfaces"
)

type DefaultRuntime struct {
	logger interfaces.Logger
}

func (r *DefaultRuntime) FS() interfaces.FS                   { return DefaultFS{} }
func (r *DefaultRuntime) Env() interfaces.Env                 { return DefaultEnv{} }
func (r *DefaultRuntime) Executor() interfaces.Executor       { return DefaultExecutor{} }
func (r *DefaultRuntime) CommandLine() interfaces.CommandLine { return DefaultCommandLine{} }
func (r *DefaultRuntime) Console() interfaces.Console         { return DefaultConsole{} }

func (r *DefaultRuntime) Logger() interfaces.Logger {
	if r.logger == nil {
		return NewDefaultLogger()
	}
	return r.logger
}

func (r *DefaultRuntime) SetLogger(l interfaces.Logger) {
	r.logger = l
}
