package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/tehsis/logmeup-api/internal/handlers"
	"github.com/tehsis/logmeup-api/internal/repository"
	"github.com/tehsis/logmeup-api/internal/routes"
	"github.com/tehsis/logmeup-api/pkg/config"
	"github.com/tehsis/logmeup-api/pkg/database"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDBConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	noteRepo := repository.NewNoteRepository(db)
	actionRepo := repository.NewActionRepository(db)

	// Initialize handlers
	noteHandler := handlers.NewNoteHandler(noteRepo)
	actionHandler := handlers.NewActionHandler(actionRepo)

	// Initialize router
	r := gin.Default()

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Setup routes
	routes.SetupRoutes(r, noteHandler, actionHandler)

	// Start server
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
