@echo off
title Atlas Dev Server

:: Always run from project root regardless of where this bat is invoked from
cd /d "%~dp0.."

:: Check for air, install if missing
where air >nul 2>&1
if %errorlevel% neq 0 (
    echo air not found - installing...
    go install github.com/air-verse/air@latest
    if %errorlevel% neq 0 (
        echo Failed to install air. Make sure Go is in your PATH.
        pause
        exit /b 1
    )
    echo air installed.
)

:: Initial template compile so the server builds on first run
echo Compiling templates...
templ generate
if %errorlevel% neq 0 (
    echo templ generate failed. Make sure templ is installed: go install github.com/a-h/templ/cmd/templ@latest
    pause
    exit /b 1
)

:: Start templ watcher in a separate window
start "Templ Watcher" cmd /k "templ generate --watch"

:: Run server with hot reload (air watches *.go, templ watcher feeds _templ.go changes into it)
echo Starting server with hot reload on http://localhost:8080
air
