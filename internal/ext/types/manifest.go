package types

import (
	"encoding/json"
	"fmt"
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"os"
	"path/filepath"
)

// ManifestEntry は manifest ファイルの 1 エントリ。
type ManifestEntry struct {
	TempFilePath string `json:"temp_file_path"`
	EECPID       int    `json:"eec_pid"`
}

// Manifest は後方互換のために残す既存型。
// 新規コードでは ManifestEntry を直接使うこと。
type Manifest struct {
	TempFilePath string
	EECPID       int
}

// WriteToManifest は manifest ファイルに JSON Lines 形式で追記する。
// 各行が 1 つの ManifestEntry JSON オブジェクト。
func (m *Manifest) WriteToManifest(rt interfaces.Runtime) (string, error) {
	manifestDir := filepath.Dir(m.TempFilePath)
	manifestPath := filepath.Join(manifestDir, DEFAULT_MANIFEST_FILE_NAME+".jsonl")

	entry := ManifestEntry{
		TempFilePath: m.TempFilePath,
		EECPID:       m.EECPID,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return "", fmt.Errorf("failed to marshal manifest entry: %w", err)
	}
	line = append(line, '\n')

	file, err := rt.FS().OpenFile(
		manifestPath,
		rt.FS().O_APPEND()|rt.FS().O_CREATE()|rt.FS().O_WRONLY(),
		uint32(os.FileMode(0644)),
	)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err = file.Write(line); err != nil {
		return "", err
	}
	return manifestPath, nil
}

// ReadManifest は manifest ファイルの全エントリを読み込んで返す。
func ReadManifest(rt interfaces.Runtime) ([]ManifestEntry, error) {
	tmpDir := rt.FS().TempDir()
	manifestPath := filepath.Join(tmpDir, DEFAULT_MANIFEST_FILE_NAME+".jsonl")

	content, err := rt.FS().ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var entries []ManifestEntry
	for _, line := range splitLines(content) {
		if len(line) == 0 {
			continue
		}
		var e ManifestEntry
		if err := json.Unmarshal(line, &e); err != nil {
			// 旧テキスト形式（"<path> <pid>\n"）へのフォールバック
			var path string
			var pid int
			if n, _ := fmt.Sscanf(string(line), "%s %d", &path, &pid); n == 2 {
				entries = append(entries, ManifestEntry{TempFilePath: path, EECPID: pid})
			}
			continue
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func splitLines(data []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, data[start:i])
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, data[start:])
	}
	return lines
}
