/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

*/
package main

import "main/cmd"
import (
	"os"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)
func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
func main() {
	cmd.Execute()
}
