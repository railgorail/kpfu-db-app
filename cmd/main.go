package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/railgorail/kpfu-db-app/internal/config"
	"github.com/railgorail/kpfu-db-app/internal/database"
	"github.com/railgorail/kpfu-db-app/internal/handler"
	"github.com/railgorail/kpfu-db-app/internal/repository"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to the database
	dbpool, err := database.NewConnection(cfg.DBURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer dbpool.Close()

	// Create repository and handler
	repo := repository.New(dbpool)
	h := handler.New(repo)

	// Set up router
	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./static")
	h.RegisterRoutes(r)

	// Start server
	fmt.Println("Starting server on :80")
	if err := r.Run(":80"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
