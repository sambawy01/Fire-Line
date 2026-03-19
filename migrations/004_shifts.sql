-- Shift tracking for labor cost intelligence

CREATE TABLE shifts (
    shift_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    employee_id UUID NOT NULL REFERENCES employees(employee_id),
    role TEXT NOT NULL DEFAULT 'staff',
    clock_in TIMESTAMPTZ NOT NULL,
    clock_out TIMESTAMPTZ,
    hourly_rate INT NOT NULL DEFAULT 0 CHECK (hourly_rate >= 0),
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'no_show')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (clock_out IS NULL OR clock_out > clock_in)
);

CREATE INDEX idx_shifts_org ON shifts(org_id);
CREATE INDEX idx_shifts_location ON shifts(org_id, location_id);
CREATE INDEX idx_shifts_employee ON shifts(employee_id);
CREATE INDEX idx_shifts_clock ON shifts(org_id, location_id, clock_in DESC);

ALTER TABLE shifts ENABLE ROW LEVEL SECURITY;
ALTER TABLE shifts FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON shifts
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON shifts TO fireline_app;
