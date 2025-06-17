package meta

import (
	"github.com/google/uuid"
	"github.com/pelletier/go-toml/v2"
	"github.com/rs/zerolog/log"
	"main/utils"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Config struct {
	Envs []Env `toml:"envs"`
	Program ProgramData `toml:"program"` 
}
type ProgramData struct {
	Path string   `toml:"path"`
	Args []string `toml:"args"`
}
type Env struct {
	Key   string      `toml:"key"`
	Value interface{} `toml:"value"`
}

func ReadConfig(fileName string) (Config, error) {
	ext := utils.FileExt(fileName)
	if ext == ".toml" {
		return ReadToml(fileName)
	} else if ext == ".yaml" || ext == ".yml" {
		return ReadYaml(fileName)
	} else if ext == ".json" {
		return ReadJson(fileName)
	}
	return Config{}, nil
}
func ReadInlineConfig(fileName string) (Config, error) {
	ext := utils.FileExt(fileName)
	if ext == ".toml" {
		return ReadInlineToml(fileName)
	} else if ext == ".yaml" || ext == ".yml" {
		return ReadInlineYaml(fileName)
	} else if ext == ".json" {
		return ReadInlineJson(fileName)
	}
	return Config{}, nil
}

func ReadJson(fileName string) (Config, error) {
	return Config{}, nil
}
func ReadYaml(fileName string) (Config, error) {
	return Config{}, nil
}

func ReadToml(fileName string) (Config, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return Config{}, err
	}

	var config Config
	err = toml.Unmarshal(data, &config)
	return config, err
}

func ReadInlineToml(tomlData string) (Config, error) {
	// UUID を使って一時ファイル名を生成
	tmpFileName := filepath.Join(os.TempDir(), "inline-"+uuid.NewString()+".toml")

	// 一時ファイルに書き込み
	err := os.WriteFile(tmpFileName, []byte(tomlData), 0600)
	if err != nil {
		return Config{}, err
	}

	// defer で削除を確実に実行
	defer os.Remove(tmpFileName)

	// 通常の読み込み処理を使う
	return ReadToml(tmpFileName)
}
func ReadInlineJson(fileName string) (Config, error) {
	return Config{}, nil
}
func ReadInlineYaml(fileName string) (Config, error) {
	return Config{}, nil
}

func (c *Config) ApplyEnvs() error {

	// -- OSごとの区切り文字分岐 --
	var separator string
	if runtime.GOOS == "windows" {
		separator = ";" // Windowsではセミコロン
	} else {
		separator = ":" // Unix/Linux/macOSではコロン
	}

	currentPaths := strings.Split(os.Getenv("PATH"), separator)
	for _, env := range c.Envs {
		key := env.Key
		if key == "" {
			log.Warn().Interface("env", env).Msg("envのキーが空です")
			continue
		}

		switch val := env.Value.(type) {
		case string:
			// スカラー文字列
			expanded := utils.ExpandEnvVariables(val)
			os.Setenv(key, expanded)

		case []interface{}:
			// 文字列の配列（interface{}スライス）
			strVals := make([]string, 0, len(val))
			for _, v := range val {
				if s, ok := v.(string); ok {
					strVals = append(strVals, utils.ExpandEnvVariables(s))
				} else {
					log.Warn().
						Str("key", key).
						Interface("element", v).
						Msg("env配列の要素が文字列でない")
				}
			}

			if strings.EqualFold(key, "Path") {
				configPaths := strings.Join(strVals, separator)
				newPaths := append(currentPaths, configPaths)
				os.Setenv(key, strings.Join(newPaths, separator))
			} else {

				os.Setenv(key, strings.Join(strVals, separator))
			}
		default:
			log.Warn().
				Str("key", key).
				Interface("value", env.Value).
				Msg("envの値の型が未対応")
		}
	}

	return nil
}
