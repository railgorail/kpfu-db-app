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

func (r *Repository) GetTask2(ctx context.Context) ([]domain.Task2, error) {
	rows, err := r.db.Query(ctx, `
		SELECT contract_no, part_code, plan_qty, end_date, 
			SUM(plan_qty) OVER(PARTITION BY contract_no) AS total_plan_qty,
			DENSE_RANK() OVER (ORDER BY end_date) AS priority
		FROM contracts
		WHERE contract_price > 100;
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
