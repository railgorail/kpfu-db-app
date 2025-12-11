package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	// API routes for updates
	api := r.Group("/api")
	api.PUT("/warehouses", h.UpdateWarehouse)
	api.PUT("/contracts", h.UpdateContract)
	api.PUT("/deliveries", h.UpdateDelivery)
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
	priceStr := c.DefaultQuery("price", "100")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		price = 100
	}

	t, err := h.repo.GetTask1(c.Request.Context(), price)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching task1 data: %v", err)
		return
	}
	c.HTML(http.StatusOK, "task1.html", gin.H{
		"Title": "Task 1",
		"Task1": t,
		"Price": price,
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
	planQtyStr := c.DefaultQuery("plan_qty", "1000")
	planQty, err := strconv.Atoi(planQtyStr)
	if err != nil {
		planQty = 1000
	}

	deliveryQtyStr := c.DefaultQuery("delivery_qty", "50")
	deliveryQty, err := strconv.Atoi(deliveryQtyStr)
	if err != nil {
		deliveryQty = 50
	}

	t, err := h.repo.GetTask3(c.Request.Context(), planQty, deliveryQty)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching task3 data: %v", err)
		return
	}
	c.HTML(http.StatusOK, "task3.html", gin.H{
		"Title":       "Task 3",
		"Task3":       t,
		"PlanQty":     planQty,
		"DeliveryQty": deliveryQty,
	})
}

// UpdateWarehouse handles updating a warehouse.
func (h *Handler) UpdateWarehouse(c *gin.Context) {
	var req struct {
		ID             int    `json:"id" binding:"required"`
		ManagerSurname string `json:"manager_surname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateWarehouse(c.Request.Context(), req.ID, req.ManagerSurname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update warehouse: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Warehouse updated successfully"})
}

// UpdateContract handles updating a contract.
func (h *Handler) UpdateContract(c *gin.Context) {
	var req struct {
		ContractNo    int     `json:"contract_no" binding:"required"`
		PartCode      string  `json:"part_code" binding:"required"`
		Unit          string  `json:"unit" binding:"required"`
		StartDate     string  `json:"start_date" binding:"required"`
		EndDate       string  `json:"end_date" binding:"required"`
		PlanQty       float64 `json:"plan_qty" binding:"required"`
		ContractPrice float64 `json:"contract_price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate dates
	if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}
	if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if err := h.repo.UpdateContract(c.Request.Context(), req.ContractNo, req.PartCode, req.Unit, req.StartDate, req.EndDate, req.PlanQty, req.ContractPrice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update contract: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contract updated successfully"})
}

// UpdateDelivery handles updating a delivery.
func (h *Handler) UpdateDelivery(c *gin.Context) {
	var req struct {
		WarehouseNo  int     `json:"warehouse_no" binding:"required"`
		ReceiptDocNo int     `json:"receipt_doc_no" binding:"required"`
		ContractNo   int     `json:"contract_no" binding:"required"`
		PartCode     string  `json:"part_code" binding:"required"`
		Unit         string  `json:"unit" binding:"required"`
		Qty          float64 `json:"qty" binding:"required"`
		ReceivedDate string  `json:"received_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate date
	if _, err := time.Parse("2006-01-02", req.ReceivedDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid received_date format. Use YYYY-MM-DD"})
		return
	}

	if err := h.repo.UpdateDelivery(c.Request.Context(), req.WarehouseNo, req.ReceiptDocNo, req.ContractNo, req.PartCode, req.Unit, req.Qty, req.ReceivedDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to update delivery: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery updated successfully"})
}
