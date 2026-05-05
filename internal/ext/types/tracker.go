package types

import (
	"fmt"
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"sort"
	"strings"
)

// EnvSource は環境変数の定義元と上書き情報を保持する。
type EnvSource struct {
	Value        string
	DefinedIn    string // 初回定義ファイル
	OverriddenBy string // 上書きしたファイル（なければ空文字）
}

// EnvTracker は全環境変数の追跡情報を保持する。
type EnvTracker struct {
	Vars map[string]*EnvSource
}

// NewEnvTracker は空の EnvTracker を返す。
func NewEnvTracker() *EnvTracker {
	return &EnvTracker{Vars: make(map[string]*EnvSource)}
}

// TrackEnvOverrides は複数の Config を順に処理し、上書きの発生を追跡する。
// 引数の configs は優先度が低い順（後ろほど強い）で渡すこと。
func TrackEnvOverrides(configs []Config) *EnvTracker {
	tracker := NewEnvTracker()
	for _, cfg := range configs {
		source := cfg.SourcePath
		if source == "" {
			source = "<inline>"
		}
		if cfg.SourcePath == "" && cfg.RawEnvs == nil {
			continue
		}
		for _, env := range cfg.Envs {
			key := strings.ToUpper(env.Key)
			val := fmt.Sprint(env.Value)
			if existing, ok := tracker.Vars[key]; ok {
				// すでに定義済み → 上書き記録
				existing.OverriddenBy = source
				existing.Value = val
			} else {
				tracker.Vars[key] = &EnvSource{
					Value:     val,
					DefinedIn: source,
				}
			}
		}
	}
	return tracker
}

// PrintOverrides はログや標準出力に上書きされた変数の一覧を出す。
// strict=true の場合はエラーを返す。noWarn=true の場合は何もしない。
func (t *EnvTracker) PrintOverrides(logger interfaces.Logger, strict bool, noWarn bool) error {
	if noWarn {
		return nil
	}
	var overridden []string
	for key, src := range t.Vars {
		if src.OverriddenBy != "" {
			overridden = append(overridden, key)
		}
	}
	if len(overridden) == 0 {
		return nil
	}
	sort.Strings(overridden)
	for _, key := range overridden {
		src := t.Vars[key]
		msg := fmt.Sprintf("env override: %s (defined in %s, overridden by %s)",
			key, src.DefinedIn, src.OverriddenBy)
		if strict {
			logger.Error().Msg(msg)
		} else {
			logger.Warn().Msg(msg)
		}
	}
	if strict {
		return fmt.Errorf("%d environment variable(s) were overridden (--duplicate-strict)", len(overridden))
	}
	return nil
}
