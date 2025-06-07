#!/usr/bin/env bash
go mod tidy
go build -o build/eec main.go
