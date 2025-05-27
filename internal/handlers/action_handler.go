package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tehsis/logmeup-api/internal/middleware"
	"github.com/tehsis/logmeup-api/internal/models"
	"github.com/tehsis/logmeup-api/internal/repository"
)

// WebSocketHub interface for broadcasting
type WebSocketHub interface {
	BroadcastActionCreated(action *models.Action)
	BroadcastActionUpdated(action *models.Action)
	BroadcastActionDeleted(actionID int64)
}

type ActionHandler struct {
	repo *repository.ActionRepository
	hub  WebSocketHub
}

func NewActionHandler(repo *repository.ActionRepository, hub WebSocketHub) *ActionHandler {
	log.Printf("[ActionHandler] Initializing action handler with WebSocket support")
	return &ActionHandler{
		repo: repo,
		hub:  hub,
	}
}

// Helper function to log request details
func logRequest(c *gin.Context, operation string, details ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	log.Printf("[ActionHandler-%s] %s | IP: %s | User-Agent: %s | Details: %v",
		operation, timestamp, clientIP, userAgent, details)
}

// Helper function to log errors with context
func logError(c *gin.Context, operation string, err error, details ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	clientIP := c.ClientIP()

	log.Printf("[ActionHandler-%s-ERROR] %s | IP: %s | Error: %v | Details: %v",
		operation, timestamp, clientIP, err, details)
}

// Helper function to log successful operations
func logSuccess(c *gin.Context, operation string, details ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	log.Printf("[ActionHandler-%s-SUCCESS] %s | Details: %v",
		operation, timestamp, details)
}

func (h *ActionHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	logRequest(c, "Create", "Starting action creation for user", userID)

	var req models.CreateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logError(c, "Create", err, "Failed to bind JSON request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "INVALID_JSON",
		})
		return
	}

	logRequest(c, "Create", "Request data", map[string]interface{}{
		"user_id":     userID,
		"note_id":     req.NoteID,
		"description": req.Description,
	})

	action, err := h.repo.Create(userID, &req)
	if err != nil {
		logError(c, "Create", err, "Database creation failed", req)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "Create", "Action created successfully", map[string]interface{}{
		"action_id":   action.ID,
		"user_id":     action.UserID,
		"note_id":     action.NoteID,
		"description": action.Description,
	})

	h.hub.BroadcastActionCreated(action)

	c.JSON(http.StatusCreated, action)
}

func (h *ActionHandler) GetByID(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idParam := c.Param("id")
	logRequest(c, "GetByID", "Fetching action by ID", idParam, "for user", userID)

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logError(c, "GetByID", err, "Invalid ID parameter", idParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
			"code":  "INVALID_ID",
		})
		return
	}

	action, err := h.repo.GetByID(userID, id)
	if err != nil {
		logError(c, "GetByID", err, "Action not found in database", id, "for user", userID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "action not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	logSuccess(c, "GetByID", "Action retrieved successfully", map[string]interface{}{
		"action_id":   action.ID,
		"user_id":     action.UserID,
		"description": action.Description,
		"completed":   action.Completed,
	})

	c.JSON(http.StatusOK, action)
}

func (h *ActionHandler) GetAll(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	logRequest(c, "GetAll", "Fetching all actions for user", userID)

	actions, err := h.repo.GetAll(userID)
	if err != nil {
		logError(c, "GetAll", err, "Failed to retrieve actions from database for user", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "GetAll", "Actions retrieved successfully", map[string]interface{}{
		"user_id": userID,
		"count":   len(actions),
	})

	c.JSON(http.StatusOK, actions)
}

func (h *ActionHandler) GetByNoteID(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	noteIDParam := c.Param("note_id")
	logRequest(c, "GetByNoteID", "Fetching actions by note ID", noteIDParam, "for user", userID)

	noteID, err := strconv.ParseInt(noteIDParam, 10, 64)
	if err != nil {
		logError(c, "GetByNoteID", err, "Invalid note ID parameter", noteIDParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid note_id",
			"code":  "INVALID_NOTE_ID",
		})
		return
	}

	actions, err := h.repo.GetByNoteID(userID, noteID)
	if err != nil {
		logError(c, "GetByNoteID", err, "Failed to retrieve actions by note ID", noteID, "for user", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "GetByNoteID", "Actions retrieved successfully", map[string]interface{}{
		"user_id": userID,
		"note_id": noteID,
		"count":   len(actions),
	})

	c.JSON(http.StatusOK, actions)
}

func (h *ActionHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idParam := c.Param("id")
	logRequest(c, "Update", "Starting action update", idParam, "for user", userID)

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logError(c, "Update", err, "Invalid ID parameter", idParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
			"code":  "INVALID_ID",
		})
		return
	}

	var req models.UpdateActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logError(c, "Update", err, "Failed to bind JSON request")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "INVALID_JSON",
		})
		return
	}

	logRequest(c, "Update", "Request data", map[string]interface{}{
		"user_id":   userID,
		"action_id": id,
		"completed": req.Completed,
	})

	action, err := h.repo.Update(userID, id, &req)
	if err != nil {
		logError(c, "Update", err, "Database update failed", id, "for user", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "Update", "Action updated successfully", map[string]interface{}{
		"action_id": action.ID,
		"user_id":   action.UserID,
		"completed": action.Completed,
	})

	h.hub.BroadcastActionUpdated(action)

	c.JSON(http.StatusOK, action)
}

func (h *ActionHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	idParam := c.Param("id")
	logRequest(c, "Delete", "Starting action deletion", idParam, "for user", userID)

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logError(c, "Delete", err, "Invalid ID parameter", idParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
			"code":  "INVALID_ID",
		})
		return
	}

	if err := h.repo.Delete(userID, id); err != nil {
		logError(c, "Delete", err, "Database deletion failed", id, "for user", userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "Delete", "Action deleted successfully", map[string]interface{}{
		"action_id": id,
		"user_id":   userID,
	})

	h.hub.BroadcastActionDeleted(id)

	c.Status(http.StatusNoContent)
}

func (h *ActionHandler) Health(c *gin.Context) {
	logRequest(c, "Health", "Health check requested")

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "action-handler",
	})
}
