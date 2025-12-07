package repository

import (
	"context"
	"fmt"

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
	rows, err := r.db.Query(ctx, `
		SELECT d.warehouse_no, d.part_code, d.receipt_doc_no, d.received_date, d.qty, d.contract_no, c.contract_price
		FROM deliveries d
		JOIN contracts c
		ON d.contract_no = c.contract_no
		WHERE contract_price > $1
		ORDER BY end_date;
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
	fmt.Println(task1)
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
	fmt.Println(task2)
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
	fmt.Println(task3)
	return task3, nil
}