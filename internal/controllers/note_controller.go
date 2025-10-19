package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dhfai/dhfai-gobackend.git/internal/models"
	"github.com/dhfai/dhfai-gobackend.git/internal/services"
	"github.com/dhfai/dhfai-gobackend.git/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type NoteController struct {
	noteService *services.NoteService
	validator   *validator.Validate
}

func NewNoteController(noteService *services.NoteService) *NoteController {
	return &NoteController{
		noteService: noteService,
		validator:   validator.New(),
	}
}

// Helper function to get user ID from context
func getUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	log := logger.GetLogger()

	userIDStr, exists := c.Get("user_id")
	if !exists {
		return uuid.Nil, fmt.Errorf("user_id not found in context")
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.WithError(err).Error("Failed to parse user ID")
		return uuid.Nil, fmt.Errorf("invalid user ID format")
	}

	return userID, nil
}

// CreateNote godoc
// @Summary Create a new note
// @Description Create a new daily note for the authenticated user
// @Tags notes
// @Accept json
// @Produce json
// @Param note body models.CreateNoteRequest true "Note data"
// @Success 201 {object} models.APIResponse{data=models.NoteResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes [post]
func (nc *NoteController) CreateNote(c *gin.Context) {
	log := logger.GetLogger()
	var req models.CreateNoteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate request
	if err := nc.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		log.WithError(err).Error("Validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   formatValidationErrors(validationErrors),
		})
		return
	}

	// Get user ID from context (set by auth middleware)
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		log.WithError(err).Error("Failed to parse user ID")
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Invalid user ID",
		})
		return
	}

	// Create note
	note, err := nc.noteService.CreateNote(userID, &req)
	if err != nil {
		log.WithError(err).Error("Failed to create note")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to create note",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, models.APIResponse{
		Success: true,
		Message: "Note created successfully",
		Data:    note.ToNoteResponse(),
	})
}

// GetNote godoc
// @Summary Get a note by ID
// @Description Get a specific note by its ID
// @Tags notes
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {object} models.APIResponse{data=models.NoteResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/{id} [get]
func (nc *NoteController) GetNote(c *gin.Context) {
	log := logger.GetLogger()
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid note ID",
		})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	note, err := nc.noteService.GetNoteByID(uint(noteID), userID)
	if err != nil {
		log.WithError(err).Error("Failed to get note")
		c.JSON(http.StatusNotFound, models.APIResponse{
			Success: false,
			Message: "Note not found",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Note retrieved successfully",
		Data:    note.ToNoteResponse(),
	})
}

// GetAllNotes godoc
// @Summary Get all notes
// @Description Get all notes for the authenticated user with filtering and pagination
// @Tags notes
// @Produce json
// @Param search query string false "Search in title, content, or tags"
// @Param tags query string false "Filter by tags (comma-separated)"
// @Param is_favorite query bool false "Filter by favorite status"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param sort_by query string false "Sort by field (created_at, updated_at, title)" default(created_at)
// @Param sort_order query string false "Sort order (asc, desc)" default(desc)
// @Success 200 {object} models.APIResponse{data=models.NotesListResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes [get]
func (nc *NoteController) GetAllNotes(c *gin.Context) {
	log := logger.GetLogger()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	var filter models.NoteFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid query parameters",
			Error:   err.Error(),
		})
		return
	}

	notes, err := nc.noteService.GetAllNotes(userID, &filter)
	if err != nil {
		log.WithError(err).Error("Failed to get notes")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve notes",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Notes retrieved successfully",
		Data:    notes,
	})
}

// UpdateNote godoc
// @Summary Update a note
// @Description Update an existing note
// @Tags notes
// @Accept json
// @Produce json
// @Param id path int true "Note ID"
// @Param note body models.UpdateNoteRequest true "Updated note data"
// @Success 200 {object} models.APIResponse{data=models.NoteResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/{id} [put]
func (nc *NoteController) UpdateNote(c *gin.Context) {
	log := logger.GetLogger()
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid note ID",
		})
		return
	}

	var req models.UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.WithError(err).Error("Failed to bind JSON")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid request body",
			Error:   err.Error(),
		})
		return
	}

	// Validate request
	if err := nc.validator.Struct(req); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		log.WithError(err).Error("Validation failed")
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Validation failed",
			Error:   formatValidationErrors(validationErrors),
		})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	note, err := nc.noteService.UpdateNote(uint(noteID), userID, &req)
	if err != nil {
		log.WithError(err).Error("Failed to update note")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to update note",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Note updated successfully",
		Data:    note.ToNoteResponse(),
	})
}

// DeleteNote godoc
// @Summary Delete a note
// @Description Soft delete a note
// @Tags notes
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/{id} [delete]
func (nc *NoteController) DeleteNote(c *gin.Context) {
	log := logger.GetLogger()
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid note ID",
		})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	if err := nc.noteService.DeleteNote(uint(noteID), userID); err != nil {
		log.WithError(err).Error("Failed to delete note")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to delete note",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Note deleted successfully",
	})
}

// HardDeleteNote godoc
// @Summary Permanently delete a note
// @Description Permanently delete a note from the database
// @Tags notes
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/{id}/hard-delete [delete]
func (nc *NoteController) HardDeleteNote(c *gin.Context) {
	log := logger.GetLogger()
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid note ID",
		})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	if err := nc.noteService.HardDeleteNote(uint(noteID), userID); err != nil {
		log.WithError(err).Error("Failed to permanently delete note")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to permanently delete note",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Note permanently deleted successfully",
	})
}

// ToggleFavorite godoc
// @Summary Toggle favorite status
// @Description Toggle the favorite status of a note
// @Tags notes
// @Produce json
// @Param id path int true "Note ID"
// @Success 200 {object} models.APIResponse{data=models.NoteResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/{id}/favorite [patch]
func (nc *NoteController) ToggleFavorite(c *gin.Context) {
	log := logger.GetLogger()
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Invalid note ID",
		})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	note, err := nc.noteService.ToggleFavorite(uint(noteID), userID)
	if err != nil {
		log.WithError(err).Error("Failed to toggle favorite")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to toggle favorite status",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Favorite status toggled successfully",
		Data:    note.ToNoteResponse(),
	})
}

// GetFavoriteNotes godoc
// @Summary Get favorite notes
// @Description Get all favorite notes for the authenticated user
// @Tags notes
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.APIResponse{data=models.NotesListResponse}
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/favorites [get]
func (nc *NoteController) GetFavoriteNotes(c *gin.Context) {
	log := logger.GetLogger()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	notes, err := nc.noteService.GetFavoriteNotes(userID, page, pageSize)
	if err != nil {
		log.WithError(err).Error("Failed to get favorite notes")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve favorite notes",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Favorite notes retrieved successfully",
		Data:    notes,
	})
}

// SearchNotes godoc
// @Summary Search notes
// @Description Search notes by keyword
// @Tags notes
// @Produce json
// @Param q query string true "Search keyword"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.APIResponse{data=models.NotesListResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/search [get]
func (nc *NoteController) SearchNotes(c *gin.Context) {
	log := logger.GetLogger()
	keyword := c.Query("q")
	if keyword == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Search keyword is required",
		})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	notes, err := nc.noteService.SearchNotes(userID, keyword, page, pageSize)
	if err != nil {
		log.WithError(err).Error("Failed to search notes")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to search notes",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Notes search completed successfully",
		Data:    notes,
	})
}

// GetNotesByTag godoc
// @Summary Get notes by tag
// @Description Get notes filtered by specific tags
// @Tags notes
// @Produce json
// @Param tags query string true "Tags (comma-separated)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Success 200 {object} models.APIResponse{data=models.NotesListResponse}
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/tags [get]
func (nc *NoteController) GetNotesByTag(c *gin.Context) {
	log := logger.GetLogger()
	tags := c.Query("tags")
	if tags == "" {
		c.JSON(http.StatusBadRequest, models.APIResponse{
			Success: false,
			Message: "Tags parameter is required",
		})
		return
	}

	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	notes, err := nc.noteService.GetNotesByTag(userID, tags, page, pageSize)
	if err != nil {
		log.WithError(err).Error("Failed to get notes by tag")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to retrieve notes by tag",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Notes retrieved successfully",
		Data:    notes,
	})
}

// GetNotesCount godoc
// @Summary Get notes count
// @Description Get the total count of notes for the authenticated user
// @Tags notes
// @Produce json
// @Success 200 {object} models.APIResponse{data=map[string]int64}
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Security BearerAuth
// @Router /notes/count [get]
func (nc *NoteController) GetNotesCount(c *gin.Context) {
	log := logger.GetLogger()
	userID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Message: "Unauthorized",
		})
		return
	}

	count, err := nc.noteService.GetNotesCount(userID)
	if err != nil {
		log.WithError(err).Error("Failed to get notes count")
		c.JSON(http.StatusInternalServerError, models.APIResponse{
			Success: false,
			Message: "Failed to get notes count",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Notes count retrieved successfully",
		Data: map[string]int64{
			"count": count,
		},
	})
}

// Helper function to format validation errors
func formatValidationErrors(errs validator.ValidationErrors) []models.ValidationError {
	var errors []models.ValidationError
	for _, err := range errs {
		errors = append(errors, models.ValidationError{
			Field:   err.Field(),
			Message: getValidationMessage(err),
		})
	}
	return errors
}

func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "email":
		return err.Field() + " must be a valid email"
	case "min":
		return err.Field() + " must be at least " + err.Param() + " characters"
	case "max":
		return err.Field() + " must be at most " + err.Param() + " characters"
	default:
		return err.Field() + " is invalid"
	}
}
