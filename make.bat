@echo OFF

set "args=%*"
pushd "%~dp0"
setlocal ENABLEDELAYEDEXPANSION
set GOPATH="%~dp0vendor"
rem Set the GOPROXY environment variable
set GOPROXY=https://goproxy.io

if /i "%args%"=="install" goto install
if /i "%args%"=="all" goto all
if /i "%args%"=="run" goto run

goto DEFAULT_CASE
:install
    mkdir vendor
    CALL go mod tidy
    GOTO END_CASE
:all
    echo ========================
    echo build
    CALL go build -ldflags "-s -w" -o "bin/windows_amd64/laravel-echo-server.exe" main.go

    GOTO END_CASE
:run
    CALL go build -o bin\main.exe main.go && CALL %~dp0\bin\main.exe
    GOTO END_CASE
:DEFAULT_CASE
    CALL go mod tidy
    CALL go build -ldflags "-s -w" -o bin\main.exe main.go
    GOTO END_CASE
:END_CASE
    GOTO :EOF