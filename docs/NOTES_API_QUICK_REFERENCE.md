# Notes API Quick Reference

## Authentication Required
All endpoints require JWT token in Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

---

## Endpoints Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| **POST** | `/api/v1/notes` | Create a new note |
| **GET** | `/api/v1/notes` | Get all notes (with filters) |
| **GET** | `/api/v1/notes/:id` | Get single note by ID |
| **PUT** | `/api/v1/notes/:id` | Update a note |
| **DELETE** | `/api/v1/notes/:id` | Soft delete a note |
| **DELETE** | `/api/v1/notes/:id/hard-delete` | Permanently delete a note |
| **PATCH** | `/api/v1/notes/:id/favorite` | Toggle favorite status |
| **GET** | `/api/v1/notes/favorites` | Get all favorite notes |
| **GET** | `/api/v1/notes/search` | Search notes by keyword |
| **GET** | `/api/v1/notes/tags` | Get notes by tags |
| **GET** | `/api/v1/notes/count` | Get total notes count |

---

## Quick Examples

### Create Note
```bash
POST /api/v1/notes
{
  "title": "My Note",
  "content": "Note content",
  "tags": "work,personal",
  "is_favorite": false
}
```

### Get All Notes with Filters
```bash
GET /api/v1/notes?search=meeting&tags=work&is_favorite=true&page=1&page_size=20&sort_by=created_at&sort_order=desc
```

### Update Note
```bash
PUT /api/v1/notes/1
{
  "title": "Updated Title",
  "content": "Updated content"
}
```

### Search Notes
```bash
GET /api/v1/notes/search?q=meeting&page=1&page_size=20
```

### Get Favorite Notes
```bash
GET /api/v1/notes/favorites?page=1&page_size=10
```

### Get Notes by Tags
```bash
GET /api/v1/notes/tags?tags=work,important&page=1
```

---

## Query Parameters

### Pagination
- `page`: Page number (default: 1)
- `page_size`: Items per page (default: 10, max: 100)

### Filtering
- `search`: Search in title, content, or tags
- `tags`: Filter by tags (comma-separated)
- `is_favorite`: Filter by favorite status (true/false)

### Sorting
- `sort_by`: Field to sort by (`created_at`, `updated_at`, `title`)
- `sort_order`: Sort order (`asc`, `desc`)

---

## Response Format

### Success Response
```json
{
  "success": true,
  "message": "Operation successful",
  "data": { ... }
}
```

### Error Response
```json
{
  "success": false,
  "message": "Error message",
  "error": "Error details"
}
```

### List Response
```json
{
  "success": true,
  "message": "Notes retrieved successfully",
  "data": {
    "notes": [...],
    "total": 45,
    "page": 1,
    "page_size": 10,
    "total_pages": 5
  }
}
```

---

## Features Checklist

✅ **CRUD Operations**
- Create, Read, Update, Delete notes
- Soft delete with recovery option
- Hard delete for permanent removal

✅ **Search & Filter**
- Full-text search in title, content, and tags
- Filter by tags (multiple tags support)
- Filter by favorite status

✅ **Pagination**
- Configurable page size
- Total pages and count included

✅ **Sorting**
- Sort by created_at, updated_at, or title
- Ascending or descending order

✅ **Favorites**
- Mark notes as favorites
- Quick access to favorite notes
- Toggle favorite status

✅ **Tags System**
- Comma-separated tags
- Filter and search by tags

✅ **Security**
- JWT authentication required
- User-specific note isolation
- Input validation

---

## Status Codes

- `200 OK`: Success
- `201 Created`: Resource created
- `400 Bad Request`: Invalid input
- `401 Unauthorized`: Missing/invalid token
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

---

For detailed documentation, see [NOTES_API.md](./NOTES_API.md)
