/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"main/cmd/meta"
	"main/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ---------------------------
// 位置引数やフラグ 格納
// ---------------------------
var configFileRunFlag string
var programRunFlag string
var programArgsRunFlag []string
var tagRunFlag string

// ---------------------------
// 拡張子を除いたファイル名を返す関数
// ---------------------------
func RemoveExtension(filename string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext)
}

// ---------------------------
// Cobra コマンド定義
// ---------------------------
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		// -- デバッグ用 --
		log.Debug().
			Str("configFileRunFlag", configFileRunFlag).
			Msg("")
		log.Debug().
			Str("programRunFlag", programRunFlag).
			Msg("")
		log.Debug().
			Str("programArgsRunFlag", strings.Join(programArgsRunFlag, ", ")).
			Msg("")
		//

		selfProgram := os.Args[0]
		tempData := meta.TempData{}

		// -- 各自処理時の優先順位(タグデータ > コマンドライン引数) --
		var tagData TagData
		var configFile string
		var program string
		var pArgs []string

		if tagRunFlag == "" {
			configFile = configFileRunFlag
			program = programRunFlag
			pArgs = programArgsRunFlag
		} else {
			var err error
			tagData, err = ReadTagData(tagRunFlag)
			if err != nil {
				log.Error().
					Err(err).
					Str("tagRunFlag", tagRunFlag).
					Msg("タグデータの読み込みに失敗しました")
				os.Exit(1)
			}
			configFile = tagData.ConfigFile
			program = tagData.Program
			pArgs = tagData.ProgramArgs
		}

		tmpDir := os.TempDir()
		tmpPrefix := fmt.Sprintf(
			"%s_%s_%s.tmp",
			RemoveExtension(filepath.Base(selfProgram)),
			RemoveExtension(filepath.Base(program)),
			uuid.New().String(),
		)

		//fmt.Printf("configFile=%s\n",configFile)
		tmpPath := filepath.Join(tmpDir, tmpPrefix)
		tmpFile, err := os.Create(tmpPath)

		if err != nil {
			log.Error().
				Err(err).
				Str("prefix", tmpPrefix).
				Msg("一時ファイルの作成に失敗しました")
			os.Exit(1)

		}
		log.Info().
			Str("Temp file", tmpPath).
			Msg("Created temp file")

		var config meta.Config
		if configFile != "" && utils.FileExists(configFile) {
			config, err = meta.ReadConfig(configFile)
		} else {
			config, err = meta.ReadInlineConfig(configFile)
		}

		if err != nil {
			log.Error().
				Err(err).
				Str("configFile", configFile).
				Msg("tomlファイルの読み込みに失敗しました")
		}

		manifest := meta.Manifest{
			TempFilePath: tmpFile.Name(),
			EECPID:       os.Getpid(),
		}

		manifestPath, err := manifest.WriteToManifest()
		if err != nil {
			log.Error().
				Err(err).
				Str("manifestPath", manifestPath).
				Msg("マニフェストファイルの作成に失敗しました")
			os.Exit(1)
		}

		log.Info().
			Str("Manifest file", manifestPath).
			Msg("Created manifest file")

		//currentPaths := strings.Split(paths, separator)
		//newPath := currentPaths

		/* for _, envVar := range config.Envs {*/
		/*if len(envVar) >= 2 {*/
		/*// キー部分（文字列にキャスト）*/
		/*key, ok := envVar[0].(string)*/
		/*if !ok {*/
		/*log.Warn().Interface("envVar", envVar).Msg("envのキーが文字列ではありません")*/
		/*continue*/
		/*}*/

		/*// 値部分の型をチェック*/
		/*switch val := envVar[1].(type) {*/
		/*case string:*/
		/*// スカラー文字列*/
		/*expanded := utils.ExpandEnvVariables(val)*/
		/*os.Setenv(key, expanded)*/

		/*case []interface{}:*/
		/*// 文字列配列（interface{}として来る）*/
		/*strVals := make([]string, 0, len(val))*/
		/*for _, v := range val {*/
		/*if s, ok := v.(string); ok {*/
		/*strVals = append(strVals, utils.ExpandEnvVariables(s))*/
		/*}*/
		/*}*/
		/*os.Setenv(key, strings.Join(strVals, separator)) // セミコロン or コロン*/
		/*default:*/
		/*log.Warn().*/
		/*Str("key", key).*/
		/*Interface("value", envVar[1]).*/
		/*Msg("envの値の型が未対応")*/
		/*}*/
		/*}*/
		/*}*/

		/* for _, path := range config.Paths {*/
		/*expandedPath := utils.ExpandEnvVariables(path)*/
		/*if expandedPath != "" {*/
		/*newPath = append(newPath, expandedPath)*/
		/*}*/
		/*}*/
		/*os.Setenv("Path", strings.Join(newPath, separator))*/

		config.ApplyEnvs()

		executeCommand := exec.Command(program, pArgs...)
		executeCommand.Stdin = os.Stdin
		executeCommand.Stdout = os.Stdout
		executeCommand.Stderr = os.Stderr
		err = executeCommand.Start()
		if err != nil {
			log.Error().
				Err(err).
				Msg("プログラムの起動に失敗しました")
			os.Exit(1)
		}
		childPID := executeCommand.Process.Pid

		log.Info().
			Int("PID", childPID).
			Msg("Sub process started ppid")

		// -- 一時ファイル書き込みデータセット --
		tempData.ParentPID = os.Getpid()
		tempData.ChildPID = childPID
		tempData.ConfigFile = configFile
		tempData.Program = program
		tempData.ProgramArgs = pArgs

		// -- 一時ファイル書き込みデータのバイナリエンコード --
		var tempDataBin bytes.Buffer
		encoder := gob.NewEncoder(&tempDataBin)
		if err = encoder.Encode(tempData); err != nil {
			log.Error().
				Err(err).
				Msg("一時ファイル使用データのエンコードに失敗しました")
			os.Exit(1)

		}
		_, err = tmpFile.Write(tempDataBin.Bytes())
		if err != nil {
			log.Error().
				Err(err).
				Msg("一時ファイルの書き込みに失敗しました")
			os.Exit(1)

		}

		log.Info().
			Int("ParentPID", tempData.ParentPID).
			Int("ChildPID", tempData.ChildPID).
			Str("ConfigFile", tempData.ConfigFile).
			Str("Program", tempData.Program).
			Str("Program Args", strings.Join(tempData.ProgramArgs, ", ")).
			Msg("Temp file written successfully")

		err = executeCommand.Wait()
		if err != nil {
			log.Error().
				Err(err).
				Msg("プログラム終了時にエラーが発生しました")
			os.Exit(1)
		}
		// TODO: env-exec-deleter(eec-deleter)に委託するため
		/*defer os.Remove(tmpFile.Name())*/
		/*defer tmpFile.Close()*/
		/*log.Info().*/
		/*Str("Temp file", tmpFile.Name()).*/
		/*Msg("Deleted temp file")*/

	},
}

func init() {
	runCmd.Flags().StringVar(&configFileRunFlag, "config-file", "", "Config file")

	runCmd.Flags().StringVar(&programRunFlag, "program", "", "Program name")
	runCmd.Flags().StringSliceVar(&programArgsRunFlag, "program-args", []string{}, "Program args")

	runCmd.Flags().StringVar(&tagRunFlag, "tag", "", "Tag name")

	rootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
