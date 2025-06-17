@echo off
chcp 65001
echo タグを設定します...
set "USER_EEC_CONFIG_DIR=D:\win\program\go\main-project\eec\eec-configs"
REM call eec tag add powershell00 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml" --program "powershell"  --program-args="-NoExit","-Command","Set-ExecutionPolicy RemoteSigned -Scope Process; checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv"
REM call eec tag add powershell01 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml" --program "powershell"  --program-args="-NoExit","-Command","Set-ExecutionPolicy RemoteSigned -Scope Process; checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv; Set-Location -Path 'D:\win\program\'"
REM call eec tag add cmd00 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml"  --program "cmd" --program-args="/K checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv"
REM call eec tag add cmd01 --config-file "%USER_EEC_CONFIG_DIR%\eec-config.toml" --program "cmd" --program-args="/K cd D:\win\program && D: && checkitems %USER_EEC_CONFIG_DIR%\checkitems.csv"

call eec tag add android-studio00 --config-file "%USER_EEC_CONFIG_DIR%\android-studio.toml" --program "D:\win\dev-tools\android\android-studio\bin\studio64"
call eec tag add easy-up --program "D:\win\program\go\main-project\google-drive-easy-uploader\build\easy-up"
call eec tag add go-dev --config-file "%USER_EEC_CONFIG_DIR%\go-dev.toml" --import "%USER_EEC_CONFIG_DIR%\base-dev.toml" 
call eec tag add dev --import "%USER_EEC_CONFIG_DIR%\base-dev.toml" --import "%USER_EEC_CONFIG_DIR%\go-dev.toml" --import "%USER_EEC_CONFIG_DIR%\rust-dev.toml" --import "%USER_EEC_CONFIG_DIR%\java-dev.toml" --import "%USER_EEC_CONFIG_DIR%\mingw-dev.toml" 


call eec tag add dev-cmd00 --import dev --program cmd 



echo タグの設定が終了しました
:: キー入力を待機
pause
