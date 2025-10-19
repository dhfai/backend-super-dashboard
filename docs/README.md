# README - API Documentation

## Panduan Dokumentasi API Go Backend

Direktori `docs/` ini berisi dokumentasi lengkap untuk Go Backend API. Dokumentasi tersedia dalam beberapa format untuk memenuhi kebutuhan yang berbeda.

## 📋 Isi Dokumentasi

### 1. API_DOCUMENTATION.md
Dokumentasi API lengkap dalam format Markdown yang mencakup:
- Overview API dan fitur keamanan
- Daftar lengkap semua endpoint
- Format request dan response untuk setiap endpoint
- Contoh penggunaan dengan JSON
- Error codes dan handling
- Authentication flow
- Security features
- **NEW:** Registration flow dengan OTP verification

### 2. REGISTRATION_FLOW.md
Dokumentasi khusus untuk alur registrasi baru dengan OTP:
- Penjelasan lengkap alur registrasi 2-step
- Database schema dan tables
- Perbedaan endpoint verifikasi
- Testing guide
- Troubleshooting
- Security features

### 3. openapi.yaml
OpenAPI 3.0 specification untuk:
- Dokumentasi interaktif dengan Swagger UI
- Code generation untuk client libraries
- API testing dan validation
- Integration dengan tools development

## 🚀 Cara Menggunakan Dokumentasi

### Membaca Dokumentasi Markdown
```bash
# Buka file dokumentasi dengan editor favorit
code docs/API_DOCUMENTATION.md

# Atau preview di browser dengan markdown reader
```

### Menggunakan OpenAPI Spec

#### 1. Swagger UI Online
- Buka [Swagger Editor](https://editor.swagger.io/)
- Copy-paste isi `openapi.yaml` ke editor
- Dokumentasi interaktif akan muncul di panel kanan

#### 2. Swagger UI Local
```bash
# Install swagger-ui-serve
npm install -g swagger-ui-serve

# Jalankan Swagger UI dengan spec file
swagger-ui-serve docs/openapi.yaml

# Akses di browser: http://localhost:3000
```

#### 3. Redoc
```bash
# Install redoc-cli
npm install -g redoc-cli

# Generate static documentation
redoc build docs/openapi.yaml --output docs/redoc.html

# Buka docs/redoc.html di browser
```

### Testing API dengan Postman

#### Import Collection
1. Buka Postman
2. Pilih "Import" → "File"
3. Select `postman/postman-collection.json`
4. Import `postman/filestore-environment.json` untuk environment variables

#### Setup Environment
1. Set `baseUrl` ke server URL (default: `http://localhost:8080`)
2. Update `testEmail`, `testPassword`, dll sesuai kebutuhan
3. Token akan otomatis disimpan setelah login

## 🛠️ Tools Tambahan

### Code Generation
Generate client code dari OpenAPI spec:

```bash
# Install OpenAPI Generator
npm install @openapitools/openapi-generator-cli -g

# Generate JavaScript client
openapi-generator-cli generate \
  -i docs/openapi.yaml \
  -g javascript \
  -o clients/javascript

# Generate Python client
openapi-generator-cli generate \
  -i docs/openapi.yaml \
  -g python \
  -o clients/python

# Generate Go client
openapi-generator-cli generate \
  -i docs/openapi.yaml \
  -g go \
  -o clients/go
```

### API Validation
Validate OpenAPI spec:

```bash
# Install swagger-codegen
npm install -g swagger-codegen

# Validate spec
swagger-codegen validate -i docs/openapi.yaml
```

## 📝 Update Dokumentasi

### Ketika Menambah Endpoint Baru
1. Update `docs/API_DOCUMENTATION.md`:
   - Tambah endpoint ke daftar
   - Dokumentasikan request/response
   - Tambah contoh penggunaan

2. Update `docs/openapi.yaml`:
   - Tambah path baru di section `paths`
   - Definisikan schemas jika perlu
   - Update tags

3. Update Postman collection:
   - Tambah request baru ke collection
   - Test endpoint dan verifikasi response
   - Export collection yang sudah diupdate

### Best Practices
- Selalu update dokumentasi bersamaan dengan code changes
- Gunakan contoh yang realistis dalam dokumentasi
- Validasi OpenAPI spec sebelum commit
- Test semua endpoint yang didokumentasikan

## 🔗 Links Berguna

- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger Editor](https://editor.swagger.io/)
- [Postman Learning Center](https://learning.postman.com/)
- [Redoc Documentation](https://redoc.ly/)

## 📞 Support

Jika ada pertanyaan tentang API atau dokumentasi:
1. Baca dokumentasi lengkap di `API_DOCUMENTATION.md`
2. Cek OpenAPI spec untuk detail teknis
3. Test dengan Postman collection
4. Hubungi team development untuk support lebih lanjut
