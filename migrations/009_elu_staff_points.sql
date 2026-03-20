-- ELU ratings and Staff Point System

ALTER TABLE employees
    ADD COLUMN elu_ratings JSONB NOT NULL DEFAULT '{}',
    ADD COLUMN staff_points NUMERIC(10,2) NOT NULL DEFAULT 0,
    ADD COLUMN certifications TEXT[] NOT NULL DEFAULT '{}',
    ADD COLUMN availability JSONB NOT NULL DEFAULT '{}';

CREATE TABLE staff_point_events (
    event_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id        UUID NOT NULL REFERENCES organizations(org_id),
    employee_id   UUID NOT NULL REFERENCES employees(employee_id),
    points        NUMERIC(10,2) NOT NULL,
    reason        TEXT NOT NULL CHECK (reason IN (
        'task_completion', 'speed_bonus', 'accuracy_bonus',
        'attendance', 'peer_nominated', 'late', 'no_show',
        'incomplete_task', 'manager_adjustment'
    )),
    description   TEXT,
    shift_id      UUID REFERENCES shifts(shift_id),
    awarded_by    UUID REFERENCES users(user_id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_point_events_employee ON staff_point_events(org_id, employee_id, created_at DESC);
CREATE INDEX idx_point_events_shift ON staff_point_events(shift_id);

ALTER TABLE staff_point_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE staff_point_events FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON staff_point_events
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON staff_point_events TO fireline_app;
