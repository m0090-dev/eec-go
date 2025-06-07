#!/usr/bin/env bash
go mod tidy
go build -o build\eec-deleter.exe main.go
