# Go Backend API Documentation

## Overview
API ini menyediakan layanan autentikasi dan manajemen profil pengguna dengan fitur keamanan lengkap termasuk verifikasi email, reset password, dan penghapusan akun.

**Base URL:** `http://localhost:8080`
**API Version:** v1

## Authentication
API ini menggunakan JWT (JSON Web Token) untuk autentikasi. Token harus disertakan dalam header `Authorization` dengan format `Bearer <token>`.

## Response Format
Semua response API menggunakan format standar berikut:

```json
{
  "success": boolean,
  "message": string,
  "data": object (optional),
  "error": object (optional)
}
```

## Status Codes
- `200` - OK: Request berhasil
- `201` - Created: Resource berhasil dibuat
- `400` - Bad Request: Request tidak valid
- `401` - Unauthorized: Token tidak valid atau tidak ada
- `403` - Forbidden: Tidak memiliki akses
- `404` - Not Found: Resource tidak ditemukan
- `409` - Conflict: Data sudah ada
- `422` - Unprocessable Entity: Validasi gagal
- `500` - Internal Server Error: Error server

## Registration Flow (Alur Registrasi)

### Alur Lengkap Registrasi User Baru:

```
1. POST /api/v1/auth/register
   ↓ (Data disimpan temporary, OTP dikirim ke email)

2. Cek email untuk mendapatkan OTP (berlaku 15 menit)
   ↓

3. POST /api/v1/auth/verify-registration
   ↓ (User dibuat, email terverifikasi)

4. POST /api/v1/auth/login
   ✓ (Berhasil login)
```

### Jika OTP Tidak Diterima atau Expired:
```
POST /api/v1/auth/resend-registration-otp
(Kirim ulang OTP baru)
```

### ⚠️ Penting:
- **Registration berlaku 24 jam** - Jika tidak diverifikasi dalam 24 jam, harus register ulang
- **OTP berlaku 15 menit** - Setelah itu harus request OTP baru
- **User belum terdafar** sampai OTP diverifikasi di endpoint `/verify-registration`
- **Jangan gunakan** `/verify-email` untuk registrasi baru!

## Endpoints

### Health Check

#### GET /health
Mengecek status kesehatan API.

**Response:**
```json
{
  "status": "ok",
  "service": "file-store-api",
  "version": "1.0.0"
}
```

---

## Authentication Endpoints

### POST /api/v1/auth/register
Memulai proses registrasi pengguna baru. Data disimpan sementara dan OTP dikirim ke email untuk verifikasi.

**Request Body:**
```json
{
  "username": "string (required, min: 3, max: 50)",
  "email": "string (required, valid email)",
  "password": "string (required, min: 8)",
  "retype_password": "string (required, must match password)"
}
```

**Example Request:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "SecurePass123!",
  "retype_password": "SecurePass123!"
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

**Error Responses:**
- `400` - Validation error
- `409` - User already exists in users table
- `500` - Failed to process registration

**Notes:**
- Data registrasi disimpan di tabel temporary selama 24 jam
- OTP dikirim ke email dengan masa berlaku 15 menit
- User belum terdaftar sampai OTP diverifikasi
- Gunakan endpoint `/api/v1/auth/verify-registration` untuk menyelesaikan registrasi

---

### POST /api/v1/auth/verify-registration
Menyelesaikan proses registrasi dengan memverifikasi OTP. Setelah sukses, user akan terdaftar dan dapat login.

**Request Body:**
```json
{
  "email": "string (required, valid email)",
  "otp_code": "string (required, 6 characters)"
}
```

**Example Request:**
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
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "john_doe",
    "email": "john@example.com",
    "is_active": true,
    "email_verified": true,
    "email_verified_at": "2025-10-14T22:00:00Z",
    "profile": {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
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

**Error Responses:**
- `400` - Invalid or expired OTP code
- `404` - Registration not found or expired
- `500` - Failed to complete registration

**Notes:**
- OTP hanya dapat digunakan sekali
- OTP berlaku selama 15 menit
- Setelah verifikasi sukses, user langsung aktif dan email terverifikasi
- Data temporary registration akan dihapus setelah berhasil

---

### POST /api/v1/auth/resend-registration-otp
Mengirim ulang OTP untuk menyelesaikan registrasi. Gunakan jika OTP tidak diterima atau sudah expired.

**Request Body:**
```json
{
  "email": "string (required, valid email)"
}
```

**Example Request:**
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

**Error Responses:**
- `400` - Validation error or registration expired
- `404` - Registration not found
- `500` - Failed to send email

**Notes:**
- OTP lama yang belum digunakan akan dihapus
- OTP baru berlaku 15 menit
- Bisa digunakan berkali-kali selama temp registration belum expired (24 jam)

---

### POST /api/v1/auth/login
Login pengguna.

**Request Body:**
```json
{
  "email": "string (required, valid email)",
  "password": "string (required)"
}
```

**Example Request:**
```json
{
  "email": "john@example.com",
  "password": "SecurePass123!"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "token": "jwt_token_here",
    "user": {
      "id": "uuid",
      "username": "john_doe",
      "email": "john@example.com",
      "is_active": true,
      "email_verified": true,
      "email_verified_at": "2023-10-10T10:00:00Z",
      "profile": {
        "full_name": "John Doe",
        "address": "123 Main St",
        "phone_number": "+1234567890",
        "country": "USA"
      }
    }
  }
}
```

**Error Responses:**
- `400` - Invalid credentials
- `401` - Email not verified

---

### POST /api/v1/auth/verify-email
Verifikasi email untuk user yang **sudah terdaftar** tapi belum terverifikasi.

**⚠️ PERHATIAN:**
- Endpoint ini untuk user yang **sudah ada** di database
- Untuk menyelesaikan registrasi baru, gunakan `/api/v1/auth/verify-registration`

**Request Body:**
```json
{
  "email": "string (required, valid email)",
  "otp_code": "string (required, 6 characters)"
}
```

**Example Request:**
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
  "message": "Email verified successfully"
}
```

**Error Responses:**
- `400` - Invalid or expired OTP
- `404` - User not found

**Notes:**
- User harus sudah terdaftar di sistem
- Digunakan untuk re-verifikasi email user existing
- Berbeda dengan verify-registration yang membuat user baru

---

### POST /api/v1/auth/resend-verification
Mengirim ulang kode verifikasi email.

**Request Body:**
```json
{
  "email": "string (required, valid email)"
}
```

**Example Request:**
```json
{
  "email": "john@example.com"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Verification code sent successfully"
}
```

---

### POST /api/v1/auth/forget-password
Meminta reset password.

**Request Body:**
```json
{
  "email": "string (required, valid email)"
}
```

**Example Request:**
```json
{
  "email": "john@example.com"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Password reset code sent to your email"
}
```

---

### POST /api/v1/auth/reset-password
Reset password dengan OTP code.

**Request Body:**
```json
{
  "email": "string (required, valid email)",
  "otp_code": "string (required, 6 characters)",
  "new_password": "string (required, min: 8)"
}
```

**Example Request:**
```json
{
  "email": "john@example.com",
  "otp_code": "123456",
  "new_password": "NewSecurePass123!"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Password reset successfully"
}
```

---

## Protected Endpoints
Semua endpoint berikut memerlukan autentikasi dengan JWT token.

### POST /api/v1/auth/logout
Logout pengguna (memerlukan autentikasi).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Logout successful"
}
```

---

### POST /api/v1/auth/request-delete-account
Meminta penghapusan akun (memerlukan autentikasi).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "password": "string (required)"
}
```

**Example Request:**
```json
{
  "password": "SecurePass123!"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Account deletion code sent to your email"
}
```

---

### POST /api/v1/auth/delete-account
Menghapus akun dengan OTP code (memerlukan autentikasi).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "otp_code": "string (required, 6 characters)"
}
```

**Example Request:**
```json
{
  "otp_code": "123456"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Account deleted successfully"
}
```

---

## Profile Management Endpoints

### GET /api/v1/profile
Mendapatkan profil pengguna (memerlukan autentikasi).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "full_name": "John Doe",
    "address": "123 Main Street",
    "phone_number": "+1234567890",
    "country": "USA",
    "created_at": "2023-10-10T10:00:00Z",
    "updated_at": "2023-10-10T12:00:00Z"
  }
}
```

---

### PUT /api/v1/profile
Update profil pengguna (memerlukan autentikasi).

**Headers:**
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

**Request Body:**
```json
{
  "full_name": "string (optional, max: 100)",
  "address": "string (optional)",
  "phone_number": "string (optional, max: 20)",
  "country": "string (optional, max: 50)"
}
```

**Example Request:**
```json
{
  "full_name": "John Doe Updated",
  "address": "456 New Street, New City",
  "phone_number": "+1987654321",
  "country": "United States"
}
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "full_name": "John Doe Updated",
    "address": "456 New Street, New City",
    "phone_number": "+1987654321",
    "country": "United States",
    "updated_at": "2023-10-10T15:00:00Z"
  }
}
```

---

### DELETE /api/v1/profile
Menghapus profil pengguna (memerlukan autentikasi).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "Profile deleted successfully"
}
```

---

### GET /api/v1/user/info
Mendapatkan informasi pengguna lengkap (memerlukan autentikasi).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Success Response (200):**
```json
{
  "success": true,
  "message": "User info retrieved successfully",
  "data": {
    "id": "uuid",
    "username": "john_doe",
    "email": "john@example.com",
    "is_active": true,
    "email_verified": true,
    "email_verified_at": "2023-10-10T10:00:00Z",
    "profile": {
      "full_name": "John Doe",
      "address": "123 Main Street",
      "phone_number": "+1234567890",
      "country": "USA"
    }
  }
}
```

---

## Error Responses

### Validation Error (400)
```json
{
  "success": false,
  "message": "Validation failed",
  "error": [
    {
      "field": "email",
      "message": "Email is required"
    },
    {
      "field": "password",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

### Unauthorized (401)
```json
{
  "success": false,
  "message": "Unauthorized",
  "error": "Invalid or missing token"
}
```

### Resource Not Found (404)
```json
{
  "success": false,
  "message": "Resource not found",
  "error": "User not found"
}
```

### Internal Server Error (500)
```json
{
  "success": false,
  "message": "Internal server error",
  "error": "Something went wrong"
}
```

---

## Security Features

1. **JWT Authentication**: Menggunakan JWT untuk autentikasi yang aman
2. **Password Hashing**: Password di-hash menggunakan bcrypt
3. **Email Verification**: Verifikasi email dengan OTP 6 digit
4. **Rate Limiting**: Pembatasan request untuk mencegah abuse
5. **CORS Protection**: Konfigurasi CORS untuk keamanan
6. **Input Validation**: Validasi ketat untuk semua input
7. **Security Headers**: Header keamanan standar

## Rate Limiting
API ini menerapkan rate limiting untuk mencegah abuse:
- Default: 100 requests per 15 menit per IP

## Environment Variables
```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
JWT_SECRET=your_jwt_secret
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your_email@gmail.com
EMAIL_PASSWORD=your_email_password
```

## Testing
Gunakan Postman collection yang tersedia di `postman/postman-collection.json` untuk testing semua endpoint.

## Contact
Untuk pertanyaan atau masalah, silakan hubungi tim development.
