package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/railgorail/kpfu-db-app/internal/config"
	"github.com/railgorail/kpfu-db-app/internal/database"
	"github.com/railgorail/kpfu-db-app/internal/handler"
	"github.com/railgorail/kpfu-db-app/internal/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	dbpool, err := database.NewConnection(cfg.DBURL)
	if err != nil {
		log.Fatalf("could not connect to database: %v", err)
	}
	defer dbpool.Close()

	gormDB, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("could not connect to database with GORM: %v", err)
	}

	repo := repository.NewWithGORM(dbpool, gormDB)
	h := handler.New(repo)

	r := gin.Default()
	r.LoadHTMLGlob("web/templates/*.html")
	r.Static("/static", "./static")
	h.RegisterRoutes(r)

	fmt.Println("Starting server on :80")
	if err := r.Run(":80"); err != nil {
		log.Fatalf("could not run server: %v", err)
	}
}
