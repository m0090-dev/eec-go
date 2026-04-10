package domain

import (
	"fmt"
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"github.com/m0090-dev/eec/internal/ext/types"
	gos "os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

func ReadOrFallback(opts types.RunOptions, os types.OS, logger interfaces.Logger, name string) (types.Config, error) {
	return readOrFallbackInternal(opts, os, logger, name, make(map[string]bool))
}

func ReadOrFallbackRecursive(opts types.RunOptions, os types.OS, logger interfaces.Logger, name string) (types.Config, error) {
	return readOrFallbackInternal(opts, os, logger, name, make(map[string]bool))
}

func readOrFallbackInternal(opts types.RunOptions, os types.OS, logger interfaces.Logger, name string, visited map[string]bool) (types.Config, error) {
	// 循環チェック
	absPath, _ := filepath.Abs(name)
	if visited[absPath] {
		return types.Config{}, fmt.Errorf("circular import detected: %s", absPath)
	}
	visited[absPath] = true

	var cfg types.Config
	if os.FS.FileExists(name) {
		return types.ReadConfig(os, logger, name)
	}

	tagData, err := types.ReadTagData(os, logger, name)
	if err != nil {
		return cfg, err
	}

	env := os.Env.Environ()
	for _, f := range tagData.ImportConfigFiles {
		// ★ここが重要：同じ visited map を渡して再帰する
		fcfg, err := readOrFallbackInternal(opts, os, logger, f, visited)
		if err != nil {
			// 循環参照エラーなら即座に復帰（logger.Warnで流さず、上位にエラーを伝播させる）
			return cfg, err
		}
		env = fcfg.BuildEnvs(os, logger, env, opts.Separator)
	}

	// env → cfg.Envs に変換 (以下、既存ロジック)

	cfg.Envs = nil
	for _, e := range env {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			key, val := parts[0], parts[1]

			// 1. セパレータを確定させる
			sep := opts.Separator
			if sep == "" {
				sep = string(os.Env.PathListSeparator()) // Windowsなら ";"
			}

			// 2. 確定したセパレータで判定・分割
			if strings.Contains(val, sep) {
				cfg.Envs = append(cfg.Envs, types.Environ{
					Key:   key,
					Value: strings.Split(val, sep),
				})
			} else {
				cfg.Envs = append(cfg.Envs, types.Environ{Key: key, Value: val})
			}
		}
	}

	return cfg, nil
}

func IsProcessRunning(os types.OS, logger interfaces.Logger, name string) (bool, error) {
	goos := os.Env.GOOS()
	switch goos {
	case "windows":
		if !strings.HasSuffix(name, ".exe") {
			name += ".exe"
		}
		// exec.Command の代わりに os.Executor.StartProcess を使う
		// cmd.Output() を使う場合、StartProcess の stdout 引数は nil で良い
		cmd, err := os.Executor.Command(
			"tasklist",
			[]string{"/FI", fmt.Sprintf("IMAGENAME eq %s", name)},
			os.Env.Environ(),
			os.Console.Stdin(),
			nil, // Output() が内部でセットするので nil
			os.Console.Stderr(),
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
		cmd, err := os.Executor.Command(
			"pgrep",
			[]string{"-x", name},
			os.Env.Environ(),
			os.Console.Stdin(),
			os.Console.Stdout(),
			os.Console.Stderr(),
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

func IsPIDRunning(os types.OS, logger interfaces.Logger, pid int) (bool, error) {
	if pid <= 0 {
		return false, fmt.Errorf("invalid PID: %d", pid)
	}

	proc, err := os.Executor.FindProcess(pid)
	if err != nil {
		// プロセスが見つからない場合は false,errorを返す
		return false, err
	}

	// プロセスに kill 0 シグナルを送る（Linux/macOS）など、存在確認
	if runtime.GOOS != "windows" {
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

func LaunchDeleter(os types.OS, logger interfaces.Logger, opts types.RunOptions) error {
	// ----------------------*/
	// deleter起動
	// ----------------------*/
	deleterPath := opts.DeleterPath
	deleterHideWindow := opts.DeleterHideWindow
	if deleterPath == "" || !os.FS.FileExists(deleterPath) {
		deleterPath = filepath.Join(types.DEFAULT_DELETER_EXECUTE_NAME)
	}

	running, err := IsProcessRunning(os, logger, types.DEFAULT_DELETER_EXECUTE_NAME)
	if err != nil {
		if os.Env.GOOS() == "linux" || os.Env.GOOS() == "darwin" {
			// Unix系ならエラーをログに出すだけで、running = false として続行
			logger.Debug().Err(err).Msg("process check failed, assuming not running")
			running = false
		} else {
			logger.Error().Err(err).Msg("failed to check process")
			return fmt.Errorf("failed to check process: %w", err)
		}
	}

	if running {
		logger.Debug().Msgf("[%s] は既に実行中です", deleterPath)
	} else {
		logger.Debug().Msgf("[%s] を起動します...", deleterPath)
		var pid int
		var execCmd *exec.Cmd
		if os.Env.GOOS() == "windows" {
			var out, errOut *gos.File
			if !deleterHideWindow {
				out, errOut = os.Console.Stdout(), os.Console.Stderr()
			}
			execCmd, err = os.Executor.StartProcess(
				deleterPath, []string{}, os.Env.Environ(), nil, out, errOut, deleterHideWindow,
			)
			pid = execCmd.Process.Pid
		} else {
			var out, errOut *gos.File
			if !deleterHideWindow {
				out, errOut = os.Console.Stdout(), os.Console.Stderr()
			}
			execCmd, err = os.Executor.StartProcess(
				deleterPath, []string{}, os.Env.Environ(), nil, out, errOut, deleterHideWindow,
			)
			pid = execCmd.Process.Pid
		}
		if err != nil {
			logger.Error().Err(err).Msg("failed to start process")
			return fmt.Errorf("failed to start process: %w", err)
		}
		logger.Info().Msgf("deleter started (pid=%d)", pid)
	}
	return nil
}
