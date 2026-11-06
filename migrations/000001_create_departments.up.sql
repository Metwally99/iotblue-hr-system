CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT departments_name_not_blank CHECK (btrim(name) <> ''),
    CONSTRAINT departments_name_unique UNIQUE (name)
);

