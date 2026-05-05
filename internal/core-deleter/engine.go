package core_deleter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"github.com/m0090-dev/eec/internal/ext/interfaces/impl"
	"github.com/m0090-dev/eec/internal/ext/types"
	"github.com/rs/zerolog/log"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ====================
// プロセスが終了するのを待機する関数
// ====================
func waitForProcessTermination(rt interfaces.Runtime, pid int) error {
	for {
		var (
			name string
			args []string
		)

		switch rt.Env().GOOS() {
		case "windows":
			name = "tasklist"
			args = []string{"/FI", fmt.Sprintf("PID eq %d", pid)}
		default:
			name = "ps"
			args = []string{"-p", strconv.Itoa(pid)}
		}

		cmd, err := rt.Executor().Command(
			name,
			args,
			rt.Env().Environ(),
			rt.Console().Stdin(),
			nil,
			rt.Console().Stderr(),
			true,
		)
		if err != nil {
			return fmt.Errorf("failed to create check command: %w", err)
		}

		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to check process: %w", err)
		}

		if strings.Contains(string(output), strconv.Itoa(pid)) {
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	return nil
}

type Engine struct {
	Runtime interfaces.Runtime
}

func (e *Engine) FS() interfaces.FS                   { return e.Runtime.FS() }
func (e *Engine) Env() interfaces.Env                 { return e.Runtime.Env() }
func (e *Engine) Executor() interfaces.Executor       { return e.Runtime.Executor() }
func (e *Engine) CommandLine() interfaces.CommandLine { return e.Runtime.CommandLine() }
func (e *Engine) Console() interfaces.Console         { return e.Runtime.Console() }

func NewEngine(rt interfaces.Runtime) *Engine {
	if rt == nil {
		temp := impl.DefaultRuntime{}
		rt = &temp
	}
	return &Engine{
		Runtime: rt,
	}
}

func (e *Engine) Run() error {
	tempDir := e.FS().TempDir()
	manifestPath := filepath.Join(tempDir, types.DEFAULT_MANIFEST_FILE_NAME+".jsonl")

	for {
		if _, err := e.FS().Stat(manifestPath); e.FS().IsNotExist(err) {
			log.Info().Msg("Manifest does not exist. Nothing to clean.")
			time.Sleep(3 * time.Second)
			continue
		}

		file, err := e.FS().Open(manifestPath)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open manifest")
			time.Sleep(3 * time.Second)
			continue
		}

		scanner := bufio.NewScanner(file)
		var remaining []types.ManifestEntry

		for scanner.Scan() {
			line := scanner.Bytes()
			if len(strings.TrimSpace(string(line))) == 0 {
				continue
			}

			// JSON Lines 形式でパース、失敗時は旧テキスト形式にフォールバック
			var entry types.ManifestEntry
			if err := json.Unmarshal(line, &entry); err != nil {
				// 旧形式: "<path> <pid>"
				parts := strings.Fields(string(line))
				if len(parts) != 2 {
					log.Error().Str("line", string(line)).Msg("Invalid line in manifest")
					continue
				}
				entry.TempFilePath = parts[0]
				entry.EECPID, _ = strconv.Atoi(parts[1])
			}

			if entry.EECPID > 0 {
				if err := waitForProcessTermination(e.Runtime, entry.EECPID); err != nil {
					log.Error().Err(err).Int("pid", entry.EECPID).Msg("Failed waiting for process")
				}
			}

			if _, err := e.FS().Stat(entry.TempFilePath); err == nil {
				if err := e.FS().Remove(entry.TempFilePath); err != nil {
					log.Error().Err(err).Str("tempFilePath", entry.TempFilePath).Msg("Failed to delete temp file")
					remaining = append(remaining, entry)
				} else {
					log.Info().Str("tempFilePath", entry.TempFilePath).Msg("Deleted temp file")
				}
			} else {
				log.Info().Str("tempFilePath", entry.TempFilePath).Msg("Temp file already removed")
			}
		}

		file.Close()

		if len(remaining) == 0 {
			if err := e.FS().Remove(manifestPath); err != nil {
				log.Error().Err(err).Msg("Failed to delete manifest file")
			} else {
				log.Info().Msg("Deleted manifest file")
			}
			break
		}

		// 残エントリを JSON Lines で書き直す
		var sb strings.Builder
		for _, entry := range remaining {
			b, _ := json.Marshal(entry)
			sb.Write(b)
			sb.WriteByte('\n')
		}
		e.FS().WriteFile(manifestPath, []byte(sb.String()), 0644)
		log.Info().Msg("Updated manifest with remaining entries")

		time.Sleep(5 * time.Second)
	}

	if err := SendNotification("eec-deleter", "完了メッセージ", "一時ファイル等の削除が完了しました"); err != nil {
		return err
	}
	return nil
}
