package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/railgorail/kpfu-db-app/internal/domain"
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
	r.GET("/task-1", h.Task1)
	r.GET("/task-2", h.Task2)
	r.GET("/task-3", h.Task3)
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

	// Dummy data for testing
	if len(warehouses) == 0 {
		warehouses = []domain.Warehouse{
			{WarehouseNo: 99, ManagerSurname: "Test Warehouse"},
		}
	}
	if len(contracts) == 0 {
		contracts = []domain.Contract{
			{ContractNo: 999, PartCode: "TEST", Unit: "pcs", StartDate: time.Now(), EndDate: time.Now(), PlanQty: 100, ContractPrice: 9.99},
		}
	}
	if len(deliveries) == 0 {
		deliveries = []domain.Delivery{
			{WarehouseNo: 99, ReceiptDocNo: 9999, ContractNo: 999, PartCode: "TEST", Unit: "pcs", Qty: 50, ReceivedDate: time.Now()},
		}
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
	c.HTML(http.StatusOK, "view.html", gin.H{
		"Title": "View",
	})
}

// Task1 handles the task-1 page.
func (h *Handler) Task1(c *gin.Context) {
	c.HTML(http.StatusOK, "task.html", gin.H{
		"Title":    "Task 1",
		"TaskName": "Task 1",
	})
}

// Task2 handles the task-2 page.
func (h *Handler) Task2(c *gin.Context) {
	c.HTML(http.StatusOK, "task.html", gin.H{
		"Title":    "Task 2",
		"TaskName": "Task 2",
	})
}

// Task3 handles the task-3 page.
func (h *Handler) Task3(c *gin.Context) {
	c.HTML(http.StatusOK, "task.html", gin.H{
		"Title":    "Task 3",
		"TaskName": "Task 3",
	})
}
