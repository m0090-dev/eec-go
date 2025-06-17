@echo off
chcp 65001
eec run --config-file "D:/win/program/go/main-project/eec/eec-configs/eec-config.toml" --program cmd --program-args "/C %1"