package general

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	//"fmt"
)

/*


// --------------------------
// 環境変数を展開して string を返す関数（$(VAR)形式）
// --------------------------
func ExpandEnvVariables(input string) string {
    re := regexp.MustCompile(`\$\(([^)]+)\)`)
    return re.ReplaceAllStringFunc(input, func(match string) string {
        submatches := re.FindStringSubmatch(match)
        if len(submatches) == 2 {
            if val, ok := os.LookupEnv(submatches[1]); ok {
                return val
            }
        }
        return ""
    })
}


// --------------------------
// ExpandEnvVariables の []string 版
// --------------------------
func ExpandEnvVariablesSlice(inputs []string) []string {
    result := make([]string, len(inputs))
    for i, s := range inputs {
        result[i] = ExpandEnvVariables(s)
    }
    return result
}
*/

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
func getAllEnvsFromAny(envMap any) []string {
	sep := string(os.PathListSeparator)
	var result []string
	switch m := envMap.(type) {
	case map[string]map[string]struct{}:
		for k, vMap := range m {
			var vals []string
			for v := range vMap {
				vals = append(vals, v)
			}
			result = append(result, k+"="+strings.Join(vals, sep))
		}
	case map[string][]string:
		for k, v := range m {
			result = append(result, k+"="+strings.Join(v, sep))
		}
	}
	return result
}

func ExpandEnvAndCommands(input string, envMap any) string {
	sep := string(os.PathListSeparator)

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
			if val, ok := os.LookupEnv(submatches[1]); ok {
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
			currentEnv := os.Environ()
			currentEnv = append(currentEnv, getAllEnvsFromAny(envMap)...)
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

func ExpandEnvAndCommandsSlice(inputs []string, envMap any) []string {
	result := make([]string, len(inputs))
	for i, s := range inputs {
		result[i] = ExpandEnvAndCommands(s, envMap)
	}
	return result
}
