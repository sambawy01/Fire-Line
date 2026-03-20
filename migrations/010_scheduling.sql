-- Scheduling, demand forecast, and shift swaps

CREATE TABLE schedules (
    schedule_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    week_start     DATE NOT NULL,
    status         TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    created_by     UUID REFERENCES users(user_id),
    published_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, week_start)
);

CREATE TABLE scheduled_shifts (
    scheduled_shift_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    schedule_id    UUID NOT NULL REFERENCES schedules(schedule_id),
    employee_id    UUID NOT NULL REFERENCES employees(employee_id),
    shift_date     DATE NOT NULL,
    start_time     TIME NOT NULL,
    end_time       TIME NOT NULL,
    station        TEXT,
    status         TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'confirmed', 'swap_requested', 'swapped', 'cancelled')),
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE shift_swap_requests (
    swap_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    requester_shift_id UUID NOT NULL REFERENCES scheduled_shifts(scheduled_shift_id),
    target_employee_id UUID REFERENCES employees(employee_id),
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'denied', 'cancelled')),
    reason         TEXT,
    reviewed_by    UUID REFERENCES users(user_id),
    reviewed_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE labor_demand_forecast (
    forecast_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    forecast_date  DATE NOT NULL,
    time_block     TIME NOT NULL,
    forecasted_covers  INT NOT NULL DEFAULT 0,
    required_elu       NUMERIC(8,2) NOT NULL DEFAULT 0,
    required_headcount INT NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, location_id, forecast_date, time_block)
);

-- Indexes
CREATE INDEX idx_schedules_location ON schedules(org_id, location_id, week_start DESC);
CREATE INDEX idx_scheduled_shifts_schedule ON scheduled_shifts(schedule_id);
CREATE INDEX idx_scheduled_shifts_employee ON scheduled_shifts(org_id, employee_id, shift_date);
CREATE INDEX idx_swap_requests_status ON shift_swap_requests(org_id, status);
CREATE INDEX idx_forecast_location ON labor_demand_forecast(org_id, location_id, forecast_date);

-- RLS for all 4 tables
ALTER TABLE schedules ENABLE ROW LEVEL SECURITY;
ALTER TABLE schedules FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON schedules USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON schedules TO fireline_app;

ALTER TABLE scheduled_shifts ENABLE ROW LEVEL SECURITY;
ALTER TABLE scheduled_shifts FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON scheduled_shifts USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON scheduled_shifts TO fireline_app;

ALTER TABLE shift_swap_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE shift_swap_requests FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON shift_swap_requests USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON shift_swap_requests TO fireline_app;

ALTER TABLE labor_demand_forecast ENABLE ROW LEVEL SECURITY;
ALTER TABLE labor_demand_forecast FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON labor_demand_forecast USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON labor_demand_forecast TO fireline_app;
