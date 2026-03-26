-- Staff intelligence: task templates, tasks, announcements, anomalies

-- ============================================================
-- TASK TEMPLATES
-- ============================================================

CREATE TABLE task_templates (
    template_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID REFERENCES locations(location_id),
    name           TEXT NOT NULL,
    description    TEXT,
    trigger_type   TEXT NOT NULL DEFAULT 'manual' CHECK (trigger_type IN ('manual', 'shift_start', 'shift_end', 'scheduled')),
    target_role    TEXT,
    target_station TEXT,
    items          JSONB NOT NULL DEFAULT '[]',
    active         BOOLEAN NOT NULL DEFAULT true,
    created_by     UUID REFERENCES users(user_id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_task_templates_org ON task_templates(org_id);
CREATE INDEX idx_task_templates_location ON task_templates(location_id);
CREATE INDEX idx_task_templates_trigger ON task_templates(trigger_type);

ALTER TABLE task_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE task_templates FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON task_templates USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE, DELETE ON task_templates TO fireline_app;

-- ============================================================
-- TASKS
-- ============================================================

CREATE TABLE tasks (
    task_id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id            UUID NOT NULL REFERENCES organizations(org_id),
    location_id       UUID NOT NULL REFERENCES locations(location_id),
    template_id       UUID REFERENCES task_templates(template_id),
    title             TEXT NOT NULL,
    description       TEXT,
    type              TEXT NOT NULL DEFAULT 'checklist_item' CHECK (type IN ('checklist_item', 'ad_hoc', 'data_entry')),
    assigned_to       UUID REFERENCES employees(employee_id),
    assigned_by       UUID REFERENCES users(user_id),
    priority          TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'urgent')),
    due_at            TIMESTAMPTZ,
    status            TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'in_progress', 'completed', 'skipped')),
    data_entry_config JSONB,
    data_entry_value  JSONB,
    photo_url         TEXT,
    completed_at      TIMESTAMPTZ,
    completed_by      UUID REFERENCES employees(employee_id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_org ON tasks(org_id);
CREATE INDEX idx_tasks_location ON tasks(location_id);
CREATE INDEX idx_tasks_template ON tasks(template_id);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_type ON tasks(type);

ALTER TABLE tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE tasks FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON tasks USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE, DELETE ON tasks TO fireline_app;

-- ============================================================
-- ANNOUNCEMENTS
-- ============================================================

CREATE TABLE announcements (
    announcement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id          UUID NOT NULL REFERENCES organizations(org_id),
    location_id     UUID NOT NULL REFERENCES locations(location_id),
    title           TEXT NOT NULL,
    body            TEXT,
    priority        TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'urgent')),
    created_by      UUID REFERENCES users(user_id),
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_announcements_org ON announcements(org_id);
CREATE INDEX idx_announcements_location ON announcements(location_id);
CREATE INDEX idx_announcements_priority ON announcements(priority);

ALTER TABLE announcements ENABLE ROW LEVEL SECURITY;
ALTER TABLE announcements FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON announcements USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE, DELETE ON announcements TO fireline_app;

-- ============================================================
-- ANOMALIES
-- ============================================================

CREATE TABLE anomalies (
    anomaly_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id           UUID NOT NULL REFERENCES organizations(org_id),
    location_id      UUID NOT NULL REFERENCES locations(location_id),
    employee_id      UUID REFERENCES employees(employee_id),
    type             TEXT NOT NULL CHECK (type IN ('void_pattern', 'cash_variance', 'clock_irregularity', 'shrinkage', 'transaction_pattern')),
    severity         TEXT NOT NULL DEFAULT 'warning' CHECK (severity IN ('info', 'warning', 'critical')),
    title            TEXT NOT NULL,
    description      TEXT,
    evidence         JSONB,
    status           TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'investigating', 'confirmed', 'false_positive', 'resolved')),
    assigned_to      UUID REFERENCES users(user_id),
    resolved_by      UUID REFERENCES users(user_id),
    resolved_at      TIMESTAMPTZ,
    resolution_notes TEXT,
    detected_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_anomalies_org ON anomalies(org_id);
CREATE INDEX idx_anomalies_location ON anomalies(location_id);
CREATE INDEX idx_anomalies_employee ON anomalies(employee_id);
CREATE INDEX idx_anomalies_type ON anomalies(type);
CREATE INDEX idx_anomalies_status ON anomalies(status);
CREATE INDEX idx_anomalies_severity ON anomalies(severity);
CREATE INDEX idx_anomalies_assigned_to ON anomalies(assigned_to);

ALTER TABLE anomalies ENABLE ROW LEVEL SECURITY;
ALTER TABLE anomalies FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON anomalies USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE, DELETE ON anomalies TO fireline_app;

-- ============================================================
-- SEED DATA: Task Templates, Announcements, Anomalies
-- for org 3f7ef589-f499-43e3-a1c5-aaacd9d543ec
-- ============================================================

DO $$
DECLARE
    v_org  UUID := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_loc_elgouna  UUID := 'a1111111-1111-1111-1111-111111111111';
    v_loc_newcairo UUID := 'b2222222-2222-2222-2222-222222222222';
    v_loc_zayed    UUID := 'c3333333-3333-3333-3333-333333333333';
    v_loc_maadi    UUID := 'd4444444-4444-4444-4444-444444444444';
BEGIN

    -- --------------------------------------------------------
    -- TASK TEMPLATES (3)
    -- --------------------------------------------------------

    -- 1. Opening Duties (shift_start, 5 items incl. 2 data_entry)
    INSERT INTO task_templates (org_id, location_id, name, description, trigger_type, target_role, items, active)
    VALUES (
        v_org, NULL,
        'Opening Duties',
        'Standard opening checklist for all staff at shift start.',
        'shift_start',
        'staff',
        '[
            {"label": "Wipe down and sanitize all prep surfaces", "type": "checklist_item"},
            {"label": "Record walk-in cooler temperature", "type": "data_entry", "data_entry_config": {"field": "walk_in_temp", "unit": "°C", "min": 0, "max": 10}},
            {"label": "Verify cash drawer count", "type": "data_entry", "data_entry_config": {"field": "cash_drawer_count", "unit": "EGP", "min": 0, "max": 50000}},
            {"label": "Restock napkins, utensils, and condiments at all stations", "type": "checklist_item"},
            {"label": "Confirm POS terminal is online and printer loaded", "type": "checklist_item"}
        ]'::JSONB,
        true
    );

    -- 2. Closing Duties (shift_end, 4 items)
    INSERT INTO task_templates (org_id, location_id, name, description, trigger_type, target_role, items, active)
    VALUES (
        v_org, NULL,
        'Closing Duties',
        'End-of-shift closing checklist for all staff.',
        'shift_end',
        'staff',
        '[
            {"label": "Deep clean all cooking stations and grill surfaces", "type": "checklist_item"},
            {"label": "Store and label all prepped food items with date and time", "type": "checklist_item"},
            {"label": "Empty all trash bins and replace liners", "type": "checklist_item"},
            {"label": "Lock walk-in cooler and secure back entrance", "type": "checklist_item"}
        ]'::JSONB,
        true
    );

    -- 3. Mid-Shift Prep Check (manual, 3 items)
    INSERT INTO task_templates (org_id, location_id, name, description, trigger_type, target_role, items, active)
    VALUES (
        v_org, NULL,
        'Mid-Shift Prep Check',
        'Mid-shift inventory and quality verification for kitchen staff.',
        'manual',
        'staff',
        '[
            {"label": "Check prep levels for high-turnover items (rice, sauces, proteins)", "type": "checklist_item"},
            {"label": "Verify expo station mise en place is fully stocked", "type": "checklist_item"},
            {"label": "Wipe down and re-sanitize all cutting boards", "type": "checklist_item"}
        ]'::JSONB,
        true
    );

    -- --------------------------------------------------------
    -- ANNOUNCEMENTS (5 across 4 locations)
    -- --------------------------------------------------------

    INSERT INTO announcements (org_id, location_id, title, body, priority, expires_at) VALUES
    (
        v_org, v_loc_elgouna,
        'New Summer Menu Launch',
        'The new summer menu goes live this Saturday. All staff must attend the tasting session on Thursday at 3pm. Updated allergen sheets are in the kitchen office.',
        'urgent',
        now() + interval '7 days'
    ),
    (
        v_org, v_loc_newcairo,
        'Health Inspection Scheduled',
        'Government health inspection is scheduled for next Tuesday. Please ensure all stations are spotless and temperature logs are up to date. Manager walk-through on Monday evening.',
        'urgent',
        now() + interval '5 days'
    ),
    (
        v_org, v_loc_zayed,
        'Employee of the Month: March',
        'Congratulations to Fatma A. for outstanding guest feedback scores and zero missed shifts this month. Gift card will be distributed at the next team meeting.',
        'normal',
        now() + interval '30 days'
    ),
    (
        v_org, v_loc_maadi,
        'Updated Uniform Policy',
        'Starting next week, all front-of-house staff must wear the new branded aprons. Pick them up from the manager office before your next shift.',
        'low',
        now() + interval '14 days'
    ),
    (
        v_org, v_loc_elgouna,
        'Ramadan Operating Hours',
        'During Ramadan our operating hours will shift to 2pm-2am. Updated schedules have been posted in the app. Iftar staff meals will be provided daily at 6:30pm.',
        'normal',
        now() + interval '30 days'
    );

    -- --------------------------------------------------------
    -- ANOMALIES (3)
    -- --------------------------------------------------------

    -- 1. void_pattern, warning, El Gouna
    INSERT INTO anomalies (org_id, location_id, type, severity, title, description, evidence, status, detected_at) VALUES
    (
        v_org, v_loc_elgouna,
        'void_pattern', 'warning',
        'Unusual void frequency on POS #2',
        'POS terminal #2 at El Gouna has recorded 14 voids in the last 3 shifts, compared to an average of 3. Most voids occur between 9pm-11pm and involve high-value items.',
        '{"void_count": 14, "avg_void_count": 3, "period": "last_3_shifts", "terminal": "POS #2", "peak_hours": "21:00-23:00", "affected_items": ["Wagyu Ribeye", "Lobster Linguine", "Truffle Risotto"]}'::JSONB,
        'open',
        now() - interval '6 hours'
    );

    -- 2. cash_variance, critical, New Cairo
    INSERT INTO anomalies (org_id, location_id, type, severity, title, description, evidence, status, detected_at) VALUES
    (
        v_org, v_loc_newcairo,
        'cash_variance', 'critical',
        'Cash drawer short EGP 2,340 at close',
        'New Cairo location reported a cash variance of -EGP 2,340 at last night close. This is the third consecutive short in 5 days, totalling EGP 4,870. Pattern suggests systematic issue rather than counting error.',
        '{"variance_egp": -2340, "shift_date": "2026-03-25", "cumulative_5d": -4870, "consecutive_shorts": 3, "drawer_id": "DRAWER-NC-01", "cashier_on_duty": "Employee #NC-012"}'::JSONB,
        'investigating',
        now() - interval '14 hours'
    );

    -- 3. clock_irregularity, warning, Zayed
    INSERT INTO anomalies (org_id, location_id, type, severity, title, description, evidence, status, detected_at) VALUES
    (
        v_org, v_loc_zayed,
        'clock_irregularity', 'warning',
        'Buddy punching suspected at Sheikh Zayed',
        'Two employees at Sheikh Zayed clocked in within 4 seconds of each other from the same device on 3 separate occasions this week. Clock-in times do not match scheduled shifts.',
        '{"incidents": 3, "period": "2026-03-20 to 2026-03-25", "device_id": "KIOSK-SZ-01", "time_gap_seconds": 4, "employee_pair": ["Employee #SZ-008", "Employee #SZ-015"], "scheduled_vs_actual_delta_min": 22}'::JSONB,
        'open',
        now() - interval '2 days'
    );

END $$;
