package domain

import "github.com/m0090-dev/eec/internal/ext/types"
import "github.com/m0090-dev/eec/internal/ext/interfaces"
import "path/filepath"

func ResolveRunOptions(
	opts types.RunOptions,
	tagData types.TagData,
	os types.OS,
	logger interfaces.Logger,
) (configFile string, program string, programArgs []string, finalEnv []string) {

	var config types.Config
	var err error
	allConfigs := []types.Config{}

	// ------------------------
	// ConfigFile の決定
	// ------------------------
	switch {
	case opts.ConfigFile != "":
		configFile = opts.ConfigFile // CLI優先
	case tagData.ConfigFile != "":
		configFile = tagData.ConfigFile // タグ
	}

	// 絶対パス化はこのタイミングで行う
	if configFile != "" {
		if abs, err := filepath.Abs(configFile); err == nil {
			configFile = abs
		} else {
			logger.Warn().Err(err).Msg("failed to resolve absolute path for configFile")
		}
	}

	// ----------------------
	// imports（優先度：低 → 高 の順）
	// ----------------------

	// タグで指定された imports（最も低い）
	for _, imp := range tagData.ImportConfigFiles {
		if cfg, err := ReadOrFallbackRecursive(opts, os, logger, imp); err == nil {
			allConfigs = append(allConfigs, cfg)
		}
	}

	// CLI で指定された imports（中間）
	for _, imp := range opts.Imports {
		if cfg, err := ReadOrFallbackRecursive(opts, os, logger, imp); err == nil {
			allConfigs = append(allConfigs, cfg)
		}
	}

	// ----------------------
	// メイン configとインライン
	// ----------------------
	if configFile != "" && os.FS.FileExists(configFile) {
		config, err = types.ReadConfig(os, logger, configFile)
		if err != nil {
			logger.Error().Err(err).Str("configFile", configFile).Msg("failed to read config")
		} else {
			allConfigs = append(allConfigs, config)
		}
	}

	// 3. 【優先度：高】インライン設定を読み込む (ファイルの設定を上書きできるように最後に配置)
	inlineCfg, err := resolveInlineFromOptions(opts, os, logger)
	if err == nil && (inlineCfg.RawEnvs != nil || inlineCfg.Program.Path != "") {
		allConfigs = append(allConfigs, inlineCfg)

		// もしインライン側に program の指定があれば、それを優先候補にする
		if inlineCfg.Program.Path != "" && opts.Program == "" {
			program = inlineCfg.Program.Path
			if len(inlineCfg.Program.Args) > 0 && len(opts.ProgramArgs) == 0 {
				programArgs = inlineCfg.Program.Args
			}
		}
	}

	// ------------------------
	// Program の決定
	// ------------------------
	switch {
	case opts.Program != "":
		program = opts.Program
	case tagData.Program != "":
		program = tagData.Program
	default:
		program = config.Program.Path
	}

	// ------------------------
	// ProgramArgs の決定
	// ------------------------
	switch {
	case len(opts.ProgramArgs) != 0:
		programArgs = opts.ProgramArgs
	case len(tagData.ProgramArgs) != 0:
		programArgs = tagData.ProgramArgs
	default:
		programArgs = config.Program.Args
	}

	// ------------------------
	// 環境変数を構築
	// ------------------------
	finalEnv = os.Env.Environ()

	for _, cfg := range allConfigs {
		finalEnv = cfg.BuildEnvs(os, logger, finalEnv, opts.Separator)
	}

	return configFile, program, programArgs, finalEnv
}

// ヘルパー: どのインラインフラグを使うか判定
func resolveInlineFromOptions(opts types.RunOptions, os types.OS, logger interfaces.Logger) (types.Config, error) {
	if opts.InlineConfigToml != "" {
		return types.ReadInlineConfig(os, logger, opts.InlineConfigToml, "toml")
	}
	if opts.InlineConfigYaml != "" {
		return types.ReadInlineConfig(os, logger, opts.InlineConfigYaml, "yaml")
	}
	if opts.InlineConfigJson != "" {
		return types.ReadInlineConfig(os, logger, opts.InlineConfigJson, "json")
	}
	if opts.InlineConfig != "" {
		return types.ReadInlineConfig(os, logger, opts.InlineConfig, "")
	}
	return types.Config{}, nil
}

func isConfigPresent(cfg types.Config) bool {
	return cfg.RawEnvs != nil || len(cfg.Configs) > 0 || cfg.Program.Path != ""
}
