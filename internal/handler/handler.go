package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/railgorail/kpfu-db-app/internal/domain"
	"github.com/railgorail/kpfu-db-app/internal/repository"
)

type Handler struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/", h.Home)
	r.GET("/view", h.View)
	tasks := r.Group("/task")
	tasks.GET("/1", h.Task1)
	tasks.GET("/2", h.Task2)
	tasks.GET("/3", h.Task3)

	api := r.Group("/api")
	api.PUT("/warehouses", h.UpdateWarehouse)
	api.PUT("/contracts", h.UpdateContract)
	api.PUT("/deliveries", h.UpdateDelivery)
	api.POST("/warehouses", h.CreateWarehouse)
	api.POST("/contracts", h.CreateContract)
	api.POST("/deliveries", h.CreateDelivery)

	r.GET("/procedure", h.Procedure)
	r.GET("/orm/task/1", h.ORMTask1)
}

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
}

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
}

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

func (h *Handler) ORMTask1(c *gin.Context) {
	priceStr := c.DefaultQuery("price", "100")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		price = 100
	}

	t, err := h.repo.ORMGetTask1(c.Request.Context(), price)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error fetching ORM task1 data: %v", err)
		return
	}
	c.HTML(http.StatusOK, "ORMtask1.html", gin.H{
		"Title": "ORM Task 1",
		"Task1": t,
		"Price": price,
	})
}

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

func (h *Handler) CreateWarehouse(c *gin.Context) {
	var req struct {
		ManagerSurname string `json:"manager_surname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	warehouseNo, err := h.repo.CreateWarehouse(c.Request.Context(), req.ManagerSurname)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create warehouse: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Warehouse created successfully", "warehouse_no": warehouseNo})
}

func (h *Handler) CreateContract(c *gin.Context) {
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

	if _, err := time.Parse("2006-01-02", req.StartDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}
	if _, err := time.Parse("2006-01-02", req.EndDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if err := h.repo.CreateContract(c.Request.Context(), req.ContractNo, req.PartCode, req.Unit, req.StartDate, req.EndDate, req.PlanQty, req.ContractPrice); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create contract: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Contract created successfully"})
}

func (h *Handler) CreateDelivery(c *gin.Context) {
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

	if _, err := time.Parse("2006-01-02", req.ReceivedDate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid received_date format. Use YYYY-MM-DD"})
		return
	}

	if err := h.repo.CreateDelivery(c.Request.Context(), req.WarehouseNo, req.ReceiptDocNo, req.ContractNo, req.PartCode, req.Unit, req.Qty, req.ReceivedDate); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create delivery: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery created successfully"})
}

func (h *Handler) Procedure(c *gin.Context) {
	contractNoStr := c.DefaultQuery("contract_no", "")
	partCode := c.DefaultQuery("part_code", "")

	var result *domain.ContractSummary

	procedureDescription := `Процедура p_contract_summary выполняет следующие действия:

1. Вычисляет сумму всех поставок (qty) из таблицы deliveries для указанного номера договора и кода детали
2. Получает договорную цену (contract_price) из таблицы contracts для указанного номера договора и кода детали
3. Если договор не найден, возвращает total_delivered = 0 и contract_price = NULL`

	if contractNoStr != "" && partCode != "" {
		contractNo, parseErr := strconv.Atoi(contractNoStr)
		if parseErr != nil {
			c.HTML(http.StatusOK, "procedure.html", gin.H{
				"Title":                "Procedure p_contract_summary",
				"ProcedureDescription": procedureDescription,
				"Error":                fmt.Sprintf("Invalid contract_no: %v", parseErr),
			})
			return
		}

		var err error
		result, err = h.repo.CallContractSummary(c.Request.Context(), contractNo, partCode)
		if err != nil {
			c.HTML(http.StatusOK, "procedure.html", gin.H{
				"Title":                "Procedure p_contract_summary",
				"ProcedureDescription": procedureDescription,
				"Error":                fmt.Sprintf("Error calling procedure: %v", err),
				"ContractNo":           contractNo,
				"PartCode":             partCode,
			})
			return
		}
	}

	c.HTML(http.StatusOK, "procedure.html", gin.H{
		"Title":                "Procedure p_contract_summary",
		"ProcedureDescription": procedureDescription,
		"Result":               result,
		"ContractNo":           contractNoStr,
		"PartCode":             partCode,
	})
}
