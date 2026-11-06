CREATE TABLE employees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    department_id UUID NOT NULL REFERENCES departments(id) ON DELETE RESTRICT,
    manager_id UUID REFERENCES employees(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'employee',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT employees_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT employees_email_normalized CHECK (email = lower(btrim(email))),
    CONSTRAINT employees_email_unique UNIQUE (email),
    CONSTRAINT employees_role_valid CHECK (role IN ('admin', 'employee')),
    CONSTRAINT employees_manager_not_self CHECK (manager_id IS NULL OR manager_id <> id)
);

CREATE INDEX employees_department_id_idx ON employees (department_id);
CREATE INDEX employees_manager_id_idx ON employees (manager_id) WHERE manager_id IS NOT NULL;
CREATE INDEX employees_active_name_idx ON employees (is_active, name);

