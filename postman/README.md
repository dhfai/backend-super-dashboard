# File Store API - Postman Collection

Collection ini berisi semua endpoint API untuk File Store backend dengan complete authentication dan profile management system.

## 📁 Files
- `postman-collection.json` - Complete Postman collection
- `filestore-environment.json` - Environment variables template

## 🔧 Setup

### 1. Import ke Postman
1. Buka Postman aplikasi atau web
2. Click "Import" button
3. Drag & drop kedua file `.json` atau browse to select
4. Collection dan Environment akan otomatis ter-import

### 2. Setup Environment
1. Pilih "File Store API Environment" dari dropdown environment
2. Edit variables sesuai kebutuhan:
   - `baseUrl`: Server URL (default: http://localhost:8080)
   - `testEmail`: Email untuk testing
   - `testPassword`: Password untuk testing
   - `testUsername`: Username untuk testing
   - `otpCode`: OTP code (default: 123456 untuk development)

### 3. Authentication Flow
1. **Register** → Creates new account (unverified)
2. **Verify Email** → Verifies email with OTP
3. **Login** → Generates JWT token (auto-saved ke variable `token`)
4. Token otomatis digunakan untuk authenticated endpoints

## 📊 Collection Structure

### 🔓 Public Endpoints
- **Health Check** - Server health status
- **Register** - User registration
- **Login** - User authentication (auto-saves token)
- **Verify Email** - Email verification with OTP
- **Resend Verification** - Resend verification email
- **Forget Password** - Request password reset
- **Reset Password** - Reset password with OTP

### 🔒 Protected Endpoints (Requires Token)
- **Logout** - Logout user (auto-clears token)
- **Request Delete Account** - Request account deletion with OTP
- **Delete Account** - Delete account with OTP (auto-clears token)
- **Get Profile** - Get user profile
- **Update Profile** - Update user profile
- **Delete Profile** - Delete user profile data
- **Get User Info** - Get basic user information

## 🔄 Test Workflows

### Complete Registration & Login Flow
1. Run **Register** → Account created
2. Run **Verify Email** → Email verified
3. Run **Login** → Token saved automatically
4. Run any protected endpoint → Works with saved token

### Email Verification Flow
1. **Register** → Email verification OTP sent
2. **Verify Email** → Verify with OTP code
3. **Resend Verification** → If OTP expired/lost

### Password Reset Flow
1. **Forget Password** → OTP sent to email
2. **Reset Password** → Reset with OTP + new password

### Logout Flow
1. **Login** → Get token
2. **Logout** → Token invalidated & cleared
3. Try protected endpoint → Should fail (401)

### Delete Account Flow
1. **Login** → Get token
2. **Request Delete Account** → OTP sent for confirmation
3. **Delete Account** → Account deleted, token cleared
4. Try any endpoint → Should fail (user not found)

## 📝 Notes

### Auto Token Management
- Login automatically saves JWT token ke environment variable
- Logout automatically clears token
- Delete Account automatically clears token
- All protected endpoints menggunakan `Bearer {{token}}`

### Environment Variables
- `{{baseUrl}}` - API base URL
- `{{token}}` - JWT authentication token
- `{{testEmail}}` - Test email address
- `{{testPassword}}` - Test password
- `{{testUsername}}` - Test username
- `{{otpCode}}` - OTP verification code

### Response Testing
- Login endpoint includes script untuk auto-save token
- Logout/Delete includes script untuk auto-clear token
- Easy testing dengan consistent token management

## 🛠️ Development Tips

1. **Development Mode**: Use default values dalam environment
2. **Production Testing**: Update `baseUrl` ke production server
3. **Real Email Testing**: Update email variables ke real email addresses
4. **OTP Testing**: Update `otpCode` dengan real OTP dari email

## 📧 Email Testing
Dalam development mode, OTP codes akan terlihat di server logs. Untuk production testing, check email inbox untuk real OTP codes.

## 🔒 Security Notes
- Token disimpan dalam environment variable yang secure
- Password dan sensitive data tidak di-hardcode
- OTP verification required untuk critical operations
- Auto token cleanup untuk logout/delete operations
