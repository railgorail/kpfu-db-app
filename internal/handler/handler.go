package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/railgorail/kpfu-db-app/internal/repository"
)

// Handler holds the repository.
type Handler struct {
	repo *repository.Repository
}

// New creates a new Handler.
func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes registers the routes for the application.
func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", h.Home)
	r.GET("/view", h.View)
	tasks := r.Group("/task")
	tasks.GET("/1", h.Task1)
	tasks.GET("/2", h.Task2)
	tasks.GET("/3", h.Task3)
}

// Home handles the home page.
func (h *Handler) Home(c *gin.Context) {
	warehouses, err := h.repo.GetWarehouses(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching warehouses: %v", err)
		return
	}
	contracts, err := h.repo.GetContracts(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching contracts: %v", err)
		return
	}
	deliveries, err := h.repo.GetDeliveries(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching deliveries: %v", err)
		return
	}

	c.HTML(http.StatusOK, "home.html", gin.H{
		"Title":      "Home",
		"Warehouses": warehouses,
		"Contracts":  contracts,
		"Deliveries": deliveries,
	})
	fmt.Println(warehouses, contracts, deliveries)
}

// View handles the view page.
func (h *Handler) View(c *gin.Context) {
	view, err := h.repo.GetView(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching view data: %v", err)
		return
	}
	c.HTML(http.StatusOK, "view.html", gin.H{
		"Title": "View",
		"View":  view,
	})
	fmt.Println(view)
}

// Task1 handles the task/1 page.
func (h *Handler) Task1(c *gin.Context) {
	c.HTML(http.StatusOK, "task.html", gin.H{
		"Title":    "Task 1",
		"TaskName": "Task 1",
	})
}

// Task2 handles the task-2 page.
func (h *Handler) Task2(c *gin.Context) {
	t, err := h.repo.GetTask2(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching task2 data: %v", err)
		return
	}
	c.HTML(http.StatusOK, "task2.html", gin.H{
		"Title": "Task 2",
		"Task2": t,
	})
}

// Task3 handles the task-3 page.
func (h *Handler) Task3(c *gin.Context) {
	c.HTML(http.StatusOK, "task.html", gin.H{
		"Title":    "Task 3",
		"TaskName": "Task 3",
	})
}
