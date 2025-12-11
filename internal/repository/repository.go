package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/railgorail/kpfu-db-app/internal/domain"
)

// Repository holds the database connection pool.
type Repository struct {
	db *pgxpool.Pool
}

// New creates a new Repository.
func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetWarehouses retrieves all warehouses from the database.
func (r *Repository) GetWarehouses(ctx context.Context) ([]domain.Warehouse, error) {
	rows, err := r.db.Query(ctx, "SELECT warehouse_no, manager_surname FROM warehouses")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []domain.Warehouse
	for rows.Next() {
		var w domain.Warehouse
		if err := rows.Scan(&w.WarehouseNo, &w.ManagerSurname); err != nil {
			return nil, err
		}
		warehouses = append(warehouses, w)
	}
	return warehouses, nil
}

// GetContracts retrieves all contracts from the database.
func (r *Repository) GetContracts(ctx context.Context) ([]domain.Contract, error) {
	rows, err := r.db.Query(ctx, "SELECT contract_no, part_code, unit, start_date, end_date, plan_qty, contract_price FROM contracts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contracts []domain.Contract
	for rows.Next() {
		var c domain.Contract
		if err := rows.Scan(&c.ContractNo, &c.PartCode, &c.Unit, &c.StartDate, &c.EndDate, &c.PlanQty, &c.ContractPrice); err != nil {
			return nil, err
		}
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetDeliveries retrieves all deliveries from the database.
func (r *Repository) GetDeliveries(ctx context.Context) ([]domain.Delivery, error) {
	rows, err := r.db.Query(ctx, "SELECT warehouse_no, receipt_doc_no, contract_no, part_code, unit, qty, received_date FROM deliveries")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deliveries []domain.Delivery
	for rows.Next() {
		var d domain.Delivery
		if err := rows.Scan(&d.WarehouseNo, &d.ReceiptDocNo, &d.ContractNo, &d.PartCode, &d.Unit, &d.Qty, &d.ReceivedDate); err != nil {
			return nil, err
		}
		deliveries = append(deliveries, d)
	}
	return deliveries, nil
}

func (r *Repository) GetView(ctx context.Context) ([]domain.View, error) {
	rows, err := r.db.Query(ctx, `
		SELECT * FROM full_deliveries_view;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var view []domain.View
	for rows.Next() {
		var v domain.View
		err := rows.Scan(
			&v.WarehouseNo,
			&v.ManagerSurname,
			&v.ReceiptDocNo,
			&v.ReceivedDate,
			&v.Qty,
			&v.DeliveryUnit,
			&v.ContractNo,
			&v.PartCode,
			&v.ContractUnit,
			&v.StartDate,
			&v.EndDate,
			&v.PlanQty,
			&v.ContractPrice,
		)
		if err != nil {
			return nil, err
		}
		view = append(view, v)
	}
	return view, nil
}
func (r *Repository) GetTask1(ctx context.Context, price float64) ([]domain.Task1, error) {
	// #region agent log
	if f, err := os.OpenFile("/Users/rail/Documents/life/edu/kpfu/3/db/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		json.NewEncoder(f).Encode(map[string]interface{}{"location": "repository.go:113", "message": "GetTask1 entry", "data": map[string]interface{}{"price": price}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "A"})
		f.Close()
	}
	// #endregion
	rows, err := r.db.Query(ctx, `
		SELECT d.warehouse_no, d.part_code, d.receipt_doc_no, d.received_date, d.qty, d.contract_no, c.contract_price
		FROM deliveries d
		JOIN contracts c
		ON d.contract_no = c.contract_no
		WHERE contract_price > $1
		ORDER BY d.received_date;
	`, price)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var task1 []domain.Task1
	for rows.Next() {
		var t domain.Task1
		err := rows.Scan(
			&t.WarehouseNo,
			&t.PartCode,
			&t.ReceiptDocNo,
			&t.ReceivedDate,
			&t.Qty,
			&t.ContractNo,
			&t.ContractPrice,
		)
		if err != nil {
			return nil, err
		}
		// #region agent log
		if f, err := os.OpenFile("/Users/rail/Documents/life/edu/kpfu/3/db/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			json.NewEncoder(f).Encode(map[string]interface{}{"location": "repository.go:142", "message": "GetTask1 row", "data": map[string]interface{}{"contract_no": t.ContractNo, "part_code": t.PartCode, "contract_price": t.ContractPrice}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "A"})
			f.Close()
		}
		// #endregion
		task1 = append(task1, t)
	}
	// #region agent log
	if f, err := os.OpenFile("/Users/rail/Documents/life/edu/kpfu/3/db/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		json.NewEncoder(f).Encode(map[string]interface{}{"location": "repository.go:144", "message": "GetTask1 exit", "data": map[string]interface{}{"result_count": len(task1)}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "A"})
		f.Close()
	}
	// #endregion
	return task1, nil
}

func (r *Repository) GetTask2(ctx context.Context) ([]domain.Task2, error) {
	rows, err := r.db.Query(ctx, `
		SELECT contract_no, part_code, plan_qty, end_date, 
			SUM(plan_qty) OVER(PARTITION BY contract_no) AS total_plan_qty,
			DENSE_RANK() OVER (ORDER BY end_date) AS priority
		FROM contracts
		WHERE contract_price > 100
		ORDER BY end_date;
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var task2 []domain.Task2
	for rows.Next() {
		var t domain.Task2
		err := rows.Scan(
			&t.ContractNo,
			&t.PartCode,
			&t.PlanQty,
			&t.EndDate,
			&t.SumQty,
			&t.Priotity,
		)
		if err != nil {
			return nil, err
		}
		task2 = append(task2, t)
	}
	return task2, nil
}

func (r *Repository) GetTask3(ctx context.Context, planQty, deliveryQty int) ([]domain.Contract, error) {
	rows, err := r.db.Query(ctx, `SELECT *
	FROM contracts c
	WHERE c.plan_qty > $1
	AND EXISTS (
		SELECT 1
		FROM deliveries d
		WHERE d.contract_no = c.contract_no
		AND d.part_code = c.part_code
			AND $2 < ALL (
				SELECT d2.qty
				FROM deliveries d2
				WHERE d2.contract_no = c.contract_no
				AND d2.part_code = c.part_code
				AND d2.warehouse_no = d.warehouse_no
        )
  );

	`, planQty, deliveryQty)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var task3 []domain.Contract
	for rows.Next() {
		var t domain.Contract
		err := rows.Scan(
			&t.ContractNo,
			&t.PartCode,
			&t.Unit,
			&t.StartDate,
			&t.EndDate,
			&t.PlanQty,
			&t.ContractPrice,
		)
		if err != nil {
			return nil, err
		}
		task3 = append(task3, t)
	}
	return task3, nil
}

// UpdateWarehouse updates a warehouse in the database.
func (r *Repository) UpdateWarehouse(ctx context.Context, warehouseNo int, managerSurname string) error {
	_, err := r.db.Exec(ctx, "UPDATE warehouses SET manager_surname = $1 WHERE warehouse_no = $2", managerSurname, warehouseNo)
	return err
}

// UpdateContract updates a contract in the database.
func (r *Repository) UpdateContract(ctx context.Context, contractNo int, partCode string, unit string, startDate, endDate string, planQty, contractPrice float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE contracts 
		SET unit = $1, start_date = $2, end_date = $3, plan_qty = $4, contract_price = $5 
		WHERE contract_no = $6 AND part_code = $7
	`, unit, startDate, endDate, planQty, contractPrice, contractNo, partCode)
	return err
}

// UpdateDelivery updates a delivery in the database.
func (r *Repository) UpdateDelivery(ctx context.Context, warehouseNo, receiptDocNo int, contractNo int, partCode, unit string, qty float64, receivedDate string) error {
	// #region agent log
	if f, err := os.OpenFile("/Users/rail/Documents/life/edu/kpfu/3/db/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		json.NewEncoder(f).Encode(map[string]interface{}{"location": "repository.go:241", "message": "UpdateDelivery entry", "data": map[string]interface{}{"warehouseNo": warehouseNo, "receiptDocNo": receiptDocNo, "contractNo": contractNo, "partCode": partCode}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
		f.Close()
	}
	// #endregion
	// Parse the received date
	parsedDate, err := time.Parse("2006-01-02", receivedDate)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	// Check if the receivedDate is within the contract's date interval
	var startDate, endDate time.Time
	err = r.db.QueryRow(ctx, `
		SELECT start_date, end_date 
		FROM contracts 
		WHERE contract_no = $1 AND part_code = $2
	`, contractNo, partCode).Scan(&startDate, &endDate)
	if err != nil {
		// #region agent log
		if f, err2 := os.OpenFile("/Users/rail/Documents/life/edu/kpfu/3/db/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
			json.NewEncoder(f).Encode(map[string]interface{}{"location": "repository.go:256", "message": "UpdateDelivery contract not found", "data": map[string]interface{}{"contractNo": contractNo, "partCode": partCode, "error": err.Error()}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
			f.Close()
		}
		// #endregion
		return fmt.Errorf("contract not found: %w", err)
	}
	// #region agent log
	if f, err2 := os.OpenFile("/Users/rail/Documents/life/edu/kpfu/3/db/.cursor/debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err2 == nil {
		json.NewEncoder(f).Encode(map[string]interface{}{"location": "repository.go:258", "message": "UpdateDelivery contract found", "data": map[string]interface{}{"contractNo": contractNo, "partCode": partCode, "startDate": startDate.Format("2006-01-02"), "endDate": endDate.Format("2006-01-02")}, "timestamp": time.Now().UnixMilli(), "sessionId": "debug-session", "runId": "run1", "hypothesisId": "C"})
		f.Close()
	}
	// #endregion
	fmt.Println(startDate, endDate, parsedDate)

	// Verify that receivedDate is within the contract date interval
	if parsedDate.Before(startDate) || parsedDate.After(endDate) {
		return fmt.Errorf("received_date %s is outside the contract date interval [%s, %s]",
			receivedDate, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	// Update the delivery
	_, err = r.db.Exec(ctx, `
		UPDATE deliveries 
		SET contract_no = $1, part_code = $2, unit = $3, qty = $4, received_date = $5 
		WHERE warehouse_no = $6 AND receipt_doc_no = $7
	`, contractNo, partCode, unit, qty, receivedDate, warehouseNo, receiptDocNo)
	return err
}

// CreateWarehouse creates a new warehouse in the database.
func (r *Repository) CreateWarehouse(ctx context.Context, managerSurname string) (int, error) {
	var warehouseNo int
	err := r.db.QueryRow(ctx, "INSERT INTO warehouses (manager_surname) VALUES ($1) RETURNING warehouse_no", managerSurname).Scan(&warehouseNo)
	return warehouseNo, err
}

// CreateContract creates a new contract in the database.
func (r *Repository) CreateContract(ctx context.Context, contractNo int, partCode string, unit string, startDate, endDate string, planQty, contractPrice float64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO contracts (contract_no, part_code, unit, start_date, end_date, plan_qty, contract_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, contractNo, partCode, unit, startDate, endDate, planQty, contractPrice)
	return err
}

// CreateDelivery creates a new delivery in the database.
func (r *Repository) CreateDelivery(ctx context.Context, warehouseNo, receiptDocNo int, contractNo int, partCode, unit string, qty float64, receivedDate string) error {
	// Parse the received date
	parsedDate, err := time.Parse("2006-01-02", receivedDate)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	// Check if the receivedDate is within the contract's date interval
	var startDate, endDate time.Time
	err = r.db.QueryRow(ctx, `
		SELECT start_date, end_date 
		FROM contracts 
		WHERE contract_no = $1 AND part_code = $2
	`, contractNo, partCode).Scan(&startDate, &endDate)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}
	fmt.Println(startDate, endDate, parsedDate)

	// Verify that receivedDate is within the contract date interval
	if parsedDate.Before(startDate) || parsedDate.After(endDate) {
		return fmt.Errorf("received_date %s is outside the contract date interval [%s, %s]",
			receivedDate, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}
	_, err = r.db.Exec(ctx, `
		INSERT INTO deliveries (warehouse_no, receipt_doc_no, contract_no, part_code, unit, qty, received_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, warehouseNo, receiptDocNo, contractNo, partCode, unit, qty, receivedDate)
	return err
}

func (r *Repository) CallContractSummary(ctx context.Context, contractNo int, partCode string) (*domain.ContractSummary, error) {
	var result domain.ContractSummary
	result.ContractNo = contractNo
	result.PartCode = partCode

	// Escape single quotes in partCode for SQL
	escapedPartCode := fmt.Sprintf("'%s'", strings.ReplaceAll(partCode, "'", "''"))
	query := fmt.Sprintf(`
		DO $$
		DECLARE
			v_total_delivered DECIMAL(10,2);
			v_contract_price DECIMAL(10,2);
		BEGIN
			CALL p_contract_summary(%d, %s, v_total_delivered, v_contract_price);
			DELETE FROM proc_result;
			INSERT INTO proc_result VALUES (v_total_delivered, v_contract_price);
		END $$;
	`, contractNo, escapedPartCode)
	_, err := r.db.Exec(ctx, query)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRow(ctx, `SELECT total_delivered, contract_price FROM proc_result`).Scan(
		&result.TotalDelivered,
		&result.ContractPrice,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
