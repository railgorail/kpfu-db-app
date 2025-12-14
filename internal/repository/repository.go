package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/railgorail/kpfu-db-app/internal/domain"
	"gorm.io/gorm"
)

type Repository struct {
	db     *pgxpool.Pool
	gormDB *gorm.DB
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func NewWithGORM(db *pgxpool.Pool, gormDB *gorm.DB) *Repository {
	return &Repository{db: db, gormDB: gormDB}
}

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
		task1 = append(task1, t)
	}
	return task1, nil
}

func (r *Repository) ORMGetTask1(ctx context.Context, price float64) ([]domain.Task1, error) {
	var deliveries []domain.Delivery
	err := r.gormDB.WithContext(ctx).
    Preload("Contract").
    Order("received_date").
    Find(&deliveries).Error
	if err != nil {
		return nil, err
	}

	var task1 []domain.Task1
	for _, d := range deliveries {
		if d.Contract.ContractPrice > price {
			task1 = append(task1, domain.Task1{
				WarehouseNo:   d.WarehouseNo,
				PartCode:      d.PartCode,
				ReceiptDocNo:  d.ReceiptDocNo,
				ReceivedDate:  d.ReceivedDate,
				Qty:           d.Qty,
				ContractNo:    d.ContractNo,
				ContractPrice: d.Contract.ContractPrice,
			})
		}
	}
	
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

func (r *Repository) UpdateWarehouse(ctx context.Context, warehouseNo int, managerSurname string) error {
	_, err := r.db.Exec(ctx, "UPDATE warehouses SET manager_surname = $1 WHERE warehouse_no = $2", managerSurname, warehouseNo)
	return err
}

func (r *Repository) UpdateContract(ctx context.Context, contractNo int, partCode string, unit string, startDate, endDate string, planQty, contractPrice float64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE contracts 
		SET unit = $1, start_date = $2, end_date = $3, plan_qty = $4, contract_price = $5 
		WHERE contract_no = $6 AND part_code = $7
	`, unit, startDate, endDate, planQty, contractPrice, contractNo, partCode)
	return err
}

func (r *Repository) UpdateDelivery(ctx context.Context, warehouseNo, receiptDocNo int, contractNo int, partCode, unit string, qty float64, receivedDate string) error {
	parsedDate, err := time.Parse("2006-01-02", receivedDate)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	var startDate, endDate time.Time
	err = r.db.QueryRow(ctx, `
		SELECT start_date, end_date 
		FROM contracts 
		WHERE contract_no = $1 AND part_code = $2
	`, contractNo, partCode).Scan(&startDate, &endDate)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

	if parsedDate.Before(startDate) || parsedDate.After(endDate) {
		return fmt.Errorf("received_date %s is outside the contract date interval [%s, %s]",
			receivedDate, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	}

	_, err = r.db.Exec(ctx, `
		UPDATE deliveries 
		SET contract_no = $1, part_code = $2, unit = $3, qty = $4, received_date = $5 
		WHERE warehouse_no = $6 AND receipt_doc_no = $7
	`, contractNo, partCode, unit, qty, receivedDate, warehouseNo, receiptDocNo)
	return err
}

func (r *Repository) CreateWarehouse(ctx context.Context, managerSurname string) (int, error) {
	var warehouseNo int
	err := r.db.QueryRow(ctx, "INSERT INTO warehouses (manager_surname) VALUES ($1) RETURNING warehouse_no", managerSurname).Scan(&warehouseNo)
	return warehouseNo, err
}

func (r *Repository) CreateContract(ctx context.Context, contractNo int, partCode string, unit string, startDate, endDate string, planQty, contractPrice float64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO contracts (contract_no, part_code, unit, start_date, end_date, plan_qty, contract_price)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, contractNo, partCode, unit, startDate, endDate, planQty, contractPrice)
	return err
}

func (r *Repository) CreateDelivery(ctx context.Context, warehouseNo, receiptDocNo int, contractNo int, partCode, unit string, qty float64, receivedDate string) error {
	parsedDate, err := time.Parse("2006-01-02", receivedDate)
	if err != nil {
		return fmt.Errorf("invalid date format: %w", err)
	}

	var startDate, endDate time.Time
	err = r.db.QueryRow(ctx, `
		SELECT start_date, end_date 
		FROM contracts 
		WHERE contract_no = $1 AND part_code = $2
	`, contractNo, partCode).Scan(&startDate, &endDate)
	if err != nil {
		return fmt.Errorf("contract not found: %w", err)
	}

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

func (r *Repository) DeleteWarehouse(ctx context.Context, warehouseNo int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM warehouses WHERE warehouse_no = $1", warehouseNo)
	return err
}

func (r *Repository) DeleteContract(ctx context.Context, contractNo int, partCode string) error {
	_, err := r.db.Exec(ctx, "DELETE FROM contracts WHERE contract_no = $1 AND part_code = $2", contractNo, partCode)
	return err
}

func (r *Repository) DeleteDelivery(ctx context.Context, warehouseNo int, receiptDocNo int) error {
	_, err := r.db.Exec(ctx, "DELETE FROM deliveries WHERE warehouse_no = $1 AND receipt_doc_no = $2", warehouseNo, receiptDocNo)
	return err
}
