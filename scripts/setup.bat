@echo off
echo 🚀 Starting File Store API Setup...

:: Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go is not installed. Please install Go 1.21 or higher.
    exit /b 1
)
echo ✅ Go is installed

:: Copy environment file if it doesn't exist
if not exist ".env" (
    echo 📋 Copying .env.example to .env...
    copy .env.example .env
    echo ⚠️  Please edit .env file with your database and email configuration
) else (
    echo ✅ .env file already exists
)

:: Download dependencies
echo 📦 Downloading Go dependencies...
go mod tidy

if %errorlevel% equ 0 (
    echo ✅ Dependencies downloaded successfully
) else (
    echo ❌ Failed to download dependencies
    exit /b 1
)

:: Create bin directory if it doesn't exist
if not exist "bin" mkdir bin

:: Build the application
echo 🔨 Building the application...
go build -o bin/filestore.exe cmd/main.go

if %errorlevel% equ 0 (
    echo ✅ Application built successfully
) else (
    echo ❌ Failed to build application
    exit /b 1
)

echo.
echo 🎉 Setup completed successfully!
echo.
echo To start the server:
echo   bin\filestore.exe
echo.
echo Or run in development mode:
echo   go run cmd/main.go
echo.
echo The API will be available at: http://localhost:8080
echo Health check: http://localhost:8080/health
echo.
echo Don't forget to:
echo 1. Configure your .env file
echo 2. Start PostgreSQL database
echo 3. Update database connection settings

pause
