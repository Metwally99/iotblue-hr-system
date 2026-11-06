DROP INDEX IF EXISTS leave_requests_type_idx;
DROP INDEX IF EXISTS leave_balances_type_idx;

ALTER TABLE leave_requests
    DROP CONSTRAINT leave_requests_type_valid,
    ADD COLUMN reason TEXT NOT NULL DEFAULT 'Legacy leave request',
    ADD CONSTRAINT leave_requests_reason_not_blank CHECK (btrim(reason) <> ''),
    DROP COLUMN leave_type;

ALTER TABLE leave_balances
    DROP CONSTRAINT leave_balances_type_valid,
    DROP CONSTRAINT leave_balances_employee_year_type_unique,
    ADD CONSTRAINT leave_balances_employee_year_unique UNIQUE (employee_id, leave_year),
    DROP COLUMN leave_type;

ALTER TABLE employees
    DROP CONSTRAINT employees_role_valid,
    ADD CONSTRAINT employees_role_valid CHECK (role IN ('admin', 'employee'));
