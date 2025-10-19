#!/bin/bash

echo "🚀 Starting File Store API Setup..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or higher."
    exit 1
fi

echo "✅ Go is installed"

# Check if PostgreSQL is running
if ! command -v psql &> /dev/null; then
    echo "⚠️  PostgreSQL client not found. Make sure PostgreSQL is installed and running."
fi

# Copy environment file if it doesn't exist
if [ ! -f .env ]; then
    echo "📋 Copying .env.example to .env..."
    cp .env.example .env
    echo "⚠️  Please edit .env file with your database and email configuration"
else
    echo "✅ .env file already exists"
fi

# Download dependencies
echo "📦 Downloading Go dependencies..."
go mod tidy

if [ $? -eq 0 ]; then
    echo "✅ Dependencies downloaded successfully"
else
    echo "❌ Failed to download dependencies"
    exit 1
fi

# Build the application
echo "🔨 Building the application..."
go build -o bin/filestore cmd/main.go

if [ $? -eq 0 ]; then
    echo "✅ Application built successfully"
else
    echo "❌ Failed to build application"
    exit 1
fi

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "To start the server:"
echo "  ./bin/filestore"
echo ""
echo "Or run in development mode:"
echo "  go run cmd/main.go"
echo ""
echo "The API will be available at: http://localhost:8080"
echo "Health check: http://localhost:8080/health"
echo ""
echo "Don't forget to:"
echo "1. Configure your .env file"
echo "2. Start PostgreSQL database"
echo "3. Update database connection settings"
