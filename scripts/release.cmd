@echo off
setlocal enabledelayedexpansion

:: --- デフォルト設定 ---
set BRANCH=beta/stable
set MESSAGE=chore: update version and push
set SKIP_CI=false

:: --- 引数の解析 ---
:parse_args
if "%~1"=="" goto finalize_message
if /i "%~1"=="--branch" (
    set BRANCH=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--message" (
    set MESSAGE=%~2
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--skip-ci" (
    set SKIP_CI=true
    shift
    goto parse_args
)
shift
goto parse_args

:finalize_message
:: skip-ciフラグが立っている場合、メッセージの末尾に付け加える
if "%SKIP_CI%"=="true" (
    set MESSAGE=%MESSAGE% [skip ci]
)

:run_scripts
echo [1/4] Running version update script...
python scripts/update_version.py
if %ERRORLEVEL% neq 0 (
    echo Error: update_version.py failed.
    exit /b %ERRORLEVEL%
)

echo [2/4] Git add...
git add .

echo [3/4] Git commit with message: "%MESSAGE%"...
git commit -m "%MESSAGE%"
if %ERRORLEVEL% neq 0 (
    echo No changes to commit or git error.
    goto push_stage
)

:push_stage
echo [4/4] Pushing to %BRANCH%...
git push origin %BRANCH%

echo.
if "%SKIP_CI%"=="true" (
    echo Done! (Actions was skipped as requested)
) else (
    echo Done! Factory is now working on GitHub Actions.
)
pause
