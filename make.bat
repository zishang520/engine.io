@echo OFF

set "args=%*"
pushd "%~dp0"
setlocal ENABLEDELAYEDEXPANSION
set GOPATH="%~dp0vendor"
rem Set the GOPROXY environment variable
set GOPROXY=https://goproxy.io

if /i "%args%"=="update" goto %args%
if /i "%args%"=="install" goto %args%
if /i "%args%"=="all" goto %args%
if /i "%args%"=="run" goto %args%
if /i "%args%"=="init" goto %args%

goto DEFAULT_CASE
:update
    if not exist vendor (
        CALL go mod vendor
    )
    CALL go mod tidy
    GOTO END_CASE
:install
    CALL go mod vendor
    GOTO END_CASE
:all
    echo ========================
    echo build
    CALL go build -ldflags "-s -w" -o "bin/windows_amd64/laravel-echo-server.exe" main.go

    GOTO END_CASE
:run
    CALL go run main.go
    GOTO END_CASE
:init
    GOTO END_CASE
:DEFAULT_CASE
    CALL go mod tidy
    CALL go build -ldflags "-s -w" -o bin\main.exe main.go
    GOTO END_CASE
:END_CASE
    GOTO :EOF