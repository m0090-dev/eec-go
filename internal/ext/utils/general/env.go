package general

import (
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"regexp"
	"strings"
)

func findMatchingParen(runes []rune, start int) (content string, nextIdx int) {
	stack := 1
	for i := start; i < len(runes); i++ {
		if runes[i] == '(' {
			stack++
		} else if runes[i] == ')' {
			stack--
			if stack == 0 {
				// 対応する閉じ括弧が見つかった
				return string(runes[start:i]), i
			}
		}
	}
	return "", -1
}

// 型がバラバラな envMap から値を取り出すヘルパー
func getValuesFromAny(envMap any, key string) []string {
	key = strings.ToUpper(key)
	//fmt.Printf("DEBUG: Looking for [%s]\n", key)
	switch m := envMap.(type) {
	case map[string]map[string]struct{}:
		if vMap, ok := m[key]; ok {
			var vals []string
			for v := range vMap {
				vals = append(vals, v)
			}
			return vals
		}
	case map[string][]string:
		return m[key]
	}
	return nil
}

// すべての envMap を OS の環境変数形式 (KEY=VALUE) に変換するヘルパー
func getAllEnvsFromAny(env interfaces.Env, fs interfaces.FS, envMap any) []string {
	sep := string(env.PathListSeparator())
	var result []string
	switch m := envMap.(type) {
	case map[string]map[string]struct{}:
		for k, vMap := range m {
			if strings.HasPrefix(k, "=") || k == "" {
				continue
			}
			var vals []string
			for v := range vMap {
				vals = append(vals, v)
			}
			result = append(result, k+"="+strings.Join(vals, sep))
		}
	case map[string][]string:
		for k, v := range m {
			if strings.HasPrefix(k, "=") || k == "" {
				continue
			}
			result = append(result, k+"="+strings.Join(v, sep))
		}
	}
	return result
}

/*
	func ExpandEnvAndCommands(env interfaces.Env,fs interfaces.FS,input string, envMap any) string {
		sep := string(env.PathListSeparator())

		// 1. 環境変数の展開: ${VAR}
		reVar := regexp.MustCompile(`\$\{([^}]+)\}`)
		result := reVar.ReplaceAllStringFunc(input, func(match string) string {
			submatches := reVar.FindStringSubmatch(match)
			if len(submatches) == 2 {
				// envMap (any型) から値を探す
				vals := getValuesFromAny(envMap, submatches[1])
				if len(vals) > 0 {
					return strings.Join(vals, sep)
				}
				// なければOS環境変数
				if val, ok := env.LookupEnv(submatches[1]); ok {
					return val
				}
			}
			return match
		})

		// 2. コマンド展開: $(cmd)
		reCmd := regexp.MustCompile(`\$\((.+?)\)`)
		result = reCmd.ReplaceAllStringFunc(result, func(match string) string {
			submatches := reCmd.FindStringSubmatch(match)
			if len(submatches) == 2 {
				cmdLine := strings.TrimSpace(submatches[1])
				var cmd *exec.Cmd
				if runtime.GOOS == "windows" {
					cmd = exec.Command("cmd", "/c", cmdLine)
				} else {
					cmd = exec.Command("sh", "-c", cmdLine)
				}

				// これまでの全環境変数をコマンド実行環境に注入
				currentEnv := env.Environ()
				currentEnv = append(currentEnv, getAllEnvsFromAny(env,fs,envMap)...)
				cmd.Env = currentEnv

				out, err := cmd.Output()
				if err != nil {
					return ""
				}
				return strings.TrimSpace(string(out))
			}
			return match
		})
		return result
	}
*/
func replaceVariables(env interfaces.Env, fs interfaces.FS, input string, envMap any) string {
	sep := env.PathListSeparator()
	reVar := regexp.MustCompile(`\$\{([^}]+)\}`)

	return reVar.ReplaceAllStringFunc(input, func(match string) string {
		submatches := reVar.FindStringSubmatch(match)
		if len(submatches) == 2 {
			key := submatches[1]
			// 1. envMap (JSONなど) から値を探す
			vals := getValuesFromAny(envMap, key)
			if len(vals) > 0 {
				return strings.Join(vals, sep)
			}
			// 2. なければ OS 環境変数から探す
			if val, ok := env.LookupEnv(key); ok {
				return val
			}
		}
		return match
	})
}
func ExpandEnvAndCommands(env interfaces.Env, fs interfaces.FS, exec interfaces.Executor, console interfaces.Console, input string, envMap any) string {
	result := replaceVariables(env, fs, input, envMap)
	// --------------------------------------------------
	// 1. 環境変数の展開: ${VAR} (Regex で OK)
	// --------------------------------------------------

	// --------------------------------------------------
	// 2. コマンド展開: $(cmd) (スタック解析でネスト対応)
	// --------------------------------------------------
	runes := []rune(result)
	var sb strings.Builder

	for i := 0; i < len(runes); i++ {
		// "$(" の開始を検知
		if i+1 < len(runes) && runes[i] == '$' && runes[i+1] == '(' {
			start := i + 2
			content, nextIdx := findMatchingParen(runes, start)

			if nextIdx != -1 {
				// 抽出したコマンドラインを Executor 経由で実行
				cmdLine := strings.TrimSpace(content)
				// 再帰的に中身を展開 (例: $(echo $(date)) の内側を先に解決)
				//expandedCmdLine := ExpandEnvAndCommands(env, fs, exec, console, cmdLine, envMap)
				expandedCmdLine := replaceVariables(env, fs, cmdLine, envMap)
				// 実行
				out := executeCommand(env, fs, exec, console, expandedCmdLine, envMap)
				sb.WriteString(out)

				i = nextIdx // 閉じ括弧までスキップ
				continue
			}
		}
		sb.WriteRune(runes[i])
	}

	return sb.String()
}

// 内部用：Executor を使ってコマンドを実行するヘルパー
func executeCommand(env interfaces.Env, fs interfaces.FS, executor interfaces.Executor, console interfaces.Console, cmdLine string, envMap any) string {
	var name string
	var args []string

	// OSごとのシェル呼び出しを定義
	if env.GOOS() == "windows" {
		name = "cmd"
		args = []string{"/c", cmdLine}
	} else {
		name = "sh"
		args = []string{"-c", cmdLine}
	}

	// 環境変数の組み立て
	currentEnv := env.Environ()
	currentEnv = append(currentEnv, getAllEnvsFromAny(env, fs, envMap)...)

	// Executor.Command を呼び出す (exec.Cmd 型を直接宣言せずに実行)
	// stdout を nil にすることで Output() 相当のキャプチャを可能にする設計を想定
	cmd, err := executor.Command(
		name,
		args,
		currentEnv,
		console.Stdin(),
		nil, // stdout: nil を渡して Output() で取得
		console.Stderr(),
		true, // suppress window (Windows)
	)
	if err != nil {
		return ""
	}

	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	res := strings.TrimSpace(string(out))
	if env.GOOS() == "windows" {
		res = strings.Trim(res, "\"")
	}
	return res
}

func ExpandEnvAndCommandsSlice(env interfaces.Env, fs interfaces.FS, exec interfaces.Executor, console interfaces.Console, inputs []string, envMap any) []string {
	result := make([]string, len(inputs))
	for i, s := range inputs {
		result[i] = ExpandEnvAndCommands(env, fs, exec, console, s, envMap)
	}
	return result
}
