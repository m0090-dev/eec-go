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

	// sh -c 経由で実行
	if _, err := exec.LookPath("sh"); err == nil {
		cmdArgs := append([]string{"-c", path}, args...)
		cmd = exec.Command("sh", cmdArgs...)
	} else {
		// sh すら無い環境用（一応）
		cmd = exec.Command(path, args...)
	}

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

/*
func (d DefaultExecutor) StartProcess(path string, args []string, env []string, stdin, stdout, stderr *os.File, hideWindow bool) (*exec.Cmd, error) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows の場合、cmd.exe が存在するか確認
		if _, err := exec.LookPath("cmd.exe"); err == nil {
			cmd = exec.Command("cmd.exe", "/C", path+" "+strings.Join(args, " "))
		}
	case "linux", "darwin":
		// Unix系の場合、sh が存在するか確認
		if _, err := exec.LookPath("sh"); err == nil {
			cmd = exec.Command("sh", "-c", path+" "+strings.Join(args, " "))
		}
	}

	// どちらも存在しない場合はそのまま実行
	if cmd == nil {
		cmd = exec.Command(path, args...)
	}

	cmd.Env = env
	if stdin != nil { cmd.Stdin = stdin }
	if stdout != nil { cmd.Stdout = stdout }
	if stderr != nil { cmd.Stderr = stderr }
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}
*/

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
