CREATE TABLE leave_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    reason TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    reviewed_by UUID REFERENCES employees(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    review_note TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT leave_requests_dates_valid CHECK (start_date <= end_date),
    CONSTRAINT leave_requests_one_year CHECK (extract(year FROM start_date) = extract(year FROM end_date)),
    CONSTRAINT leave_requests_reason_not_blank CHECK (btrim(reason) <> ''),
    CONSTRAINT leave_requests_status_valid CHECK (status IN ('pending', 'approved', 'rejected')),
    CONSTRAINT leave_requests_review_consistent CHECK (
        (status = 'pending' AND reviewed_by IS NULL AND reviewed_at IS NULL)
        OR
        (status IN ('approved', 'rejected') AND reviewed_by IS NOT NULL AND reviewed_at IS NOT NULL)
    )
);

CREATE INDEX leave_requests_employee_created_idx ON leave_requests (employee_id, created_at DESC);
CREATE INDEX leave_requests_status_created_idx ON leave_requests (status, created_at DESC);
CREATE INDEX leave_requests_date_range_idx ON leave_requests (start_date, end_date);

