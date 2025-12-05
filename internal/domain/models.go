package domain

import "time"

// Warehouse represents a warehouse in the database.
type Warehouse struct {
	WarehouseNo    int    `json:"warehouse_no"`
	ManagerSurname string `json:"manager_surname"`
}

// Contract represents a contract in the database.
type Contract struct {
	ContractNo    int       `json:"contract_no"`
	PartCode      string    `json:"part_code"`
	Unit          string    `json:"unit"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	PlanQty       float64   `json:"plan_qty"`
	ContractPrice float64   `json:"contract_price"`
}

// Delivery represents a delivery in the database.
type Delivery struct {
	WarehouseNo  int     `json:"warehouse_no"`
	ReceiptDocNo int     `json:"receipt_doc_no"`
	ContractNo   int     `json:"contract_no"`
	PartCode     string  `json:"part_code"`
	Unit         string  `json:"unit"`
	Qty          float64 `json:"qty"`
	ReceivedDate time.Time  `json:"received_date"`
}
