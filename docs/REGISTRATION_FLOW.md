# Registration Flow dengan OTP Verification

## Overview

Alur registrasi baru mengharuskan user untuk memverifikasi email mereka menggunakan kode OTP sebelum akun benar-benar terdaftar di sistem. Data registrasi disimpan sementara dan hanya akan dipindahkan ke tabel users setelah OTP diverifikasi.

## Alur Registrasi

### 1. Initiate Registration

**Endpoint:** `POST /api/v1/auth/register`

**Request Body:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePassword123",
  "retype_password": "SecurePassword123"
}
```

**Success Response (201):**
```json
{
  "success": true,
  "message": "Registration initiated successfully. Please check your email for verification code to complete registration.",
  "data": {
    "email": "john@example.com",
    "expires_in": "24 hours"
  }
}
```

**Behavior:**
- Data disimpan ke tabel `temp_registrations` (bukan `users`)
- OTP 6 digit dikirim ke email dengan type `register_verify`
- OTP berlaku selama 15 menit
- Temp registration berlaku selama 24 jam
- Jika email/username sudah terdaftar di `users`, akan return error 409
- Jika email/username sudah ada di `temp_registrations`, data lama akan dihapus dan dibuat yang baru

---

### 2. Verify Registration OTP

**Endpoint:** `POST /api/v1/auth/verify-registration`

**Request Body:**
```json
{
  "email": "john@example.com",
  "otp_code": "123456"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Registration completed successfully. You can now login.",
  "data": {
    "id": "uuid-here",
    "username": "john_doe",
    "email": "john@example.com",
    "is_active": true,
    "email_verified": true,
    "email_verified_at": "2025-10-14T22:00:00Z",
    "profile": {
      "id": "uuid-here",
      "user_id": "uuid-here",
      "full_name": "",
      "address": "",
      "phone_number": "",
      "country": "",
      "created_at": "2025-10-14T22:00:00Z",
      "updated_at": "2025-10-14T22:00:00Z"
    }
  }
}
```

**Behavior:**
- Memvalidasi OTP yang dikirim
- Jika valid, membuat user baru di tabel `users`
- User langsung aktif (`is_active: true`) dan email verified (`email_verified: true`)
- Profile kosong dibuat otomatis
- Temp registration dan OTP terkait dihapus
- Semua operasi dilakukan dalam transaction untuk memastikan data consistency

**Error Responses:**
- **404:** Registration tidak ditemukan atau sudah expired
- **400:** OTP tidak valid, expired, atau sudah digunakan

---

### 3. Resend Registration OTP (Optional)

**Endpoint:** `POST /api/v1/auth/resend-registration-otp`

**Request Body:**
```json
{
  "email": "john@example.com"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "New verification code sent to your email."
}
```

**Behavior:**
- Menghapus OTP lama yang belum digunakan
- Generate OTP baru
- Kirim OTP baru ke email
- OTP baru berlaku 15 menit

---

## Database Schema

### temp_registrations
```sql
CREATE TABLE temp_registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

### otps (updated)
```sql
CREATE TABLE otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users(id),
    temp_registration_id UUID NULL REFERENCES temp_registrations(id),
    code VARCHAR(6) NOT NULL,
    type VARCHAR(30) NOT NULL, -- 'register_verify', 'verify_email', 'reset_password', 'delete_account'
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT false,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

## OTP Types

- `register_verify` - Untuk verifikasi registrasi baru
- `verify_email` - Untuk verifikasi email user yang sudah terdaftar
- `reset_password` - Untuk reset password
- `delete_account` - Untuk konfirmasi hapus akun

## Cleanup Service

Service cleanup otomatis akan menghapus:
- Temp registrations yang sudah expired (>24 jam)
- OTP yang sudah expired atau used
- Token blacklist yang expired

## Testing Flow

### 1. Test Successful Registration
```bash
# Step 1: Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "retype_password": "password123"
  }'

# Step 2: Check email for OTP code

# Step 3: Verify registration
curl -X POST http://localhost:8080/api/v1/auth/verify-registration \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "otp_code": "123456"
  }'

# Step 4: Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 2. Test Resend OTP
```bash
# After registering, if OTP not received or expired
curl -X POST http://localhost:8080/api/v1/auth/resend-registration-otp \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

## Security Features

1. **OTP Expiration:** OTP expires in 15 minutes
2. **Temp Registration Expiration:** Temp registration expires in 24 hours
3. **One-time Use:** OTP can only be used once
4. **Transaction Safety:** User creation uses database transaction
5. **Auto Cleanup:** Expired data automatically cleaned up
6. **Password Hashing:** Passwords hashed with bcrypt before storage

## Migration Notes

Jika sudah ada database yang berjalan:

1. **Backup database terlebih dahulu**
2. **Run migration** untuk membuat tabel `temp_registrations` dan update tabel `otps`
3. **Restart aplikasi** agar auto-migration berjalan
4. **Test registration flow** dengan user baru

## Troubleshooting

### OTP tidak diterima
- Cek email spam/junk folder
- Gunakan endpoint resend-registration-otp
- Pastikan email service berfungsi dengan baik

### Registration expired
- User perlu register ulang dari awal
- Data lama otomatis dihapus

### OTP expired
- Gunakan endpoint resend-registration-otp untuk mendapatkan OTP baru
- OTP baru akan menggantikan yang lama
