CREATE TABLE attendance_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    attendance_date DATE NOT NULL,
    check_in TIMESTAMPTZ NOT NULL,
    check_out TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT attendance_checkout_valid CHECK (check_out IS NULL OR check_out >= check_in),
    CONSTRAINT attendance_employee_date_unique UNIQUE (employee_id, attendance_date)
);

CREATE INDEX attendance_date_idx ON attendance_records (attendance_date DESC);
CREATE INDEX attendance_employee_date_idx ON attendance_records (employee_id, attendance_date DESC);

