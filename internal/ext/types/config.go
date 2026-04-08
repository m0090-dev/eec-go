package types

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"github.com/m0090-dev/eec/internal/ext/utils/general"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"reflect"
)

type Config struct {
	RawConfigs interface{}  `toml:"configs" yaml:"configs" json:"configs"`
	Configs    []MetaConfig `toml:"-" yaml:"-" json:"-"`
	RawEnvs    interface{}  `toml:"envs" yaml:"envs" json:"envs"`
	Envs       []Environ    `toml:"-" yaml:"-" json:"-"`
	Program    ProgramData  `toml:"program" yaml:"program" json:"program"`
}
type MetaConfig struct {
	Separator   string `toml:"separator" yaml:"separator" json:"separator"`
	Description string `toml:"description" yaml:"description" json:"description"`
}
type ProgramData struct {
	Path string   `toml:"path" yaml:"path" json:"path"`
	Args []string `toml:"args" yaml:"args" json:"args"`
}
type Environ struct {
	Key   string      `toml:"key" yaml:"key" json:"key"`
	Value interface{} `toml:"value" yaml:"value" json:"value"`
}

// extractDependencies は値の中から ${VAR} 形式の依存先を抽出します
func extractDependencies(val interface{}) []string {
	var deps []string
	reVar := regexp.MustCompile(`\$\{([^}]+)\}`) // env.go の正規表現と統一

	process := func(s string) {
		matches := reVar.FindAllStringSubmatch(s, -1)
		for _, m := range matches {
			if len(m) > 1 {
				deps = append(deps, m[1])
			}
		}
	}

	switch v := val.(type) {
	case string:
		process(v)
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				process(s)
			}
		}
	}
	return deps
}

func (c *Config) SortEnvsByDependency() error {
	resolved := make([]Environ, 0, len(c.Envs))
	seen := make(map[string]bool)

	var resolve func(e Environ, visiting map[string]bool) error
	resolve = func(e Environ, visiting map[string]bool) error {
		if visiting[strings.ToUpper(e.Key)] {
			return fmt.Errorf("循環参照を検知しました: %s", e.Key)
		}
		if seen[strings.ToUpper(e.Key)] {
			return nil
		}

		visiting[strings.ToUpper(e.Key)] = true

		// 値から依存している変数名を取得
		deps := extractDependencies(e.Value)
		for _, depKey := range deps {
			// 設定ファイル内の他の変数に依存しているか確認
			for _, other := range c.Envs {
				if strings.EqualFold(other.Key, depKey) {
					if err := resolve(other, visiting); err != nil {
						return err
					}
				}
			}
		}

		visiting[strings.ToUpper(e.Key)] = false
		seen[strings.ToUpper(e.Key)] = true
		resolved = append(resolved, e)
		return nil
	}

	for _, e := range c.Envs {
		if !seen[strings.ToUpper(e.Key)] {
			if err := resolve(e, make(map[string]bool)); err != nil {
				return err
			}
		}
	}

	c.Envs = resolved
	return nil
}

func (c *Config) NormalizeEnvs(logger interfaces.Logger) error {
	if c.RawEnvs == nil {
		return nil
	}

	switch data := c.RawEnvs.(type) {
	case []interface{}:
		// 【パターン1：リスト形式】
		// 1. [ {key: "K", value: "V"}, ... ] (従来形式)
		// 2. [ {TEST: "aiueo"}, ... ] (リスト内直接定義)
		for _, item := range data {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}

			// A. 従来形式のチェック: - key: "NAME", value: "VAL"
			if key, ok := m["key"].(string); ok && key != "" {
				val := m["value"]
				c.Envs = append(c.Envs, Environ{Key: key, Value: val})
				continue
			}

			// B. リスト内直接定義のチェック: - TEST: "aiueo"
			// map の中身をスキャンして最初の 1 つを取り出す
			for k, v := range m {
				if k != "" {
					c.Envs = append(c.Envs, Environ{Key: k, Value: v})
				}
				break // 1つの要素につき1変数を想定
			}
		}

	case map[string]interface{}:
		// 【パターン2：マップ形式】 { "VARIABLE": "VALUE", "PATH": ["A", "B"] }
		for k, v := range data {
			c.Envs = append(c.Envs, Environ{Key: k, Value: v})
		}

	case map[interface{}]interface{}:
		// YAMLパーサーの型互換性ケア
		for k, v := range data {
			if strKey, ok := k.(string); ok {
				c.Envs = append(c.Envs, Environ{Key: strKey, Value: v})
			}
		}
	default:
		// 想定外の型（文字列や数値などが直接 envs に指定された場合）
		return fmt.Errorf("unsupported type for envs: %T (must be map or list)", c.RawEnvs)
	}
	return nil
}
func (c *Config) NormalizeConfigs(logger interfaces.Logger) error {
	if c.RawConfigs == nil {
		return nil
	}

	switch data := c.RawConfigs.(type) {
	case []interface{}:
		// [[configs]] (リスト形式) の処理
		for _, item := range data {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			c.Configs = append(c.Configs, mapToMetaConfig(m))
		}
	case map[string]interface{}:
		// [configs] (マップ形式) の処理
		c.Configs = append(c.Configs, mapToMetaConfig(data))
	case map[interface{}]interface{}:
		// YAML用互換性
		m := make(map[string]interface{})
		for k, v := range data {
			if strKey, ok := k.(string); ok {
				m[strKey] = v
			}
		}
		c.Configs = append(c.Configs, mapToMetaConfig(m))
	}
	return nil
}

// 補助関数: map から MetaConfig 構造体に変換
func mapToMetaConfig(m map[string]interface{}) MetaConfig {
	var mc MetaConfig
	if s, ok := m["separator"].(string); ok {
		mc.Separator = s
	}
	if d, ok := m["description"].(string); ok {
		mc.Description = d
	}
	return mc
}

func ReadConfig(os OS, logger interfaces.Logger, fileName string) (Config, error) {
	ext := general.FileExt(fileName)
	if ext == ".toml" {
		return ReadToml(os, logger, fileName)
	} else if ext == ".yaml" || ext == ".yml" {
		return ReadYaml(os, logger, fileName)
	} else if ext == ".json" {
		return ReadJson(os, logger, fileName)
	} else if ext == ".env" {
		return ReadEnv(os, logger, fileName)
	}
	return Config{}, nil
}

func ReadInlineConfig(os OS, logger interfaces.Logger, content string, format string) (Config, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return Config{}, nil
	}

	var cfg Config
	var err error
	format = strings.ToLower(strings.TrimSpace(format))

	// --- 1. パース処理 (手動指定 or 自動判別) ---
	if format != "" {
		// フォーマットが明示されている場合
		switch format {
		case "toml":
			err = toml.Unmarshal([]byte(content), &cfg)
		case "yaml", "yml":
			err = yaml.Unmarshal([]byte(content), &cfg)
		case "json":
			err = json.Unmarshal([]byte(content), &cfg)
		default:
			return Config{}, fmt.Errorf("unsupported format: %s", format)
		}
	} else {
		// フォーマットが空の場合：自動判別
		// A. JSON: 先頭が { か [ なら JSON として試行
		if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
			if err = json.Unmarshal([]byte(content), &cfg); err == nil {
				goto PARSED
			}
		}

		// B. TOML: 構造が厳格なので YAML より先に試す
		if err = toml.Unmarshal([]byte(content), &cfg); err == nil {
			// TOMLとして受理されても、中身が全くパースできていない(空の)場合は次へ
			if cfg.RawEnvs != nil || len(cfg.Configs) > 0 || cfg.Program.Path != "" {
				goto PARSED
			}
		}

		// C. YAML: 最も寛容なので最後に試す
		if err = yaml.Unmarshal([]byte(content), &cfg); err == nil {
			goto PARSED
		}

		return Config{}, fmt.Errorf("failed to auto-detect inline config format (tried JSON, TOML, YAML)")
	}

PARSED:
	if err != nil {
		return Config{}, fmt.Errorf("parse error (%s): %w", format, err)
	}

	// --- 2. 共通の正規化・依存関係ソート ---
	// 先ほど修正した、errorを返すようになった NormalizeEnvs を使用
	if err := cfg.NormalizeConfigs(logger); err != nil {
		return Config{}, fmt.Errorf("normalize configs error: %w", err)
	}
	if err := cfg.NormalizeEnvs(logger); err != nil {
		return Config{}, fmt.Errorf("normalize error: %w", err)
	}

	if err := cfg.SortEnvsByDependency(); err != nil {
		return Config{}, fmt.Errorf("dependency sort error: %w", err)
	}

	return cfg, nil
}

func ReadJson(os OS, logger interfaces.Logger, fileName string) (Config, error) {
	data, err := os.FS.ReadFile(fileName)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err == nil {
		config.NormalizeConfigs(logger)
		config.NormalizeEnvs(logger)
		config.SortEnvsByDependency()
	}
	return config, err

}
func ReadYaml(os OS, logger interfaces.Logger, fileName string) (Config, error) {
	data, err := os.FS.ReadFile(fileName)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err == nil {
		config.NormalizeConfigs(logger)
		config.NormalizeEnvs(logger)
		config.SortEnvsByDependency()
	}
	return config, err

}

func ReadToml(os OS, logger interfaces.Logger, fileName string) (Config, error) {
	data, err := os.FS.ReadFile(fileName)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = toml.Unmarshal(data, &config)
	if err == nil {
		config.NormalizeConfigs(logger)
		config.NormalizeEnvs(logger)
		config.SortEnvsByDependency()
	}
	return config, err
}

func ReadEnv(os OS, logger interfaces.Logger, fileName string) (Config, error) {
	var config Config
	var envMap map[string]string
	var err error
	envMap, err = godotenv.Read(fileName)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error reading .env file")
	}

	config.Envs = make([]Environ, 0, len(envMap))
	for key, value := range envMap {
		config.Envs = append(config.Envs, Environ{
			Key:   key,
			Value: value,
		})
	}

	return config, nil
}

func ReadInlineToml(os OS, logger interfaces.Logger, tomlData string) (Config, error) {
	// UUID を使って一時ファイル名を生成
	tmpFileName := filepath.Join(os.FS.TempDir(), "inline-"+uuid.NewString()+".toml")

	// 一時ファイルに書き込み
	err := os.FS.WriteFile(tmpFileName, []byte(tomlData), 0600)
	if err != nil {
		return Config{}, err
	}

	// defer で削除を確実に実行
	defer os.FS.Remove(tmpFileName)

	// 通常の読み込み処理を使う
	return ReadToml(os, logger, tmpFileName)
}
func ReadInlineJson(os OS, logger interfaces.Logger, jsonData string) (Config, error) {
	// UUID を使って一時ファイル名を生成
	tmpFileName := filepath.Join(os.FS.TempDir(), "inline-"+uuid.NewString()+".json")

	// 一時ファイルに書き込み
	err := os.FS.WriteFile(tmpFileName, []byte(jsonData), 0600)
	if err != nil {
		return Config{}, err
	}

	// defer で削除を確実に実行
	defer os.FS.Remove(tmpFileName)

	// 通常の読み込み処理を使う
	return ReadJson(os, logger, tmpFileName)

}
func ReadInlineYaml(os OS, logger interfaces.Logger, yamlData string) (Config, error) {
	// UUID を使って一時ファイル名を生成
	tmpFileName := filepath.Join(os.FS.TempDir(), "inline-"+uuid.NewString()+".yaml")

	// 一時ファイルに書き込み
	err := os.FS.WriteFile(tmpFileName, []byte(yamlData), 0600)
	if err != nil {
		return Config{}, err
	}

	// defer で削除を確実に実行
	defer os.FS.Remove(tmpFileName)

	// 通常の読み込み処理を使う
	return ReadYaml(os, logger, tmpFileName)

}

func (c *Config) BuildEnvs(os OS, logger interfaces.Logger, baseEnv []string, separator string) []string {

	// Description 出力とセパレータ決定ロジック（中略）
	for _, cfgs := range c.Configs {
		if cfgs.Description != "" {
			logger.Debug().Str("Config Description", cfgs.Description).Msg("")
		}
	}
	if separator == "" {
		for _, cfgs := range c.Configs {
			if cfgs.Separator != "" {
				separator = cfgs.Separator
				break
			}
		}
	}
	if separator == "" {
		if runtime.GOOS == "windows" {
			separator = ";"
		} else {
			separator = ":"
		}
	}

	envMap := make(map[string][]string)
	// baseEnv (既存の環境変数) を map に変換
	for _, e := range baseEnv {
		if strings.HasPrefix(e, "=") {
			continue
		}
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			key := strings.ToUpper(parts[0])
			if key == "" {
				continue
			}
			envMap[key] = append(envMap[key], parts[1])
		}
	}

	// c.Envs (設定ファイルの envs) を処理
	for _, env := range c.Envs {
		if env.Key == "" {
			logger.Warn().Interface("env", env).Msg("envのキーが空です")
			continue
		}
		keyUpper := strings.ToUpper(env.Key)

		// 1. まず「生の値 (raw)」をスライスとして取り出す
		var rawStrings []string
		isListType := false
		/*
		switch val := env.Value.(type) {
		case string:
			rawStrings = []string{val}
			isListType = false
		case []interface{}:
			isListType = true
			for _, v := range val {
				if s, ok := v.(string); ok {
					rawStrings = append(rawStrings, s)
				}
			}
		default:
			logger.Warn().Interface("env", env).Msg("無効な値タイプ")
			continue
		}
		*/
		rv := reflect.ValueOf(env.Value)
if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
    isListType = true
    for i := 0; i < rv.Len(); i++ {
        // 中身を文字列に変換して追加
        rawStrings = append(rawStrings, fmt.Sprint(rv.Index(i).Interface()))
    }
} else if s, ok := env.Value.(string); ok {
    rawStrings = []string{s}
    isListType = false
} else {
    continue
}
	logger.Debug().
			Str("key", keyUpper).
			Bool("isListType", isListType).
			Interface("rawValues", rawStrings).
			Msg("Processing environment variable")

		// 2. 重要：ここまでの envMap（上の行の変数が反映済み）を使って展開する
		var expandedValues []string
		for _, raw := range rawStrings {
			// この時点の envMap には直前のループの結果が入っている
			expanded := general.ExpandEnvAndCommands(os.Env, os.FS, os.Executor, os.Console, raw, envMap)
			expandedValues = append(expandedValues, expanded)
		}

		// 3. 既存値と展開後の新規値をマージして重複排除
		existing := make(map[string]struct{})
		merged := []string{}
		// 場合によっては使用
		//isSystemList := (keyUpper == "PATH" || keyUpper == "TEMP" || keyUpper == "TMP")
		if isListType {

			// a) 既存の値を登録
			for _, v := range envMap[keyUpper] {
				for _, part := range strings.Split(v, separator) {
					part = strings.TrimSpace(part)
					if part != "" {
						if _, ok := existing[part]; !ok {
							existing[part] = struct{}{}
							merged = append(merged, part)
						}
					}
				}
			}

			// b) 展開された新しい値を登録
			for _, v := range expandedValues {
				for _, part := range strings.Split(v, separator) {
					part = strings.TrimSpace(part)
					if part != "" {
						if _, ok := existing[part]; !ok {
							existing[part] = struct{}{}
							merged = append(merged, part)
						}
					}
				}
			}
		} else {
			merged = expandedValues
		}

		// 4. 重要：envMap を即座に更新 (これで次のループの展開で参照可能になる)
		envMap[keyUpper] = merged
	}

	// map を []string に変換して返す
	newEnv := make([]string, 0, len(envMap))
	for k, v := range envMap {
		joined := strings.Join(v, separator)
		newEnv = append(newEnv, k+"="+joined)
	}

	return newEnv
}
