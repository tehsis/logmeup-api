package handlers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	logRequest(c, "Create", "Starting action creation")

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
		"note_id":     req.NoteID,
		"description": req.Description,
	})

	action, err := h.repo.Create(&req)
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
		"note_id":     action.NoteID,
		"description": action.Description,
	})

	h.hub.BroadcastActionCreated(action)

	c.JSON(http.StatusCreated, action)
}

func (h *ActionHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	logRequest(c, "GetByID", "Fetching action by ID", idParam)

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logError(c, "GetByID", err, "Invalid ID parameter", idParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
			"code":  "INVALID_ID",
		})
		return
	}

	action, err := h.repo.GetByID(id)
	if err != nil {
		logError(c, "GetByID", err, "Action not found in database", id)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "action not found",
			"code":  "NOT_FOUND",
		})
		return
	}

	logSuccess(c, "GetByID", "Action retrieved successfully", map[string]interface{}{
		"action_id":   action.ID,
		"description": action.Description,
		"completed":   action.Completed,
	})

	c.JSON(http.StatusOK, action)
}

func (h *ActionHandler) GetAll(c *gin.Context) {
	logRequest(c, "GetAll", "Fetching all actions")

	actions, err := h.repo.GetAll()
	if err != nil {
		logError(c, "GetAll", err, "Failed to retrieve actions from database")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "GetAll", "Actions retrieved successfully", map[string]interface{}{
		"count": len(actions),
	})

	c.JSON(http.StatusOK, actions)
}

func (h *ActionHandler) GetByNoteID(c *gin.Context) {
	noteIDParam := c.Param("note_id")
	logRequest(c, "GetByNoteID", "Fetching actions by note ID", noteIDParam)

	noteID, err := strconv.ParseInt(noteIDParam, 10, 64)
	if err != nil {
		logError(c, "GetByNoteID", err, "Invalid note ID parameter", noteIDParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid note_id",
			"code":  "INVALID_NOTE_ID",
		})
		return
	}

	actions, err := h.repo.GetByNoteID(noteID)
	if err != nil {
		logError(c, "GetByNoteID", err, "Failed to retrieve actions by note ID", noteID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "GetByNoteID", "Actions retrieved successfully", map[string]interface{}{
		"note_id": noteID,
		"count":   len(actions),
	})

	c.JSON(http.StatusOK, actions)
}

func (h *ActionHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	logRequest(c, "Update", "Starting action update", idParam)

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
		logError(c, "Update", err, "Failed to bind JSON request", id)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
			"code":  "INVALID_JSON",
		})
		return
	}

	logRequest(c, "Update", "Update data", map[string]interface{}{
		"action_id": id,
		"completed": req.Completed,
	})

	action, err := h.repo.Update(id, &req)
	if err != nil {
		logError(c, "Update", err, "Database update failed", map[string]interface{}{
			"action_id": id,
			"completed": req.Completed,
		})
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "Update", "Action updated successfully", map[string]interface{}{
		"action_id":  action.ID,
		"completed":  action.Completed,
		"updated_at": action.UpdatedAt,
	})

	h.hub.BroadcastActionUpdated(action)

	c.JSON(http.StatusOK, action)
}

func (h *ActionHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	logRequest(c, "Delete", "Starting action deletion", idParam)

	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		logError(c, "Delete", err, "Invalid ID parameter", idParam)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid id",
			"code":  "INVALID_ID",
		})
		return
	}

	if err := h.repo.Delete(id); err != nil {
		logError(c, "Delete", err, "Database deletion failed", id)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
			"code":  "DATABASE_ERROR",
		})
		return
	}

	logSuccess(c, "Delete", "Action deleted successfully", map[string]interface{}{
		"action_id": id,
	})

	h.hub.BroadcastActionDeleted(id)

	c.Status(http.StatusNoContent)
}

func (h *ActionHandler) Health(c *gin.Context) {
	logRequest(c, "Health", "Health check request")
	logSuccess(c, "Health", "Health check successful")
	c.Status(http.StatusOK)
}
