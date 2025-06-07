@echo off
go mod tidy
go build -o build\eec.exe main.go
