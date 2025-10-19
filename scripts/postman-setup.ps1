# File Store API - Postman Setup Guide (PowerShell)
# Quick setup script untuk Postman Collection

Write-Host "🚀 File Store API - Postman Setup Guide" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""

Write-Host "📁 Files tersedia:" -ForegroundColor Yellow
Write-Host "   - postman/postman-collection.json (API Collection)" -ForegroundColor White
Write-Host "   - postman/filestore-environment.json (Environment Variables)" -ForegroundColor White
Write-Host "   - postman/README.md (Detailed Setup Guide)" -ForegroundColor White
Write-Host ""

Write-Host "🔧 Setup Steps:" -ForegroundColor Yellow
Write-Host "   1. Buka Postman (Desktop atau Web)" -ForegroundColor White
Write-Host "   2. Click 'Import' button" -ForegroundColor White
Write-Host "   3. Drag & drop kedua .json files:" -ForegroundColor White
Write-Host "      - postman-collection.json" -ForegroundColor Gray
Write-Host "      - filestore-environment.json" -ForegroundColor Gray
Write-Host "   4. Pilih 'File Store API Environment' dari dropdown" -ForegroundColor White
Write-Host "   5. Edit environment variables sesuai kebutuhan" -ForegroundColor White
Write-Host ""

Write-Host "🎯 Quick Test Workflow:" -ForegroundColor Yellow
Write-Host "   1. Start server: .\bin\app.exe" -ForegroundColor White
Write-Host "   2. Register → Creates account (unverified)" -ForegroundColor White
Write-Host "   3. Verify Email → Verifies with OTP" -ForegroundColor White
Write-Host "   4. Login → Auto-saves JWT token" -ForegroundColor White
Write-Host "   5. Test protected endpoints → Uses saved token" -ForegroundColor White
Write-Host ""

Write-Host "📊 Available Test Collections:" -ForegroundColor Yellow
Write-Host "   - Health Check (Public)" -ForegroundColor White
Write-Host "   - Authentication Flow (Register, Login, Verify, etc.)" -ForegroundColor White
Write-Host "   - Profile Management (Get, Update, Delete)" -ForegroundColor White
Write-Host "   - Account Management (Logout, Delete Account)" -ForegroundColor White
Write-Host ""

Write-Host "🔐 Security Features:" -ForegroundColor Yellow
Write-Host "   - Auto token management" -ForegroundColor White
Write-Host "   - Email verification required" -ForegroundColor White
Write-Host "   - OTP confirmation for account deletion" -ForegroundColor White
Write-Host "   - Token blacklist untuk logout" -ForegroundColor White
Write-Host ""

Write-Host "📋 Environment Variables:" -ForegroundColor Yellow
Write-Host "   - baseUrl: http://localhost:8080 (change for production)" -ForegroundColor White
Write-Host "   - testEmail: test@example.com" -ForegroundColor White
Write-Host "   - testPassword: SecurePass123!" -ForegroundColor White
Write-Host "   - testUsername: testuser" -ForegroundColor White
Write-Host "   - otpCode: 123456 (for development)" -ForegroundColor White
Write-Host ""

Write-Host "✨ Features:" -ForegroundColor Yellow
Write-Host "   ✅ Complete API coverage" -ForegroundColor Green
Write-Host "   ✅ Auto token save/clear" -ForegroundColor Green
Write-Host "   ✅ Environment variables" -ForegroundColor Green
Write-Host "   ✅ Test automation scripts" -ForegroundColor Green
Write-Host "   ✅ Error handling examples" -ForegroundColor Green
Write-Host ""

Write-Host "📖 For detailed instructions, see: postman/README.md" -ForegroundColor Cyan
Write-Host ""
Write-Host "Happy Testing! 🚀" -ForegroundColor Green

# Optional: Open postman folder
$openFolder = Read-Host "`nOpen postman folder? (y/n)"
if ($openFolder -eq 'y' -or $openFolder -eq 'Y') {
    Start-Process explorer.exe -ArgumentList "postman"
}
