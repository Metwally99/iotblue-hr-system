CREATE EXTENSION IF NOT EXISTS btree_gist;

ALTER TABLE leave_requests
ADD CONSTRAINT leave_requests_no_active_overlap
EXCLUDE USING gist (
    employee_id WITH =,
    daterange(start_date, end_date, '[]') WITH &&
)
WHERE (status IN ('pending', 'approved'));

