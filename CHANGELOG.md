# CHANGELOG - API Updates

## [2.0.0] - 2025-10-14

### 🎉 Major Changes - New Registration Flow

#### Added
- **New Registration Flow dengan OTP Verification**
  - Registrasi sekarang memerlukan 2 langkah (register + verify OTP)
  - Data registrasi disimpan temporary sampai diverifikasi
  - User baru dibuat setelah OTP valid

#### New Endpoints
1. **POST `/api/v1/auth/register`** - Modified
   - Tidak langsung membuat user
   - Menyimpan data ke tabel temporary
   - Mengirim OTP ke email
   - Response: `201 Created` dengan info registration pending

2. **POST `/api/v1/auth/verify-registration`** - New ⭐
   - Memverifikasi OTP registrasi
   - Membuat user dari data temporary
   - Email langsung terverifikasi
   - Response: `200 OK` dengan data user lengkap

3. **POST `/api/v1/auth/resend-registration-otp`** - New
   - Mengirim ulang OTP registrasi
   - Menghapus OTP lama yang belum digunakan
   - Response: `200 OK` dengan confirmation

#### Modified Endpoints
- **POST `/api/v1/auth/verify-email`**
  - Tetap untuk verifikasi email existing user
  - Ditambahkan dokumentasi untuk membedakan dengan verify-registration
  - Tidak berubah functionality

#### Database Changes
- **New Table:** `temp_registrations`
  - `id` (UUID, PK)
  - `username` (VARCHAR 50)
  - `email` (VARCHAR 100)
  - `password` (VARCHAR 255, hashed)
  - `expires_at` (TIMESTAMP)
  - `created_at` (TIMESTAMP)
  - `updated_at` (TIMESTAMP)

- **Modified Table:** `otps`
  - Added: `temp_registration_id` (UUID, nullable)
  - Modified: `user_id` (UUID, nullable - was required)
  - Modified: `type` (VARCHAR 30 - was 20)
  - New type: `register_verify` (untuk registrasi)

#### OTP Types
- `register_verify` - NEW: Untuk verifikasi registrasi baru
- `verify_email` - Existing: Untuk verifikasi email existing user
- `reset_password` - Existing: Untuk reset password
- `delete_account` - Existing: Untuk konfirmasi hapus akun

#### Cleanup Service Updates
- Added automatic cleanup for expired temp registrations
- Cleanup runs periodically to remove:
  - Temp registrations > 24 hours
  - Related OTPs for expired registrations

#### Security Improvements
- User tidak dibuat sampai email diverifikasi
- Tidak ada "dangling users" dengan email unverified
- Transaction safety saat membuat user dari temp registration
- Auto cleanup untuk data temporary yang expired

#### Breaking Changes ⚠️
- **Registration flow berbeda dari sebelumnya**
  - Old: Register → User created → (Optional) Verify email
  - New: Register → Verify OTP → User created ✅

- **Login akan gagal jika registrasi belum selesai**
  - User harus verify OTP dulu sebelum bisa login

- **Existing users tidak terpengaruh**
  - User yang sudah terdaftar tetap bisa login seperti biasa

#### Migration Guide
1. **Backup database**
2. **Stop aplikasi**
3. **Pull latest code**
4. **Restart aplikasi** (auto-migration akan membuat tabel baru)
5. **Update client aplikasi** untuk menggunakan endpoint baru
6. **Test registration flow**

#### Documentation Updates
- ✅ `API_DOCUMENTATION.md` - Updated dengan endpoint baru
- ✅ `REGISTRATION_FLOW.md` - NEW: Panduan lengkap registration flow
- ✅ `API_QUICK_REFERENCE.md` - NEW: Quick reference card
- ✅ `tests/registration_flow.http` - NEW: HTTP test file
- ✅ `README.md` - Updated dengan info flow baru

#### Testing
- Added comprehensive HTTP test file untuk registration flow
- Test cases untuk:
  - Successful registration
  - Invalid OTP
  - Expired registration
  - Resend OTP
  - Duplicate registration

---

## [1.0.0] - Previous Version

### Features
- Basic authentication (register, login, logout)
- Email verification (optional)
- Password reset with OTP
- Profile management
- Account deletion with OTP
- JWT token authentication
- Token blacklisting on logout
- Automatic cleanup service

### Original Endpoints
- POST `/auth/register` - Create user immediately
- POST `/auth/login`
- POST `/auth/logout`
- POST `/auth/verify-email`
- POST `/auth/resend-verification`
- POST `/auth/forget-password`
- POST `/auth/reset-password`
- GET `/profile`
- PUT `/profile`
- POST `/profile/request-delete`
- DELETE `/profile/delete`

---

## Upgrade Notes

### For API Consumers
1. Update registration flow in your client apps
2. Handle 2-step registration (register + verify)
3. Show appropriate messages for pending registration
4. Implement resend OTP functionality
5. Update error handling for new error codes

### For Backend Developers
1. Review new models and migrations
2. Test cleanup service functionality
3. Monitor temp_registrations table growth
4. Update any integration tests
5. Review security implications

### Backward Compatibility
- ❌ Registration flow is NOT backward compatible
- ✅ Login endpoint is backward compatible
- ✅ All other endpoints unchanged
- ✅ Existing users not affected

---

## Roadmap

### Planned Features
- [ ] Rate limiting for OTP requests
- [ ] SMS OTP as alternative to email
- [ ] Social authentication (Google, Facebook)
- [ ] 2FA (Two-Factor Authentication)
- [ ] Account recovery without email access
- [ ] Password strength meter
- [ ] Registration via invitation

### Under Consideration
- [ ] Magic link authentication
- [ ] Biometric authentication support
- [ ] Session management dashboard
- [ ] Login history tracking
- [ ] Suspicious activity detection

---

## Support

For questions or issues:
1. Check `REGISTRATION_FLOW.md` for detailed guide
2. Review `API_QUICK_REFERENCE.md` for quick answers
3. Test with `tests/registration_flow.http`
4. Contact development team

---

Last Updated: October 14, 2025
