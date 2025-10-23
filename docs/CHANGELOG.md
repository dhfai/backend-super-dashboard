# CHANGELOG - API Updates

## [2.2.0] - 2025-10-24

### 🔄 Breaking Changes - Daily Target Simplification

#### Removed
- **Savings Target Field** from Daily Target feature
  - Removed `savings_target` from request body (Create & Update)
  - Removed `savings_target` from response
  - Removed `savings_progress` from response
  - Removed savings-related calculations from summary endpoint

#### Modified Endpoints
- **POST `/api/v1/daily-targets`** - Create Daily Target
  - No longer accepts `savings_target` field
  - Request body now only includes: `date`, `income_target`, `expense_limit`, `notes`

- **PUT `/api/v1/daily-targets/:id`** - Update Daily Target
  - No longer accepts `savings_target` field
  - Update body now only includes: `income_target`, `expense_limit`, `notes`

- **GET `/api/v1/daily-targets/:id`** - Get Daily Target
  - Response no longer includes `savings_target` or `savings_progress` fields

- **GET `/api/v1/daily-targets/summary`** - Get Summary
  - Response no longer includes `total_savings_target` or `days_met_savings` fields

#### Database Migration
- Migration file: `002_remove_savings_target.sql`
- Removes `savings_target` column from `daily_targets` table

#### Documentation
- See `docs/FEATURE_REMOVAL_SAVINGS_TARGET.md` for detailed migration guide

---

## [2.1.0] - 2025-10-19

### 🎉 New Feature - Daily Notes API

#### Added
- **Complete Notes/Journal Management System**
  - Full CRUD operations for daily notes
  - Search and filtering capabilities
  - Tags system for organization
  - Favorites functionality
  - Pagination and sorting support

#### New Endpoints (11 Total)
1. **POST `/api/v1/notes`** - Create Note ⭐
   - Create new daily note with title, content, tags
   - Support for favorite marking
   - Response: `201 Created` with note data

2. **GET `/api/v1/notes`** - Get All Notes
   - Retrieve all user's notes
   - Advanced filtering: search, tags, favorite status
   - Pagination support (page, page_size)
   - Sorting: by created_at, updated_at, or title (asc/desc)
   - Response: `200 OK` with paginated notes list

3. **GET `/api/v1/notes/:id`** - Get Single Note
   - Retrieve specific note by ID
   - Response: `200 OK` with note data

4. **PUT `/api/v1/notes/:id`** - Update Note
   - Update title, content, tags, or favorite status
   - Partial updates supported
   - Response: `200 OK` with updated note

5. **DELETE `/api/v1/notes/:id`** - Soft Delete Note
   - Soft delete (can be recovered)
   - Response: `200 OK`

6. **DELETE `/api/v1/notes/:id/hard-delete`** - Hard Delete Note ⚠️
   - Permanent deletion (cannot be undone)
   - Response: `200 OK`

7. **PATCH `/api/v1/notes/:id/favorite`** - Toggle Favorite
   - Toggle favorite status on/off
   - Response: `200 OK` with updated note

8. **GET `/api/v1/notes/favorites`** - Get Favorite Notes
   - Retrieve all favorite notes
   - Pagination support
   - Response: `200 OK` with favorites list

9. **GET `/api/v1/notes/search`** - Search Notes
   - Search by keyword in title, content, or tags
   - Pagination support
   - Query param: `q` (required)
   - Response: `200 OK` with search results

10. **GET `/api/v1/notes/tags`** - Get Notes by Tags
    - Filter notes by specific tags
    - Multiple tags support (comma-separated)
    - Pagination support
    - Query param: `tags` (required)
    - Response: `200 OK` with filtered notes

11. **GET `/api/v1/notes/count`** - Get Notes Count
    - Get total count of user's notes
    - Response: `200 OK` with count

#### Database Changes
- **New Table: `notes`**
  - `id` (Primary Key)
  - `user_id` (Foreign Key to users)
  - `title` (varchar 255, required)
  - `content` (text, required)
  - `tags` (varchar 500, comma-separated)
  - `is_favorite` (boolean, default: false)
  - `created_at`, `updated_at`, `deleted_at` (timestamps)

#### New Files
- `internal/models/note.go` - Note model definition
- `internal/services/note_service.go` - Notes business logic
- `internal/controllers/note_controller.go` - HTTP handlers
- `docs/NOTES_API.md` - Complete API documentation
- `docs/NOTES_API_QUICK_REFERENCE.md` - Quick reference guide
- `tests/notes_api.http` - API testing examples

#### Features
✅ **Full CRUD Operations**
- Create, read, update, and delete notes
- Soft delete with recovery option
- Hard delete for permanent removal

✅ **Advanced Search & Filtering**
- Full-text search in title, content, and tags
- Filter by tags (supports multiple tags)
- Filter by favorite status
- Combine multiple filters

✅ **Pagination & Sorting**
- Configurable page size (max 100)
- Sort by created_at, updated_at, or title
- Ascending or descending order
- Total pages and count in response

✅ **Favorites System**
- Mark/unmark notes as favorites
- Quick access to favorite notes only
- Toggle favorite status easily

✅ **Tags Organization**
- Comma-separated tags for categorization
- Filter notes by specific tags
- Search within tags

✅ **Security**
- All endpoints require JWT authentication
- User-specific note isolation
- Comprehensive input validation

#### Documentation
- Complete API documentation with examples
- Quick reference guide for developers
- HTTP test file with 50+ test cases
- Error handling examples
- Integration test flow

---

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
