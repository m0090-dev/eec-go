/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// ====================
// プロセスが終了するのを待機する関数
// ====================
func waitForProcessTermination(pid int) error {
	for {
		var cmd *exec.Cmd

		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))
		default:
			cmd = exec.Command("ps", "-p", strconv.Itoa(pid))
		}

		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to check process: %w", err)
		}

		if strings.Contains(string(output), strconv.Itoa(pid)) {
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}
	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "main",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
	Run: func(cmd *cobra.Command, args []string) {
		// 一時ディレクトリのパスを取得
		tempDir := os.TempDir()
		manifestPath := filepath.Join(tempDir, "eec_manifest.txt")

		for {
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				log.Error().
					Err(err).
					Msg("eec_manifest.txt does not exist. Skipping...")
				time.Sleep(3 * time.Second)
				continue
			}

			file, err := os.Open(manifestPath)
			if err != nil {
				log.Error().
					Err(err).
					Msg("Failed to open manifest")
				time.Sleep(3 * time.Second)
				continue
			}
			scanner := bufio.NewScanner(file)
			deletedAll := true

			for scanner.Scan() {
				line := scanner.Text()
				parts := strings.Fields(line)
				if len(parts) != 2 {

					log.Error().
						Err(err).
						Str("line", line).
						Msg("Invalid line in manifest")

					continue
				}

				tempFilePath := parts[0]
				pid, err := strconv.Atoi(parts[1])
				if err != nil || pid == 0 {

					log.Error().
						Err(err).
						Int("pid", pid).
						Msg("PID is invalid. Skipping...")

					deletedAll = false
					continue
				}

				if err := waitForProcessTermination(pid); err != nil {
					log.Error().
						Err(err).
						Msg("Failed waiting for process")

					deletedAll = false
					continue
				}

				if _, err := os.Stat(tempFilePath); err == nil {
					if err := os.Remove(tempFilePath); err != nil {
						log.Error().
							Err(err).
							Str("tempFilePath", tempFilePath).
							Msg("Failed to delete temp file")

						deletedAll = false
					} else {
						log.Info().
							Msg("Deleted temp file")
					}
				} else {
					log.Error().
						Err(err).
						Str("tempFilePath", tempFilePath).
						Msg("Temp file does not exist. Skipping...")

					deletedAll = false
				}
			}

			file.Close()
			if deletedAll {
				if err := os.Remove(manifestPath); err != nil {
					log.Error().
						Err(err).
						Msg("Failed to delete manifest file")
				} else {
						
					log.Info().
						
						Msg("Deleted manifest file")

				}
				break
			}

			time.Sleep(5 * time.Second)
		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.main.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
