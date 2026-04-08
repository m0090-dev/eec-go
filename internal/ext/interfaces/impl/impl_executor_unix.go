//go:build linux || darwin
// +build linux darwin

// executor_unix.go
package impl

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// DefaultExecutor uses os/exec
type DefaultExecutor struct{}

func (d DefaultExecutor) Command(path string, args []string, env []string, stdin, stdout, stderr *os.File, hideWindow bool) (*exec.Cmd, error) {
	var cmd *exec.Cmd
	cmd = exec.Command(path, args...)
	cmd.Env = env
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}

	return cmd, nil
}

// StartProcess: Command を呼んでから Start する
func (d DefaultExecutor) StartProcess(path string, args []string, env []string, stdin, stdout, stderr *os.File, hideWindow bool) (*exec.Cmd, error) {
	cmd, err := d.Command(path, args, env, stdin, stdout, stderr, hideWindow)
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func (d DefaultExecutor) WaitProcess(proc *os.Process, timeout time.Duration) error {
	// We need the *Cmd to call Wait; but we only have *os.Process here.
	// Simpler: poll process state.
	done := make(chan error, 1)
	go func() {
		_, err := proc.Wait()
		done <- err
	}()
	if timeout == 0 {
		return <-done
	}
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("wait timeout after %s", timeout)
	}
}
func (d DefaultExecutor) Getpid() int { return os.Getpid() }

func (d DefaultExecutor) FindProcess(pid int) (*os.Process, error) { return os.FindProcess(pid) }
