/*

16а:Склад <–>> 13б:Учет поставок деталей <<–> 09в:Поставки деталей.

Склады 
•	номер склада; PK
•	фамилия материально ответственного лица.

Договорные поставки деталей
•	номер договора; PK
•	код детали; PK
•	единица измерения;
•	дата начала поставки;
•	дата завершения поставки;
•	план поставки (в количестве единиц измерения);
•	договорная цена за единицу.   

Учет поставок деталей
•	номер склада; PK
•	номер документа о приеме; PK
•	номер договора;
•	код покупной детали;
•	единица измерения;
•	количество покупных деталей;
•	дата поступления.,=
*/

BEGIN;

DROP TABLE IF EXISTS deliveries_audit CASCADE;
DROP TABLE IF EXISTS deliveries CASCADE;
DROP TABLE IF EXISTS contracts CASCADE;
DROP TABLE IF EXISTS warehouses CASCADE;

DROP FUNCTION IF EXISTS fn_cascade_delete_deliveries() CASCADE;
DROP FUNCTION IF EXISTS fn_log_delivery_insert() CASCADE;
DROP PROCEDURE IF EXISTS p_contract_summary(INT, TEXT, OUT DECIMAL(10,2), OUT DECIMAL(10,2));
DROP FUNCTION IF EXISTS fn_warehouse_count(fn_manager_surname text);
DROP FUNCTION IF EXISTS fn_deliveries_in_range(DATE, DATE);
DROP VIEW IF EXISTS full_deliveries_view CASCADE;


-- Склады
CREATE TABLE warehouses (
    warehouse_no         INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    manager_surname      TEXT NOT NULL 
);

-- Договорные поставки деталей 
CREATE TABLE contracts (
    contract_no          INT NOT NULL,
    part_code            TEXT NOT NULL,
    unit                 TEXT NOT NULL CHECK (unit IN ('pcs','kg','m','set')),
    start_date           DATE NOT NULL,
    end_date             DATE NOT NULL,
    plan_qty             DECIMAL(10,2) NOT NULL CHECK (plan_qty > 0),
    contract_price       DECIMAL(10,2) NOT NULL CHECK (contract_price >= 0),
    PRIMARY KEY (contract_no, part_code),
    CONSTRAINT chk_dates CHECK (start_date < end_date)
);

-- Учет поставок деталей
CREATE TABLE deliveries (
    warehouse_no         INT NOT NULL,
    receipt_doc_no       INT NOT NULL,
    contract_no          INT NOT NULL,
    part_code            TEXT NOT NULL,
    unit                 TEXT NOT NULL CHECK (unit IN ('pcs','kg','m','set')),
    qty                  DECIMAL(10,2) NOT NULL CHECK (qty > 0),
    received_date        DATE NOT NULL DEFAULT CURRENT_DATE,
    PRIMARY KEY (warehouse_no, receipt_doc_no),
    CONSTRAINT fk_delivery_warehouse FOREIGN KEY (warehouse_no) 
        REFERENCES warehouses(warehouse_no)
        ON DELETE CASCADE ON UPDATE cascade,
    CONSTRAINT fk_delivery_contract FOREIGN KEY (contract_no, part_code) 
        REFERENCES contracts(contract_no, part_code)
);

-- Таблица аудита для триггеров
CREATE TABLE deliveries_audit (
    audit_id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    warehouse_no    INT,
    receipt_doc_no  INT,
    contract_no     INT,
    part_code       TEXT,
    qty             DECIMAL(10,2),
    received_date   DATE,
    action          TEXT,
    action_time     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- Триггерная функция — ручной каскад удаления поставок при удалении договора
CREATE OR REPLACE FUNCTION fn_cascade_delete_deliveries() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    DELETE FROM deliveries 
    WHERE contract_no = OLD.contract_no AND part_code = OLD.part_code;
    RETURN OLD;
END;
$$;

DROP TRIGGER IF EXISTS trg_contracts_after_delete ON contracts;
CREATE TRIGGER trg_contracts_after_delete
AFTER DELETE ON contracts
FOR EACH ROW
EXECUTE FUNCTION fn_cascade_delete_deliveries();

-- Триггер: логирование вставок в deliveries (AFTER INSERT)
CREATE OR REPLACE FUNCTION fn_log_delivery_insert() RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO deliveries_audit(warehouse_no, receipt_doc_no, contract_no, part_code, qty, received_date, action)
    VALUES (NEW.warehouse_no, NEW.receipt_doc_no, NEW.contract_no, NEW.part_code, NEW.qty, NEW.received_date, 'INSERT');
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_deliveries_after_insert ON deliveries;
CREATE TRIGGER trg_deliveries_after_insert
AFTER INSERT ON deliveries
FOR EACH ROW
EXECUTE FUNCTION fn_log_delivery_insert();

-- хранимая процедура с выходными параметрами
-- возвращает суммарное количество поставок и договорную цену по договору и детали
CREATE OR REPLACE PROCEDURE p_contract_summary(
    IN p_contract_no INT,
    IN p_part_code TEXT,
    OUT total_delivered DECIMAL(10,2),
    OUT contract_price DECIMAL(10,2)
)
LANGUAGE plpgsql
AS $$
BEGIN
    -- Суммарное количество поставленных деталей
    SELECT SUM(d.qty) INTO total_delivered
    FROM deliveries d
    WHERE d.contract_no = p_contract_no AND d.part_code = p_part_code;

    -- Договорная цена
    SELECT c.contract_price INTO contract_price
    FROM contracts c
    WHERE c.contract_no = p_contract_no AND c.part_code = p_part_code;

    -- Если договора нет, вернуть 0 и NULL
    IF NOT FOUND THEN
        total_delivered := 0;
        contract_price := NULL;
    END IF;
END;
$$;

-- Скалярная функция: количество дней между датами
CREATE OR REPLACE FUNCTION fn_warehouse_count(fn_manager_surname text) 
RETURNS INT 
LANGUAGE sql 
AS $$
    SELECT count(*)
	FROM warehouses w
	WHERE w.manager_surname = fn_manager_surname
$$;

-- Табличная функция: список поставок в интервале дат
CREATE OR REPLACE FUNCTION fn_deliveries_in_range(p_start DATE, p_end DATE)
RETURNS TABLE(
    warehouse_no INT, 
    receipt_doc_no INT, 
    contract_no INT, 
    part_code TEXT, 
    qty DECIMAL(10,2), 
    received_date DATE
)
LANGUAGE sql
AS $$
    SELECT warehouse_no, receipt_doc_no, contract_no, part_code, qty, received_date
    FROM deliveries
    WHERE received_date BETWEEN p_start AND p_end
    ORDER BY received_date;
$$;

CREATE VIEW full_deliveries_view AS
	SELECT
		d.warehouse_no,
		w.manager_surname,
		
		d.receipt_doc_no,
		d.received_date,
		d.qty,
		d.unit AS delivery_unit,

		d.contract_no,
		d.part_code,

		c.unit AS contract_unit,
		c.start_date,
		c.end_date,
		c.plan_qty,
		c.contract_price

	FROM deliveries d
	LEFT JOIN warehouses w 
		ON d.warehouse_no = w.warehouse_no
	LEFT JOIN contracts c
		ON d.contract_no = c.contract_no
	AND d.part_code = c.part_code;
    
-- filling with example data
INSERT INTO warehouses (manager_surname) VALUES
('Иванов'),
('Петров'),
('Сидоров'),
('Смирнов'),
('Кузнецов');

INSERT INTO contracts VALUES
(101, 'A100', 'pcs', '2024-01-01', '2024-06-01', 1000, 12.50),
(101, 'B200', 'kg',  '2024-02-01', '2024-05-01', 500, 8.00),
(102, 'A100', 'pcs', '2024-03-01', '2024-10-01', 1500, 11.90),
(103, 'C300', 'set', '2024-01-15', '2024-12-31', 200, 40.00),
(104, 'D400', 'm',   '2024-04-10', '2024-07-20', 3000, 3.50),
(105, 'B200', 'kg',  '2024-02-15', '2024-08-15', 700, 7.80);

INSERT INTO deliveries VALUES
(1, 1, 101, 'A100', 'pcs', 120, '2024-01-10'),
(1, 2, 101, 'A100', 'pcs', 230, '2024-02-12'),
(1, 3, 101, 'B200', 'kg',  50,  '2024-03-02'),

(2, 1, 102, 'A100', 'pcs', 300, '2024-03-15'),
(2, 2, 102, 'A100', 'pcs', 410, '2024-04-01'),

(3, 1, 103, 'C300', 'set', 10,  '2024-02-20'),
(3, 2, 103, 'C300', 'set', 15,  '2024-03-18'),

(4, 1, 104, 'D400', 'm',   500, '2024-05-10'),
(4, 2, 104, 'D400', 'm',   700, '2024-06-14'),

(5, 1, 105, 'B200', 'kg', 120, '2024-02-28'),
(5, 2, 105, 'B200', 'kg', 160, '2024-03-20'),
(5, 3, 105, 'B200', 'kg', 200, '2024-04-25');

COMMIT;