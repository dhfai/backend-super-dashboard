#!/bin/bash

# Quick setup script for Postman Collection
# This script helps setup File Store API testing environment

echo "🚀 File Store API - Postman Setup Guide"
echo "========================================"
echo ""

echo "📁 Files tersedia:"
echo "   - postman/postman-collection.json (API Collection)"
echo "   - postman/filestore-environment.json (Environment Variables)"
echo "   - postman/README.md (Detailed Setup Guide)"
echo ""

echo "🔧 Setup Steps:"
echo "   1. Buka Postman (Desktop atau Web)"
echo "   2. Click 'Import' button"
echo "   3. Drag & drop kedua .json files:"
echo "      - postman-collection.json"
echo "      - filestore-environment.json"
echo "   4. Pilih 'File Store API Environment' dari dropdown"
echo "   5. Edit environment variables sesuai kebutuhan"
echo ""

echo "🎯 Quick Test Workflow:"
echo "   1. Start server: ./scripts/run.sh"
echo "   2. Register → Creates account (unverified)"
echo "   3. Verify Email → Verifies with OTP"
echo "   4. Login → Auto-saves JWT token"
echo "   5. Test protected endpoints → Uses saved token"
echo ""

echo "📊 Available Test Collections:"
echo "   - Health Check (Public)"
echo "   - Authentication Flow (Register, Login, Verify, etc.)"
echo "   - Profile Management (Get, Update, Delete)"
echo "   - Account Management (Logout, Delete Account)"
echo ""

echo "🔐 Security Features:"
echo "   - Auto token management"
echo "   - Email verification required"
echo "   - OTP confirmation for account deletion"
echo "   - Token blacklist untuk logout"
echo ""

echo "📋 Environment Variables:"
echo "   - baseUrl: http://localhost:8080 (change for production)"
echo "   - testEmail: test@example.com"
echo "   - testPassword: SecurePass123!"
echo "   - testUsername: testuser"
echo "   - otpCode: 123456 (for development)"
echo ""

echo "✨ Features:"
echo "   ✅ Complete API coverage"
echo "   ✅ Auto token save/clear"
echo "   ✅ Environment variables"
echo "   ✅ Test automation scripts"
echo "   ✅ Error handling examples"
echo ""

echo "📖 For detailed instructions, see: postman/README.md"
echo ""
echo "Happy Testing! 🚀"
