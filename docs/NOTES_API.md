# Notes API Documentation

## Overview
The Notes API provides comprehensive endpoints for managing daily notes/journal entries. All endpoints require authentication via JWT token.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All Notes endpoints require authentication. Include the JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

---

## Endpoints

### 1. Create Note
Create a new note for the authenticated user.

**Endpoint:** `POST /notes`

**Request Body:**
```json
{
  "title": "My Daily Note",
  "content": "This is the content of my note",
  "tags": "work,personal,ideas",
  "is_favorite": false
}
```

**Field Validation:**
- `title`: Required, 1-255 characters
- `content`: Required, minimum 1 character
- `tags`: Optional, maximum 500 characters, comma-separated
- `is_favorite`: Optional, boolean (default: false)

**Response (201 Created):**
```json
{
  "success": true,
  "message": "Note created successfully",
  "data": {
    "id": 1,
    "user_id": 123,
    "title": "My Daily Note",
    "content": "This is the content of my note",
    "tags": "work,personal,ideas",
    "is_favorite": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 2. Get All Notes
Retrieve all notes for the authenticated user with filtering, pagination, and sorting.

**Endpoint:** `GET /notes`

**Query Parameters:**
- `search` (optional): Search in title, content, or tags
- `tags` (optional): Filter by tags (comma-separated)
- `is_favorite` (optional): Filter by favorite status (true/false)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 10, max: 100)
- `sort_by` (optional): Sort field - `created_at`, `updated_at`, `title` (default: `created_at`)
- `sort_order` (optional): Sort order - `asc`, `desc` (default: `desc`)

**Example Request:**
```
GET /notes?search=meeting&page=1&page_size=20&sort_by=created_at&sort_order=desc
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Notes retrieved successfully",
  "data": {
    "notes": [
      {
        "id": 1,
        "user_id": 123,
        "title": "Meeting Notes",
        "content": "Important meeting discussion",
        "tags": "work,meeting",
        "is_favorite": true,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 45,
    "page": 1,
    "page_size": 20,
    "total_pages": 3
  }
}
```

---

### 3. Get Single Note
Retrieve a specific note by ID.

**Endpoint:** `GET /notes/:id`

**Path Parameters:**
- `id`: Note ID (integer)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Note retrieved successfully",
  "data": {
    "id": 1,
    "user_id": 123,
    "title": "My Daily Note",
    "content": "This is the content of my note",
    "tags": "work,personal",
    "is_favorite": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

---

### 4. Update Note
Update an existing note.

**Endpoint:** `PUT /notes/:id`

**Path Parameters:**
- `id`: Note ID (integer)

**Request Body:**
```json
{
  "title": "Updated Title",
  "content": "Updated content",
  "tags": "updated,tags",
  "is_favorite": true
}
```

**Note:** All fields are optional. Only provided fields will be updated.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Note updated successfully",
  "data": {
    "id": 1,
    "user_id": 123,
    "title": "Updated Title",
    "content": "Updated content",
    "tags": "updated,tags",
    "is_favorite": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T12:45:00Z"
  }
}
```

---

### 5. Delete Note (Soft Delete)
Soft delete a note (can be recovered).

**Endpoint:** `DELETE /notes/:id`

**Path Parameters:**
- `id`: Note ID (integer)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Note deleted successfully"
}
```

---

### 6. Hard Delete Note
Permanently delete a note from the database.

**Endpoint:** `DELETE /notes/:id/hard-delete`

**Path Parameters:**
- `id`: Note ID (integer)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Note permanently deleted successfully"
}
```

---

### 7. Toggle Favorite Status
Toggle the favorite status of a note.

**Endpoint:** `PATCH /notes/:id/favorite`

**Path Parameters:**
- `id`: Note ID (integer)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Favorite status toggled successfully",
  "data": {
    "id": 1,
    "user_id": 123,
    "title": "My Daily Note",
    "content": "This is the content of my note",
    "tags": "work,personal",
    "is_favorite": true,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T12:45:00Z"
  }
}
```

---

### 8. Get Favorite Notes
Retrieve all favorite notes for the authenticated user.

**Endpoint:** `GET /notes/favorites`

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 10)

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Favorite notes retrieved successfully",
  "data": {
    "notes": [
      {
        "id": 1,
        "user_id": 123,
        "title": "Important Note",
        "content": "This is important",
        "tags": "work",
        "is_favorite": true,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "page_size": 10,
    "total_pages": 1
  }
}
```

---

### 9. Search Notes
Search notes by keyword in title, content, or tags.

**Endpoint:** `GET /notes/search`

**Query Parameters:**
- `q` (required): Search keyword
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 10)

**Example Request:**
```
GET /notes/search?q=meeting&page=1&page_size=20
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Notes search completed successfully",
  "data": {
    "notes": [
      {
        "id": 1,
        "user_id": 123,
        "title": "Meeting Notes",
        "content": "Discussion about project",
        "tags": "work,meeting",
        "is_favorite": false,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 3,
    "page": 1,
    "page_size": 20,
    "total_pages": 1
  }
}
```

---

### 10. Get Notes by Tag
Retrieve notes filtered by specific tags.

**Endpoint:** `GET /notes/tags`

**Query Parameters:**
- `tags` (required): Tags to filter by (comma-separated)
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 10)

**Example Request:**
```
GET /notes/tags?tags=work,personal&page=1&page_size=20
```

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Notes retrieved successfully",
  "data": {
    "notes": [
      {
        "id": 1,
        "user_id": 123,
        "title": "Work Note",
        "content": "Project tasks",
        "tags": "work,personal,project",
        "is_favorite": false,
        "created_at": "2024-01-15T10:30:00Z",
        "updated_at": "2024-01-15T10:30:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "page_size": 20,
    "total_pages": 1
  }
}
```

---

### 11. Get Notes Count
Get the total count of notes for the authenticated user.

**Endpoint:** `GET /notes/count`

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Notes count retrieved successfully",
  "data": {
    "count": 45
  }
}
```

---

## Error Responses

### 400 Bad Request
Invalid request data or validation error.
```json
{
  "success": false,
  "message": "Validation failed",
  "error": [
    {
      "field": "title",
      "message": "title is required"
    }
  ]
}
```

### 401 Unauthorized
Missing or invalid authentication token.
```json
{
  "success": false,
  "message": "Unauthorized"
}
```

### 404 Not Found
Note not found or doesn't belong to the user.
```json
{
  "success": false,
  "message": "Note not found",
  "error": "note not found"
}
```

### 500 Internal Server Error
Server-side error.
```json
{
  "success": false,
  "message": "Failed to create note",
  "error": "database connection error"
}
```

---

## Features

### 1. **Full CRUD Operations**
- Create, Read, Update, and Delete notes
- Soft delete with recovery option
- Hard delete for permanent removal

### 2. **Search & Filter**
- Full-text search in title, content, and tags
- Filter by tags (multiple tags support)
- Filter by favorite status

### 3. **Pagination**
- Configurable page size (max 100 items)
- Total pages and total count included in response

### 4. **Sorting**
- Sort by creation date, update date, or title
- Ascending or descending order

### 5. **Favorites**
- Mark notes as favorites
- Quick access to favorite notes
- Toggle favorite status easily

### 6. **Tags System**
- Comma-separated tags for organization
- Filter notes by specific tags
- Search within tags

### 7. **Security**
- All endpoints require authentication
- Users can only access their own notes
- Input validation on all requests

---

## Best Practices

1. **Always include authentication token** in the Authorization header
2. **Use pagination** for large note collections to improve performance
3. **Use soft delete** instead of hard delete to allow recovery
4. **Use tags** to organize notes effectively
5. **Use search** for quick access to specific notes
6. **Validate input** before sending requests to avoid errors

---

## Examples

### Creating a Daily Journal Entry
```bash
curl -X POST http://localhost:8080/api/v1/notes \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Daily Journal - January 15, 2024",
    "content": "Today was productive. Completed project milestone and had a great team meeting.",
    "tags": "journal,daily,work",
    "is_favorite": false
  }'
```

### Searching for Work-Related Notes
```bash
curl -X GET "http://localhost:8080/api/v1/notes/search?q=project&page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Getting All Favorite Notes
```bash
curl -X GET "http://localhost:8080/api/v1/notes/favorites?page=1&page_size=20" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### Filtering by Tags
```bash
curl -X GET "http://localhost:8080/api/v1/notes/tags?tags=work,important" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```
