/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/

package cmd

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"main/ext"
	"os"
	"path/filepath"
	"strings"
)

type TagData struct {
	ConfigFile  string
	Program     string
	ProgramArgs []string
	ImportConfigFiles []string
	//Description string TODO: 要追加検討
}

// ---------------------------
// TagData バイナリ保存処理
// ---------------------------
func (t *TagData) Write(tagName string) error {
	homeDir, err := os.UserHomeDir()
	if homeDir == "" {
		return err
	}
	dir := filepath.Join(homeDir, ext.DEFAULT_TAG_DIR)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tagPath := filepath.Join(dir, fmt.Sprintf("%s.tag", tagName))

	var buf bytes.Buffer

	// 各フィールドを手動で書き込む
	if err := writeString(&buf, t.ConfigFile); err != nil {
		return err
	}
	if err := writeString(&buf, t.Program); err != nil {
		return err
	}
	if err := writeStringSlice(&buf, t.ProgramArgs); err != nil {
		return err
	}
	if err := writeStringSlice(&buf, t.ImportConfigFiles); err != nil {
		return err
	}

	if err := os.WriteFile(tagPath, buf.Bytes(), 0644); err != nil {
		return err
	}

	log.Info().
		Str("tagPath", tagPath).
		Str("ConfigFile", t.ConfigFile).
		Str("Program", t.Program).
		Str("Args", strings.Join(t.ProgramArgs, ", ")).
		Str("Import config files", strings.Join(t.ImportConfigFiles, ", ")).
		Msg("TagData written successfully")

	return nil
}

// --------------------------
// 読み取り処理
// --------------------------
func ReadTagData(tagName string) (TagData, error) {
	homeDir, _ := os.UserHomeDir()
	if homeDir == "" {
		return TagData{}, fmt.Errorf(fmt.Sprintf("%s not set", homeDir))
	}
	tagPath := filepath.Join(homeDir, ".eec", fmt.Sprintf("%s.tag", tagName))

	content, err := os.ReadFile(tagPath)
	if err != nil {
		return TagData{}, err
	}

	buf := bytes.NewReader(content)

	var data TagData
	if data.ConfigFile, err = readString(buf); err != nil {
		return TagData{}, err
	}
	if data.Program, err = readString(buf); err != nil {
		return TagData{}, err
	}
	if data.ProgramArgs, err = readStringSlice(buf); err != nil {
		return TagData{}, err
	}
	
	if data.ImportConfigFiles, err = readStringSlice(buf); err != nil {
		return TagData{}, err
	}

	log.Info().
		Str("ConfigFile", data.ConfigFile).
		Str("Program", data.Program).
		Str("Args", strings.Join(data.ProgramArgs, " ")).
		Str("Import config files", strings.Join(data.ImportConfigFiles, ", ")).
		Msg("TagData read successfully")

	return data, nil
}

// ---------------------------
// ヘルパー関数
// ---------------------------
func writeString(buf *bytes.Buffer, s string) error {
	length := int32(len(s))
	if err := binary.Write(buf, binary.LittleEndian, length); err != nil {
		return err
	}
	return binary.Write(buf, binary.LittleEndian, []byte(s))
}

func writeStringSlice(buf *bytes.Buffer, ss []string) error {
	count := int32(len(ss))
	if err := binary.Write(buf, binary.LittleEndian, count); err != nil {
		return err
	}
	for _, s := range ss {
		if err := writeString(buf, s); err != nil {
			return err
		}
	}
	return nil
}

func readString(buf *bytes.Reader) (string, error) {
	var length int32
	if err := binary.Read(buf, binary.LittleEndian, &length); err != nil {
		return "", err
	}
	bytes := make([]byte, length)
	if err := binary.Read(buf, binary.LittleEndian, &bytes); err != nil {
		return "", err
	}
	return string(bytes), nil
}

func readStringSlice(buf *bytes.Reader) ([]string, error) {
	var count int32
	if err := binary.Read(buf, binary.LittleEndian, &count); err != nil {
		return nil, err
	}
	result := make([]string, 0, count)
	for i := int32(0); i < count; i++ {
		s, err := readString(buf)
		if err != nil {
			return nil, err
		}
		result = append(result, s)
	}
	return result, nil
}

// ---------------------------
// 位置引数やフラグ 格納
// ---------------------------
var (
	//tagNameTagFlag     string
	configFileTagFlag  string
	programTagFlag     string
	programArgsTagFlag []string
	importConfigFilesTagFlag []string
)

// ---------------------------
// 指定ディレクトリ内の特定拡張子ファイルを取得する関数
// ---------------------------
func GetFilesWithExtension(dir string, ext string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.Type().IsRegular() && strings.EqualFold(filepath.Ext(entry.Name()), ext) {
			fullPath := filepath.Join(dir, entry.Name())
			files = append(files, fullPath)
		}
	}

	return files, nil
}

// ---------------------------
// Cobra コマンド定義
// ---------------------------
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
}

var addTagCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new tag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tagName := args[0]
		// -- デバッグ用 --
		log.Debug().
			Str("tagName", tagName).
			Msg("")
		log.Debug().
			Str("configFileFlag", configFileTagFlag).
			Msg("")
		log.Debug().
			Str("programFlag", programTagFlag).
			Msg("")
		log.Debug().
			Str("programArgsFlag", strings.Join(programArgsTagFlag, ", ")).
			Msg("")
		log.Debug().
			Str("Import config files", strings.Join(importConfigFilesTagFlag, ", ")).
			Msg("")	
		//

		data := TagData{
			ConfigFile:  configFileTagFlag,
			Program:     programTagFlag,
			ProgramArgs: programArgsTagFlag,
			ImportConfigFiles: importConfigFilesTagFlag,
		}
		if err := data.Write(tagName); err != nil {
			log.Error().Err(err).Msg("タグファイルの書き込みに失敗しました")
			os.Exit(1)
		}
		fmt.Println("Tag added:", tagName)
	},
}

var readTagCmd = &cobra.Command{
	Use:   "read [name]",
	Short: "Read a tag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tagName := args[0]
		data, err := ReadTagData(tagName)
		if err != nil {
			log.Error().Err(err).Msg("タグファイルの読み込みに失敗しました")
			os.Exit(1)
		}
		fmt.Printf("Tag: %s\n  Config: %s\n  Program: %s\n  Args: %v\n  Import config files: %v\n",
			tagName, data.ConfigFile, data.Program, data.ProgramArgs,data.ImportConfigFiles)
	},
}
var listTagCmd = &cobra.Command{
	Use:   "list",
	Short: "List a tags",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		if homeDir == "" {
			log.Error().Err(err).Msg(fmt.Sprintf("homeDir(%s)が設定されていません", homeDir))
			os.Exit(1)
		}
		tagDir := filepath.Join(homeDir, ext.DEFAULT_TAG_DIR)
		fileLists, err := GetFilesWithExtension(tagDir, ".tag")
		if err != nil {
			log.Error().Err(err).Msg("タグファイルが見つかりませんでした")
			os.Exit(1)
		}
		fmt.Printf("-- current tag lists  --\n")
		fmt.Printf("%s\n", strings.Join(fileLists, "\n"))
	},
}

var removeTagCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a tag",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		tagName := args[0]
		homeDir, err := os.UserHomeDir()
		if homeDir == "" {
			log.Error().Err(err).Msg(fmt.Sprintf("homeDir(%s)が設定されていません", homeDir))
			os.Exit(1)
		}
		tagDir := filepath.Join(homeDir, ext.DEFAULT_TAG_DIR)
		tagPath := filepath.Join(tagDir, fmt.Sprintf("%s.tag", tagName))
		err = os.Remove(tagPath)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("タグ名: %sの削除に失敗しました", tagName))
			os.Exit(1)
		}
		fmt.Printf("タグ名: %sを削除しました\n", tagName)
	
		fileLists, err := GetFilesWithExtension(tagDir, ".tag")
		if err != nil {
			log.Error().Err(err).Msg("タグファイルが見つかりませんでした")
			os.Exit(1)
		}
		fmt.Printf("-- current tag lists  --\n")
		fmt.Printf("%s\n", strings.Join(fileLists, "\n"))

	},
}

func init() {
	//addTagCmd.Flags().StringVar(&tagNameTagFlag, "name", "", "Tag name")
	addTagCmd.Flags().StringVar(&configFileTagFlag, "config-file", "", "Config file")
	addTagCmd.Flags().StringVar(&programTagFlag, "program", "", "Program name")
	addTagCmd.Flags().StringSliceVar(&programArgsTagFlag, "program-args", []string{}, "Program args")
	addTagCmd.Flags().StringSliceVar(&importConfigFilesTagFlag, "import", []string{}, "Import config files")

	tagCmd.AddCommand(addTagCmd)
	tagCmd.AddCommand(readTagCmd)
	tagCmd.AddCommand(listTagCmd)
	tagCmd.AddCommand(removeTagCmd)
	rootCmd.AddCommand(tagCmd)
}
