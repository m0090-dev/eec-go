@echo off
chcp 65001 > nul
setlocal enabledelayedexpansion

:: --- デフォルト設定 ---
set BRANCH=beta/stable
set MESSAGE=chore: update version and push
set SKIP_CI=false
set DO_PUBLISH=false

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
if /i "%~1"=="--publish" (
    set DO_PUBLISH=true
    shift
    goto parse_args
)
shift
goto parse_args

:finalize_message
REM if DO_PUBLISH is true, append [release] marker to the commit message
REM (auto-tag.yml detects this marker and creates a new tag automatically)
if "%DO_PUBLISH%"=="true" (
    set MESSAGE=%MESSAGE% [release]

    REM auto-tag.yml only watches pushes to beta/stable,
    REM so warn if --publish is used with a different branch
    if /i not "%BRANCH%"=="beta/stable" (
        echo.
        echo Warning: --publish が指定されていますが、push先が beta/stable ではありません 現在: %BRANCH%。
        echo          auto-tag.yml は beta/stable への push しか監視していないため
        echo          このままではタグの自動発行が行われません。
        echo.
    )
)

REM if SKIP_CI is true, append [skip ci] marker
if "%SKIP_CI%"=="true" (
    set MESSAGE=%MESSAGE% [skip ci]
)

:run_scripts

echo [1/5] Fetching latest tags from origin...
git fetch origin --tags --force

echo [2/5] Running version update script (local/cosmetic only)...
python scripts/update_version.py
if %ERRORLEVEL% neq 0 (
    echo Error: update_version.py failed.
    exit /b %ERRORLEVEL%
)

echo [3/5] Git add...
git add .

echo [4/5] Git commit with message: "%MESSAGE%"...
git commit -m "%MESSAGE%"
if %ERRORLEVEL% neq 0 (
    echo No changes to commit or git error.
    goto push_stage
)

:push_stage
echo [5/5] Pushing to %BRANCH%...
git push origin %BRANCH%

echo.
if "%DO_PUBLISH%"=="true" (
    echo Done! [release] marker included - auto-tag.yml will create a new tag if pushed to beta/stable.
    echo Note: the actual embedded VERSION in the built binary is now stamped by release.yml from the real tag, so it will always match.
) else (
    if "%SKIP_CI%"=="true" (
        echo Done! Actions was skipped as requested
    ) else (
        echo Done! Factory is now working on GitHub Actions.
    )
)
pause
