@echo off
set "eec_deleter=eec-deleter.exe"
set "eec_exe=D:\win\program\go\main-project\eec\build\eec"
tasklist /FI "IMAGENAME eq %eec_deleter%" /NH | find /I "%eec_deleter%" >nul

if "%1"=="run" (
    rem エラーレベルが 0（すでに実行中）か確認
    if %ERRORLEVEL% equ 0 (
        echo [%eec_deleter%] は既に実行中です。
    ) else (
        echo [%eec_deleter%] を起動します…
        powershell -WindowStyle Normal -Command "Start-Process -FilePath '%eec_deleter%' -WindowStyle Hidden"
    )
)
%eec_exe% %*
