/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import "main/cmd"
import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}
func main() {
	cmd.Execute()
}
