package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/tehsis/logmeup-api/internal/handlers"
	"github.com/tehsis/logmeup-api/internal/middleware"
)

// WebSocketHub interface for the hub
type WebSocketHub interface {
	HandleWebSocket(c *gin.Context)
}

func SetupRoutes(r *gin.Engine, noteHandler *handlers.NoteHandler, actionHandler *handlers.ActionHandler, wsHub WebSocketHub) {
	// WebSocket endpoint (no auth required for now)
	r.GET("/ws", wsHub.HandleWebSocket)

	// Protected API routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		// Notes routes
		notes := api.Group("/notes")
		{
			notes.POST("", noteHandler.Create)
			notes.GET("/:id", noteHandler.GetByID)
			notes.GET("", noteHandler.GetByDate)
			notes.PUT("/:id", noteHandler.Update)
			notes.DELETE("/:id", noteHandler.Delete)
		}

		// Actions routes
		actions := api.Group("/actions")
		{
			actions.POST("", actionHandler.Create)
			actions.GET("", actionHandler.GetAll)
			actions.GET("/:id", actionHandler.GetByID)
			actions.GET("/note/:note_id", actionHandler.GetByNoteID)
			actions.PUT("/:id", actionHandler.Update)
			actions.DELETE("/:id", actionHandler.Delete)
			actions.HEAD("", actionHandler.Health)
		}
	}
}
