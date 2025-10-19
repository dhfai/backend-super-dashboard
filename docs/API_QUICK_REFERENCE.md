# API Quick Reference - Authentication Endpoints

## 🔐 Registration & Verification

### Register New User (Step 1)
```
POST /api/v1/auth/register
Body: { username, email, password, retype_password }
→ Saves to temp table, sends OTP
```

### Verify Registration (Step 2) ⭐
```
POST /api/v1/auth/verify-registration
Body: { email, otp_code }
→ Creates user, completes registration
```

### Resend Registration OTP
```
POST /api/v1/auth/resend-registration-otp
Body: { email }
→ Sends new OTP for pending registration
```

---

## 🔑 Login & Session

### Login
```
POST /api/v1/auth/login
Body: { email, password }
→ Returns JWT token
```

### Logout
```
POST /api/v1/auth/logout
Headers: Authorization: Bearer <token>
→ Blacklists token
```

---

## 📧 Email Verification (For Existing Users)

### Verify Email
```
POST /api/v1/auth/verify-email
Body: { email, otp_code }
→ Verifies email for existing user
```

### Resend Email Verification
```
POST /api/v1/auth/resend-verification
Body: { email }
→ Sends new OTP for email verification
```

---

## 🔒 Password Management

### Forget Password
```
POST /api/v1/auth/forget-password
Body: { email }
→ Sends OTP for password reset
```

### Reset Password
```
POST /api/v1/auth/reset-password
Body: { email, otp_code, new_password }
→ Resets password with OTP
```

---

## 👤 Profile Management

### Get Profile
```
GET /api/v1/profile
Headers: Authorization: Bearer <token>
→ Returns user profile
```

### Update Profile
```
PUT /api/v1/profile
Headers: Authorization: Bearer <token>
Body: { full_name, address, phone_number, country }
→ Updates profile information
```

---

## 🗑️ Account Deletion

### Request Delete Account
```
POST /api/v1/profile/request-delete
Headers: Authorization: Bearer <token>
Body: { password }
→ Sends OTP for account deletion
```

### Confirm Delete Account
```
DELETE /api/v1/profile/delete
Headers: Authorization: Bearer <token>
Body: { otp_code }
→ Permanently deletes account
```

---

## ⚠️ Important Differences

### Registration Endpoints
| Endpoint | Purpose | User Status |
|----------|---------|-------------|
| `/register` | Start registration | Not created yet |
| `/verify-registration` | Complete registration | Creates user |
| `/resend-registration-otp` | Resend reg OTP | Pending |

### Verification Endpoints
| Endpoint | Purpose | User Status |
|----------|---------|-------------|
| `/verify-email` | Verify existing user | Already exists |
| `/resend-verification` | Resend for existing | Already exists |

### 🚨 Common Mistakes
❌ Using `/verify-email` after `/register` → Will fail!
✅ Use `/verify-registration` after `/register`

❌ Using `/verify-registration` for existing user → Will fail!
✅ Use `/verify-email` for existing user

---

## 🕐 Timeouts & Limits

| Item | Duration |
|------|----------|
| Registration validity | 24 hours |
| OTP validity | 15 minutes |
| JWT token | 7 days |
| Session after logout | Invalid immediately |

---

## 📊 Response Codes

| Code | Meaning |
|------|---------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request / Invalid Data |
| 401 | Unauthorized / Invalid Token |
| 404 | Not Found |
| 409 | Conflict / Already Exists |
| 500 | Server Error |

---

## 🔄 Complete Registration Flow

```
1. POST /auth/register
   ↓
2. Check email for OTP
   ↓
3. POST /auth/verify-registration
   ↓
4. POST /auth/login
   ✅ Done!
```

If OTP not received:
```
POST /auth/resend-registration-otp
```

---

## 📝 Example Requests

### Complete Registration
```bash
# Step 1: Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john_doe",
    "email": "john@example.com",
    "password": "SecurePass123!",
    "retype_password": "SecurePass123!"
  }'

# Step 2: Verify (use OTP from email)
curl -X POST http://localhost:8080/api/v1/auth/verify-registration \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "otp_code": "123456"
  }'

# Step 3: Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "SecurePass123!"
  }'
```

---

## 🆘 Troubleshooting

### "Invalid email or verification code"
- Check if you're using the correct endpoint
- For new registration: use `/verify-registration`
- For existing user: use `/verify-email`

### "Registration not found or expired"
- Registration expired (>24 hours)
- Need to register again

### "Invalid or expired OTP code"
- OTP expired (>15 minutes)
- Use resend endpoint to get new OTP

### "User already exists"
- Email/username already registered
- Try login instead
- Or use different email/username
