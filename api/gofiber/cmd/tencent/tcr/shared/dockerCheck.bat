@echo off

:: Try to run a simple docker command
docker ps >nul 2>&1

:: Check the exit code of the previous command
if %errorlevel% neq 0 (
    echo [ERROR] Docker is not running or the daemon is unreachable.
    exit /b 1
) else (
    echo [SUCCESS] Docker is up and running.
    exit /b 0
)