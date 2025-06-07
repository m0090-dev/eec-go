@echo off
chcp 65001
echo タグを設定します...
set "USER_EEC_CONFIG_DIR=D:\win\program\go\main-project\eec\eec-configs"
call eec tag add powershell00 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml" --program "powershell"  --program-args="-NoExit","-Command","Set-ExecutionPolicy RemoteSigned -Scope Process; checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv"
call eec tag add powershell01 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml" --program "powershell"  --program-args="-NoExit","-Command","Set-ExecutionPolicy RemoteSigned -Scope Process; checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv; Set-Location -Path 'D:\win\program\'"
call eec tag add cmd00 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml"  --program "cmd" --program-args="/K checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv"
call eec tag add cmd01 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml" --program "cmd" --program-args="/K cd D:\win\program && D: && checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv"

echo タグの設定が終了しました
:: キー入力を待機
pause
