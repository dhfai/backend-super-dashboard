package services

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/google/uuid"

	"gorm.io/gorm"
)

type NoteService struct {
	db *gorm.DB
}

func NewNoteService(db *gorm.DB) *NoteService {
	return &NoteService{db: db}
}

// CreateNote creates a new note for a user
func (s *NoteService) CreateNote(userID uuid.UUID, req *models.CreateNoteRequest) (*models.Note, error) {
	note := &models.Note{
		UserID:     userID,
		Title:      req.Title,
		Content:    req.Content,
		Tags:       req.Tags,
		IsFavorite: req.IsFavorite,
	}

	if err := s.db.Create(note).Error; err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

// GetNoteByID retrieves a note by its ID
func (s *NoteService) GetNoteByID(noteID uint, userID uuid.UUID) (*models.Note, error) {
	var note models.Note
	err := s.db.Where("id = ? AND user_id = ?", noteID, userID).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("note not found")
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}

	return &note, nil
}

// GetAllNotes retrieves all notes for a user with filtering, pagination, and sorting
func (s *NoteService) GetAllNotes(userID uuid.UUID, filter *models.NoteFilterRequest) (*models.NotesListResponse, error) {
	// Set default values
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 10
	}
	if filter.SortBy == "" {
		filter.SortBy = "created_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	// Build query
	query := s.db.Model(&models.Note{}).Where("user_id = ?", userID)

	// Apply search filter
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("title LIKE ? OR content LIKE ? OR tags LIKE ?", searchPattern, searchPattern, searchPattern)
	}

	// Apply tags filter
	if filter.Tags != "" {
		tags := strings.Split(filter.Tags, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				query = query.Where("tags LIKE ?", "%"+tag+"%")
			}
		}
	}

	// Apply favorite filter
	if filter.IsFavorite != nil {
		query = query.Where("is_favorite = ?", *filter.IsFavorite)
	}

	// Count total records
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count notes: %w", err)
	}

	// Apply sorting
	validSortFields := map[string]bool{
		"created_at": true,
		"updated_at": true,
		"title":      true,
	}
	sortBy := filter.SortBy
	if !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	sortOrder := strings.ToUpper(filter.SortOrder)
	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	query = query.Order(fmt.Sprintf("%s %s", sortBy, sortOrder))

	// Apply pagination
	offset := (filter.Page - 1) * filter.PageSize
	query = query.Offset(offset).Limit(filter.PageSize)

	// Fetch notes
	var notes []models.Note
	if err := query.Find(&notes).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch notes: %w", err)
	}

	// Convert to response
	noteResponses := make([]models.NoteResponse, len(notes))
	for i, note := range notes {
		noteResponses[i] = *note.ToNoteResponse()
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.PageSize)))

	return &models.NotesListResponse{
		Notes:      noteResponses,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateNote updates an existing note
func (s *NoteService) UpdateNote(noteID uint, userID uuid.UUID, req *models.UpdateNoteRequest) (*models.Note, error) {
	// Check if note exists and belongs to user
	note, err := s.GetNoteByID(noteID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	updates := make(map[string]interface{})

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	if req.IsFavorite != nil {
		updates["is_favorite"] = *req.IsFavorite
	}

	if len(updates) == 0 {
		return note, nil // No updates
	}

	if err := s.db.Model(note).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update note: %w", err)
	}

	// Fetch updated note
	updatedNote, err := s.GetNoteByID(noteID, userID)
	if err != nil {
		return nil, err
	}

	return updatedNote, nil
}

// DeleteNote soft deletes a note
func (s *NoteService) DeleteNote(noteID uint, userID uuid.UUID) error {
	// Check if note exists and belongs to user
	note, err := s.GetNoteByID(noteID, userID)
	if err != nil {
		return err
	}

	if err := s.db.Delete(note).Error; err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

// HardDeleteNote permanently deletes a note
func (s *NoteService) HardDeleteNote(noteID uint, userID uuid.UUID) error {
	result := s.db.Unscoped().Where("id = ? AND user_id = ?", noteID, userID).Delete(&models.Note{})
	if result.Error != nil {
		return fmt.Errorf("failed to permanently delete note: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("note not found")
	}

	return nil
}

// GetFavoriteNotes retrieves all favorite notes for a user
func (s *NoteService) GetFavoriteNotes(userID uuid.UUID, page, pageSize int) (*models.NotesListResponse, error) {
	favorite := true
	filter := &models.NoteFilterRequest{
		IsFavorite: &favorite,
		Page:       page,
		PageSize:   pageSize,
		SortBy:     "created_at",
		SortOrder:  "desc",
	}
	return s.GetAllNotes(userID, filter)
}

// SearchNotes searches notes by keyword
func (s *NoteService) SearchNotes(userID uuid.UUID, keyword string, page, pageSize int) (*models.NotesListResponse, error) {
	filter := &models.NoteFilterRequest{
		Search:    keyword,
		Page:      page,
		PageSize:  pageSize,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	return s.GetAllNotes(userID, filter)
}

// GetNotesByTag retrieves notes with specific tags
func (s *NoteService) GetNotesByTag(userID uuid.UUID, tags string, page, pageSize int) (*models.NotesListResponse, error) {
	filter := &models.NoteFilterRequest{
		Tags:      tags,
		Page:      page,
		PageSize:  pageSize,
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	return s.GetAllNotes(userID, filter)
}

// ToggleFavorite toggles the favorite status of a note
func (s *NoteService) ToggleFavorite(noteID uint, userID uuid.UUID) (*models.Note, error) {
	note, err := s.GetNoteByID(noteID, userID)
	if err != nil {
		return nil, err
	}

	note.IsFavorite = !note.IsFavorite
	if err := s.db.Save(note).Error; err != nil {
		return nil, fmt.Errorf("failed to toggle favorite: %w", err)
	}

	return note, nil
}

// GetNotesCount returns the total count of notes for a user
func (s *NoteService) GetNotesCount(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.Note{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count notes: %w", err)
	}
	return count, nil
}
