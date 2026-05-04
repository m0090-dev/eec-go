package domain

import (
	"github.com/m0090-dev/eec/internal/ext/interfaces"
	"github.com/m0090-dev/eec/internal/ext/types"
	"github.com/m0090-dev/eec/internal/ext/utils/general"
	"path/filepath"
	"strings"
)

const (
	windowsWrapEECScript = `@echo off
chcp 65001
REM set "eec_deleter=D:\win\program\go\main-project\eec\build\eec-deleter"
REM set "eec_exe=D:\win\program\go\main-project\eec\build\eec"
set "PATH=D:\win\program\go\main-project\eec\build\;%PATH%"
set "eec_deleter=eec-deleter"
set "eec_exe=eec.exe"

tasklist /FI "IMAGENAME eq %eec_deleter%" /NH | find /I "%eec_deleter%" >nul

if "%1"=="run" (
    rem エラーレベルが 0（すでに実行中）か確認
    if %ERRORLEVEL% equ 0 (
        echo [%eec_deleter%] は既に実行中です。
    ) else (
        echo [%eec_deleter%] を起動します…
        powershell -WindowStyle Normal -Command "Start-Process -FilePath '%eec_deleter%' -WindowStyle Hidden"
    )
)
%eec_exe% %*
`

	windowsWrapGEECScript = `@echo off
	chcp 65001
	set "PATH=D:\win\program\go\main-project\eec\build\;%PATH%"
	geec.exe
`

	// Windows batch script (for cmd.exe)
	windowsTagUtilsScript = `@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM Use the first argument as --program
set PROGRAM=%1
shift

REM Concatenate the remaining arguments with commas for --program-args
set ARGS=
:loop
if "%~1"=="" goto run
if defined ARGS (
  set ARGS=!ARGS!,%~1
) else (
  set ARGS=%~1
)
shift
goto loop

:run
eec run --deleter-hide-window --hide-window --tag %TAGNAME% --program %PROGRAM% --program-args=!ARGS!
`

	windowsTagUtilsScriptProgram = `@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM Use the first argument as --program
set PROGRAM=%1
shift

REM Concatenate the remaining arguments with commas for --program-args
set ARGS=
:loop
if "%~1"=="" goto run
if defined ARGS (
  set ARGS=!ARGS!,%~1
) else (
  set ARGS=%~1
)
shift
goto loop

:run
eec run --deleter-hide-window --hide-window --tag %TAGNAME% --program %PROGRAM% 
`

	windowsTagUtilsScriptProgramArgs = `@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

REM Use the first argument as --program
set PROGRAM=%1
shift

REM Concatenate the remaining arguments with commas for --program-args
set ARGS=
:loop
if "%~1"=="" goto run
if defined ARGS (
  set ARGS=!ARGS!,%~1
) else (
  set ARGS=%~1
)
shift
goto loop

:run
eec run --deleter-hide-window --hide-window --tag %TAGNAME% --program-args=!ARGS!
`

	windowsTagSimpleUtilsScript = `@echo off
setlocal enabledelayedexpansion

REM Use the first argument as --program
set PROGRAM=%1
shift

REM Concatenate the remaining arguments with commas for --program-args
set ARGS=
:loop
if "%~1"=="" goto run
if defined ARGS (
  set ARGS=!ARGS!,%~1
) else (
  set ARGS=%~1
)
shift
goto loop

:run
eec run --deleter-hide-window --hide-window --tag %TAGNAME%
`
	unixTagUtilsScript       = ``
	unixTagSimpleUtilsScript = ``
	unixWrapEECScript        = `#!/bin/bash
# UTF-8 前提

# ビルドディレクトリを PATH に追加
BUILD_DIR="/mnt/d/win/program/go/main-project/eec/build"
export PATH="$BUILD_DIR:$PATH"

# 実行ファイル名
eec_deleter="eec-deleter"
eec_exe="eec"

# 第一引数が "run" の場合は deleter を起動
if [[ "$1" == "run" ]]; then
    # プロセスが既に動いているか確認
    if pgrep -x "$eec_deleter" > /dev/null; then
        echo "[$eec_deleter] は既に実行中です。"
    else
        echo "[$eec_deleter] を起動します…"
        # バックグラウンドで起動して出力を捨てる
        nohup "$eec_deleter" >/dev/null 2>&1 &
    fi
fi

# eec 実行、引数をそのまま渡す
"$eec_exe" "$@"
`
)

func toWindowsLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n") // まずCRLFをLFに統一
	s = strings.ReplaceAll(s, "\r", "\n")   // 万一CRのみがあればLFに変換
	return strings.ReplaceAll(s, "\n", "\r\n")
}

func GenWindowsTagUtilsScript(tagName string) string {
	script := strings.ReplaceAll(windowsTagUtilsScript, "%TAGNAME%", tagName)
	return toWindowsLineEndings(script)
}

func GenWindowsSimpleTagUtilsScript(tagName string) string {
	script := strings.ReplaceAll(windowsTagSimpleUtilsScript, "%TAGNAME%", tagName)
	return toWindowsLineEndings(script)
}

func GenWindowsTagUtilsScriptProgram(tagName string) string {
	script := strings.ReplaceAll(windowsTagUtilsScriptProgram, "%TAGNAME%", tagName)
	return toWindowsLineEndings(script)
}
func GenWindowsTagUtilsScriptProgramArgs(tagName string) string {
	script := strings.ReplaceAll(windowsTagUtilsScriptProgramArgs, "%TAGNAME%", tagName)
	return toWindowsLineEndings(script)
}

func GenWindowsWrapScript() string {
	return toWindowsLineEndings(windowsWrapEECScript)
}
func GenWindowsGUIWrapScript() string {
	return toWindowsLineEndings(windowsWrapGEECScript)
}

func GenUnixTagUtilsScript(tagName string) string {
	return ""
}

func GenUnixSimpleTagUtilsScript(tagName string) string {
	return ""
}

func GenUnixWrapScript() string {
	return unixWrapEECScript
}

func GenWrapScript(rt interfaces.Runtime) {
	scriptDir := types.DEFAULT_SCRIPT_DIR
	baseName := "eec"
	guiBaseName := "geec"
	// ディレクトリがなければ作成
	if err := rt.FS().MkdirAll(scriptDir, 0755); err != nil {
		rt.Logger().Error().Err(err).Msg("Failed to create utils script directory")
		return
	}

	// --- Windows用スクリプト ---
	windowsScriptContent := GenWindowsWrapScript()
	windowsScriptFileName := general.AddExtension(baseName, ".cmd")
	windowsScriptFile := filepath.Join(scriptDir, windowsScriptFileName)

	func() {
		file, err := rt.FS().Create(windowsScriptFile)
		if err != nil {
			rt.Logger().Error().Err(err).Str("file", windowsScriptFile).Msg("Failed to create Windows wrap script")
			return
		}
		defer file.Close()
		if _, err := file.WriteString(windowsScriptContent); err != nil {
			rt.Logger().Error().Err(err).Str("file", windowsScriptFile).Msg("Failed to write Windows wrap script")
		}
	}()

	// --- Windows用スクリプト ---
	windowsGUIScriptContent := GenWindowsGUIWrapScript()
	windowsGUIScriptFileName := general.AddExtension(guiBaseName, ".cmd")
	windowsGUIScriptFile := filepath.Join(scriptDir, windowsGUIScriptFileName)

	func() {
		file, err := rt.FS().Create(windowsGUIScriptFile)
		if err != nil {
			rt.Logger().Error().Err(err).Str("file", windowsGUIScriptFile).Msg("Failed to create Windows wrap script")
			return
		}
		defer file.Close()
		if _, err := file.WriteString(windowsGUIScriptContent); err != nil {
			rt.Logger().Error().Err(err).Str("file", windowsGUIScriptFile).Msg("Failed to write Windows wrap script")
		}
	}()

	// --- Unix用スクリプト ---
	unixScriptContent := GenUnixWrapScript()
	unixScriptFileName := baseName
	unixScriptFile := filepath.Join(scriptDir, unixScriptFileName)

	func() {
		file, err := rt.FS().Create(unixScriptFile)
		if err != nil {
			rt.Logger().Error().Err(err).Str("file", unixScriptFile).Msg("Failed to create Unix wrap script")
			return
		}
		defer file.Close()
		if _, err := file.WriteString(unixScriptContent); err != nil {
			rt.Logger().Error().Err(err).Str("file", unixScriptFile).Msg("Failed to write Unix wrap script")
		}
	}()
}

func GenUtilsScript(rt interfaces.Runtime) {
	homeDir, _ := rt.Env().UserHomeDir()
	if homeDir == "" {
		return
	}
	tagDir := filepath.Join(homeDir, types.DEFAULT_TAG_DIR)
	tagFileLists, _ := general.GetFilesWithExtension(tagDir, ".tag")
	tagNameLists := general.RemoveExtensions(general.BaseSlice(tagFileLists))
	rt.Logger().Debug().
		Str("tagNameLists", strings.Join(tagNameLists, ",")).Msg("")

	for _, name := range tagNameLists {
		baseName := "t" + name
		rt.Logger().Debug().
			Str("tagName", name).Msg("")

		tagData, err := types.ReadTagData(rt, name)
		if err != nil {
			rt.Logger().Error().Err(err).Str("tag", name).Msg("Failed to read tag data")
			continue
		}

		configFile := tagData.ConfigFile
		var config types.Config
		if configFile != "" && general.FileExists(configFile) {
			config, err = types.ReadConfig(rt, configFile)
			if err != nil {
				rt.Logger().Error().Err(err).Str("configFile", configFile).Msg("Failed to read config file")
			}
		}

		var tagUtilsScriptContent string
		var tagUtilsScriptFileName string

		if rt.Env().GOOS() == "windows" {
			hasProgram := tagData.Program != "" || config.Program.Path != ""
			hasProgramArgs := len(tagData.ProgramArgs) != 0 || len(config.Program.Args) != 0

			switch {
			case hasProgram:
				tagUtilsScriptContent = GenWindowsSimpleTagUtilsScript(name)

			case !hasProgram && hasProgramArgs:
				tagUtilsScriptContent = GenWindowsTagUtilsScriptProgramArgs(name)

			case !hasProgram && !hasProgramArgs:
				if len(config.Program.Args) != 0 {
					tagUtilsScriptContent = GenWindowsSimpleTagUtilsScript(name)
				} else {
					tagUtilsScriptContent = GenWindowsTagUtilsScriptProgram(name)
				}

			default:
				tagUtilsScriptContent = GenWindowsSimpleTagUtilsScript(name)
			}

			tagUtilsScriptFileName = general.AddExtension(baseName, ".cmd")

		} else {
			tagUtilsScriptContent = GenUnixTagUtilsScript(name)
			tagUtilsScriptFileName = general.AddExtension(baseName, ".sh")
		}

		rt.Logger().Debug().
			Str("tagUtilsScriptContent", tagUtilsScriptContent).
			Str("tagUtilsScriptFileName", tagUtilsScriptFileName).
			Msg("")

		scriptDir := filepath.Join(types.DEFAULT_SCRIPT_DIR, types.DEFAULT_UTILS_SCRIPT_DIR)
		tagUtilsScriptFile := filepath.Join(scriptDir, tagUtilsScriptFileName)
		rt.Logger().Debug().
			Str("tagUtilsScriptFile", tagUtilsScriptFile).
			Msg("")

		if err := rt.FS().MkdirAll(scriptDir, 0755); err != nil {
			rt.Logger().Error().Err(err).Msg("Failed to create utils script directory")
			return
		}

		file, err := rt.FS().Create(tagUtilsScriptFile)
		if err != nil {
			rt.Logger().Error().Err(err).Str("file", tagUtilsScriptFile).Msg("Failed to create file")
			return
		}

		func() {
			defer file.Close()
			_, err := file.WriteString(tagUtilsScriptContent)
			if err != nil {
				rt.Logger().Error().Err(err).Msg("Failed to write to file")
			}
		}()
	}
}
