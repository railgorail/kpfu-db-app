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
	WarehouseNo  int       `json:"warehouse_no"`
	ReceiptDocNo int       `json:"receipt_doc_no"`
	ContractNo   int       `json:"contract_no"`
	PartCode     string    `json:"part_code"`
	Unit         string    `json:"unit"`
	Qty          float64   `json:"qty"`
	ReceivedDate time.Time `json:"received_date"`
}

type View struct {
	// From warehouses
	WarehouseNo    int    `json:"warehouse_no"`
	ManagerSurname string `json:"manager_surname"`

	// From deliveries
	ReceiptDocNo int       `json:"receipt_doc_no"`
	ReceivedDate time.Time `json:"received_date"`
	Qty          float64   `json:"qty"`
	DeliveryUnit string    `json:"delivery_unit"`

	ContractNo int    `json:"contract_no"`
	PartCode   string `json:"part_code"`

	// From contracts
	ContractUnit  string    `json:"contract_unit"`
	StartDate     time.Time `json:"start_date"`
	EndDate       time.Time `json:"end_date"`
	PlanQty       float64   `json:"plan_qty"`
	ContractPrice float64   `json:"contract_price"`
}

type Task2 struct {
	ContractNo int       `json:"contract_no"`
	PartCode   string    `json:"part_code"`
	PlanQty    float64   `json:"plan_qty"`
	EndDate    time.Time `json:"end_date"`
	SumQty     float64   `json:"sum_qty"`
	Priotity   int       `json:"priority"`
}
