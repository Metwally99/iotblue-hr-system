CREATE TABLE leave_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    leave_year INTEGER NOT NULL,
    total_days INTEGER NOT NULL,
    used_days INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT leave_balances_year_valid CHECK (leave_year BETWEEN 2000 AND 2200),
    CONSTRAINT leave_balances_total_positive CHECK (total_days > 0),
    CONSTRAINT leave_balances_used_valid CHECK (used_days >= 0 AND used_days <= total_days),
    CONSTRAINT leave_balances_employee_year_unique UNIQUE (employee_id, leave_year)
);

CREATE INDEX leave_balances_year_idx ON leave_balances (leave_year);

