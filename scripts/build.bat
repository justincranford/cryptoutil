@echo off
REM Docker Build Script with Mandatory Args Validation
REM Usage: build.bat [APP_VERSION]

if "%~1"=="" (
    echo ERROR: APP_VERSION is required as first argument
    echo Usage: build.bat ^<APP_VERSION^>
    echo Example: build.bat v1.0.0
    exit /b 1
)

set APP_VERSION=%~1

echo Building cryptoutil with version: %APP_VERSION%
echo VCS_REF will be set to current commit hash
echo BUILD_DATE will be set to current timestamp

docker build ^
  --build-arg APP_VERSION=%APP_VERSION% ^
  --build-arg VCS_REF=%CI_COMMIT_SHA% ^
  --build-arg BUILD_DATE=%CI_BUILD_DATE% ^
  -t cryptoutil:%APP_VERSION% ^
  -f deployments/Dockerfile .

if %ERRORLEVEL% EQU 0 (
    echo SUCCESS: Docker image built as cryptoutil:%APP_VERSION%
) else (
    echo ERROR: Docker build failed
    exit /b 1
)
