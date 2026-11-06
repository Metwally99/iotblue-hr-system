ALTER TABLE employees
    DROP CONSTRAINT employees_role_valid,
    ADD CONSTRAINT employees_role_valid CHECK (role IN ('admin', 'hr', 'manager', 'employee'));

ALTER TABLE leave_balances
    ADD COLUMN leave_type TEXT NOT NULL DEFAULT 'annual',
    DROP CONSTRAINT leave_balances_employee_year_unique,
    ADD CONSTRAINT leave_balances_type_valid CHECK (leave_type IN ('annual', 'sick', 'holiday')),
    ADD CONSTRAINT leave_balances_employee_year_type_unique UNIQUE (employee_id, leave_year, leave_type);

ALTER TABLE leave_requests
    ADD COLUMN leave_type TEXT NOT NULL DEFAULT 'annual',
    DROP CONSTRAINT leave_requests_reason_not_blank,
    DROP COLUMN reason,
    ADD CONSTRAINT leave_requests_type_valid CHECK (leave_type IN ('annual', 'sick', 'holiday'));

CREATE INDEX leave_balances_type_idx ON leave_balances (leave_type);
CREATE INDEX leave_requests_type_idx ON leave_requests (leave_type);
