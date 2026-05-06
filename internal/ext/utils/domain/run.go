package domain

import (
	"fmt"
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"github.com/m0090-dev/eec/internal/ext/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func ReadOrFallback(opts types.RunOptions, rt interfaces.Runtime, name string) (types.Config, error) {
	return readOrFallbackInternal(opts, rt, name, make(map[string]bool))
}

func ReadOrFallbackRecursive(opts types.RunOptions, rt interfaces.Runtime, name string) (types.Config, error) {
	return readOrFallbackInternal(opts, rt, name, make(map[string]bool))
}

func readOrFallbackInternal(opts types.RunOptions, rt interfaces.Runtime, name string, visited map[string]bool) (types.Config, error) {
	// 循環チェック
	absPath, _ := filepath.Abs(name)
	if visited[absPath] {
		return types.Config{}, fmt.Errorf("circular import detected: %s", absPath)
	}
	visited[absPath] = true

	var cfg types.Config
	if rt.FS().FileExists(name) {
		return types.ReadConfig(rt, name)
	}

	tagData, err := types.ReadTagData(rt, name)
	if err != nil {
		return cfg, err
	}
	for _, f := range tagData.ImportConfigFiles {
		// ★ここが重要：同じ visited map を渡して再帰する
		fcfg, err := readOrFallbackInternal(opts, rt, f, visited)
		if err != nil {
			// 循環参照エラーなら即座に復帰（logger.Warnで流さず、上位にエラーを伝播させる）
			return cfg, err
		}
		// env = fcfg.BuildEnvs(rt, env, opts.Separator)
		cfg.Envs = append(cfg.Envs, fcfg.Envs...)
	}
	cfg.SourcePath = name
	return cfg, nil
}

func IsProcessRunning(rt interfaces.Runtime, name string) (bool, error) {
	goos := rt.Env().GOOS()
	switch goos {
	case "windows":
		if !strings.HasSuffix(name, ".exe") {
			name += ".exe"
		}
		// exec.Command の代わりに os.Executor.StartProcess を使う
		// cmd.Output() を使う場合、StartProcess の stdout 引数は nil で良い
		cmd, err := rt.Executor().Command(
			"tasklist",
			[]string{"/FI", fmt.Sprintf("IMAGENAME eq %s", name)},
			rt.Env().Environ(),
			rt.Console().Stdin(),
			nil, // Output() が内部でセットするので nil
			rt.Console().Stderr(),
			true,
		)
		if err != nil {
			return false, err
		}

		// *exec.Cmd なのでそのまま Output() が使える
		output, err := cmd.Output()
		if err != nil {
			return false, err
		}
		return strings.Contains(string(output), name), nil

	case "linux", "darwin":
		cmd, err := rt.Executor().Command(
			"pgrep",
			[]string{"-x", name},
			rt.Env().Environ(),
			rt.Console().Stdin(),
			rt.Console().Stdout(),
			rt.Console().Stderr(),
			true,
		)
		if err != nil {
			return false, err
		}

		// Run() や Wait() で終了コードを確認
		err = cmd.Wait()
		if err == nil {
			return true, nil
		}
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return false, nil
		}
		return false, err

	default:
		return false, fmt.Errorf("unsupported OS: %s", goos)
	}
}

func IsPIDRunning(rt interfaces.Runtime, logger interfaces.Logger, pid int) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid PID: %d", pid)
	}

	proc, err := rt.Executor().FindProcess(pid)
	if err != nil {
		// プロセスが見つからない場合は false,errorを返す
		return false, err
	}

	// プロセスに kill 0 シグナルを送る（Linux/macOS）など、存在確認
	if rt.Env().GOOS() != "windows" {
		err = proc.Signal(syscall.Signal(0))
		if err == nil {
			return true, nil
		}
		if err == syscall.ESRCH {
			return false, nil
		}
		return false, err
	}

	// Windows は FindProcess が返れば存在とみなす
	return true, nil
}

func LaunchDeleter(rt interfaces.Runtime, opts types.RunOptions) error {
	// ----------------------*/
	// deleter起動
	// ----------------------*/
	deleterPath := opts.DeleterPath
	deleterHideWindow := opts.DeleterHideWindow
	if deleterPath == "" || !rt.FS().FileExists(deleterPath) {
		deleterPath = filepath.Join(types.DEFAULT_DELETER_EXECUTE_NAME)
	}

	running, err := IsProcessRunning(rt, types.DEFAULT_DELETER_EXECUTE_NAME)
	if err != nil {
		if rt.Env().GOOS() == "linux" || rt.Env().GOOS() == "darwin" {
			// Unix系ならエラーをログに出すだけで、running = false として続行
			rt.Logger().Debug().Err(err).Msg("process check failed, assuming not running")
			running = false
		} else {
			rt.Logger().Error().Err(err).Msg("failed to check process")
			return fmt.Errorf("failed to check process: %w", err)
		}
	}

	if running {
		rt.Logger().Debug().Msgf("[%s] は既に実行中です", deleterPath)
	} else {
		rt.Logger().Debug().Msgf("[%s] を起動します...", deleterPath)
		var pid int
		var execCmd *exec.Cmd
		if rt.Env().GOOS() == "windows" {
			var out, errOut *os.File
			if !deleterHideWindow {
				out, errOut = rt.Console().Stdout(), rt.Console().Stderr()
			}
			execCmd, err = rt.Executor().StartProcess(
				deleterPath, []string{}, rt.Env().Environ(), nil, out, errOut, deleterHideWindow,
			)
			if err != nil {
				rt.Logger().Error().Err(err).Msg("failed to start process")
				return err // または適切なエラー処理
			}
			pid = execCmd.Process.Pid
		} else {
			var out, errOut *os.File
			if !deleterHideWindow {
				out, errOut = rt.Console().Stdout(), rt.Console().Stderr()
			}
			execCmd, err = rt.Executor().StartProcess(
				deleterPath, []string{}, rt.Env().Environ(), nil, out, errOut, deleterHideWindow,
			)
			if err != nil {
				rt.Logger().Error().Err(err).Msg("failed to start process")
				return err // または適切なエラー処理
			}
			pid = execCmd.Process.Pid
		}
		if err != nil {
			rt.Logger().Error().Err(err).Msg("failed to start process")
			return fmt.Errorf("failed to start process: %w", err)
		}
		rt.Logger().Info().Msgf("deleter started (pid=%d)", pid)
	}
	return nil
}
