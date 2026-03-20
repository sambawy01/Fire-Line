-- Equipment and maintenance management

CREATE TABLE equipment (
    equipment_id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    name           TEXT NOT NULL,
    category       TEXT NOT NULL CHECK (category IN ('cooking', 'refrigeration', 'hvac', 'plumbing', 'electrical', 'safety', 'other')),
    make           TEXT,
    model          TEXT,
    serial_number  TEXT,
    install_date   DATE,
    warranty_expiry DATE,
    status         TEXT NOT NULL DEFAULT 'operational' CHECK (status IN ('operational', 'needs_maintenance', 'under_repair', 'out_of_service', 'retired')),
    last_maintenance DATE,
    next_maintenance DATE,
    maintenance_interval_days INT DEFAULT 90,
    health_score   INT NOT NULL DEFAULT 100 CHECK (health_score BETWEEN 0 AND 100),
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_equipment_location ON equipment(location_id);
CREATE INDEX idx_equipment_status ON equipment(status);
CREATE INDEX idx_equipment_category ON equipment(category);

CREATE TABLE maintenance_tickets (
    ticket_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    equipment_id   UUID NOT NULL REFERENCES equipment(equipment_id),
    ticket_number  TEXT NOT NULL,
    type           TEXT NOT NULL CHECK (type IN ('preventive', 'corrective', 'emergency', 'inspection')),
    priority       TEXT NOT NULL DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    status         TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_progress', 'on_hold', 'completed', 'cancelled')),
    title          TEXT NOT NULL,
    description    TEXT,
    assigned_to    TEXT,
    estimated_cost INT DEFAULT 0,
    actual_cost    INT DEFAULT 0,
    scheduled_date DATE,
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    resolution     TEXT,
    created_by     UUID REFERENCES users(user_id),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_maint_tickets_location ON maintenance_tickets(location_id);
CREATE INDEX idx_maint_tickets_equipment ON maintenance_tickets(equipment_id);
CREATE INDEX idx_maint_tickets_status ON maintenance_tickets(status);

CREATE TABLE maintenance_logs (
    log_id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    ticket_id      UUID REFERENCES maintenance_tickets(ticket_id),
    equipment_id   UUID NOT NULL REFERENCES equipment(equipment_id),
    action         TEXT NOT NULL,
    notes          TEXT,
    cost           INT DEFAULT 0,
    performed_by   TEXT,
    performed_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_maint_logs_equipment ON maintenance_logs(equipment_id);
CREATE INDEX idx_maint_logs_ticket ON maintenance_logs(ticket_id);

-- RLS
ALTER TABLE equipment ENABLE ROW LEVEL SECURITY;
ALTER TABLE equipment FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON equipment USING (org_id = current_setting('app.current_org_id')::UUID);

ALTER TABLE maintenance_tickets ENABLE ROW LEVEL SECURITY;
ALTER TABLE maintenance_tickets FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON maintenance_tickets USING (org_id = current_setting('app.current_org_id')::UUID);

ALTER TABLE maintenance_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE maintenance_logs FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON maintenance_logs USING (org_id = current_setting('app.current_org_id')::UUID);

-- ============================================================
-- SEED DATA: Equipment + Maintenance Tickets + Logs
-- for all 4 branches of org 3f7ef589-f499-43e3-a1c5-aaacd9d543ec
-- ============================================================

DO $$
DECLARE
    v_org UUID := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_loc UUID;
    v_loc_name TEXT;
    v_eq_grill UUID;
    v_eq_cooler UUID;
    v_eq_fryer UUID;
    v_eq_dish UUID;
    v_eq_hood UUID;
    v_eq_ice UUID;
    v_eq_pos UUID;
    v_eq_fire UUID;
    v_tk1 UUID;
    v_tk2 UUID;
    v_tk3 UUID;
    v_tk4 UUID;
    v_tk5 UUID;
    v_locs UUID[] := ARRAY[
        'a1111111-1111-1111-1111-111111111111'::UUID,
        'b2222222-2222-2222-2222-222222222222'::UUID,
        'c3333333-3333-3333-3333-333333333333'::UUID,
        'd4444444-4444-4444-4444-444444444444'::UUID
    ];
    v_counter INT := 0;
BEGIN
    FOREACH v_loc IN ARRAY v_locs
    LOOP
        v_counter := v_counter + 1;

        -- Equipment
        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'Commercial Grill #1', 'cooking', 'Vulcan', 'VCCB60', 'VUL-' || v_counter || '001', '2024-01-15', '2027-01-15', 'needs_maintenance', '2026-01-10', '2026-04-10', 90, 72, 'Temperature calibration needed')
        RETURNING equipment_id INTO v_eq_grill;

        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'Walk-in Cooler', 'refrigeration', 'True', 'TWT-48SD', 'TRU-' || v_counter || '002', '2023-06-01', '2026-06-01', 'operational', '2026-02-15', '2026-05-15', 90, 65, 'Compressor running slightly warm')
        RETURNING equipment_id INTO v_eq_cooler;

        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'Deep Fryer', 'cooking', 'Pitco', 'SSH55', 'PIT-' || v_counter || '003', '2024-03-20', '2027-03-20', 'operational', '2026-02-01', '2026-05-01', 90, 88, NULL)
        RETURNING equipment_id INTO v_eq_fryer;

        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'Dishwasher', 'plumbing', 'Hobart', 'AM15VL', 'HOB-' || v_counter || '004', '2023-09-10', '2025-09-10', 'under_repair', '2026-03-01', '2026-06-01', 90, 45, 'Filter replacement in progress')
        RETURNING equipment_id INTO v_eq_dish;

        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'Hood Ventilation System', 'hvac', 'CaptiveAire', 'ND-2', 'CAP-' || v_counter || '005', '2023-01-01', '2028-01-01', 'operational', '2026-03-01', '2026-06-01', 90, 92, NULL)
        RETURNING equipment_id INTO v_eq_hood;

        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'Ice Machine', 'refrigeration', 'Manitowoc', 'IYT0420A', 'MAN-' || v_counter || '006', '2024-06-15', '2027-06-15', 'operational', '2026-02-20', '2026-05-20', 90, 78, NULL)
        RETURNING equipment_id INTO v_eq_ice;

        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'POS Terminal #1', 'electrical', 'Toast', 'Flex', 'TST-' || v_counter || '007', '2024-09-01', '2027-09-01', 'operational', '2026-03-10', '2026-06-10', 180, 95, NULL)
        RETURNING equipment_id INTO v_eq_pos;

        INSERT INTO equipment (equipment_id, org_id, location_id, name, category, make, model, serial_number, install_date, warranty_expiry, status, last_maintenance, next_maintenance, maintenance_interval_days, health_score, notes)
        VALUES
            (gen_random_uuid(), v_org, v_loc, 'Fire Suppression System', 'safety', 'Ansul', 'R-102', 'ANS-' || v_counter || '008', '2023-01-01', '2029-01-01', 'operational', '2026-01-15', '2026-07-15', 180, 100, 'Annual inspection passed')
        RETURNING equipment_id INTO v_eq_fire;

        -- Maintenance Tickets
        INSERT INTO maintenance_tickets (ticket_id, org_id, location_id, equipment_id, ticket_number, type, priority, status, title, description, assigned_to, estimated_cost, actual_cost, scheduled_date, started_at, completed_at, resolution)
        VALUES
            (gen_random_uuid(), v_org, v_loc, v_eq_grill, 'MT-' || v_counter || '001', 'preventive', 'high', 'in_progress', 'Grill #1 temperature calibration', 'Grill temperature readings are off by 15 degrees. Needs sensor recalibration.', 'Maintenance Team', 25000, 0, CURRENT_DATE, now() - interval '2 hours', NULL, NULL)
        RETURNING ticket_id INTO v_tk1;

        INSERT INTO maintenance_tickets (ticket_id, org_id, location_id, equipment_id, ticket_number, type, priority, status, title, description, assigned_to, estimated_cost, actual_cost, scheduled_date, started_at, completed_at, resolution)
        VALUES
            (gen_random_uuid(), v_org, v_loc, v_eq_cooler, 'MT-' || v_counter || '002', 'preventive', 'critical', 'open', 'Walk-in cooler compressor inspection', 'Compressor is running louder than usual. Needs immediate inspection to prevent failure.', NULL, 50000, 0, CURRENT_DATE + interval '1 day', NULL, NULL, NULL)
        RETURNING ticket_id INTO v_tk2;

        INSERT INTO maintenance_tickets (ticket_id, org_id, location_id, equipment_id, ticket_number, type, priority, status, title, description, assigned_to, estimated_cost, actual_cost, scheduled_date, started_at, completed_at, resolution)
        VALUES
            (gen_random_uuid(), v_org, v_loc, v_eq_dish, 'MT-' || v_counter || '003', 'corrective', 'medium', 'completed', 'Dishwasher filter replacement', 'Replaced clogged water filter and cleaned spray arms.', 'Maintenance Team', 15000, 12500, CURRENT_DATE - interval '5 days', now() - interval '5 days', now() - interval '4 days', 'Replaced water filter model HF-25. Cleaned all spray arms and checked water pressure.')
        RETURNING ticket_id INTO v_tk3;

        INSERT INTO maintenance_tickets (ticket_id, org_id, location_id, equipment_id, ticket_number, type, priority, status, title, description, assigned_to, estimated_cost, actual_cost, scheduled_date, started_at, completed_at, resolution)
        VALUES
            (gen_random_uuid(), v_org, v_loc, v_eq_hood, 'MT-' || v_counter || '004', 'preventive', 'low', 'open', 'Monthly hood cleaning', 'Scheduled monthly degreasing and filter cleaning for hood ventilation system.', NULL, 8000, 0, CURRENT_DATE + interval '5 days', NULL, NULL, NULL)
        RETURNING ticket_id INTO v_tk4;

        INSERT INTO maintenance_tickets (ticket_id, org_id, location_id, equipment_id, ticket_number, type, priority, status, title, description, assigned_to, estimated_cost, actual_cost, scheduled_date, started_at, completed_at, resolution)
        VALUES
            (gen_random_uuid(), v_org, v_loc, v_eq_fryer, 'MT-' || v_counter || '005', 'emergency', 'critical', 'completed', 'Emergency fryer oil leak repair', 'Oil leak detected at the drain valve. Emergency repair completed.', 'External Contractor', 35000, 42000, CURRENT_DATE - interval '10 days', now() - interval '10 days', now() - interval '9 days', 'Replaced faulty drain valve gasket and tightened all fittings. Tested under pressure - no leaks detected.')
        RETURNING ticket_id INTO v_tk5;

        -- Maintenance Logs
        INSERT INTO maintenance_logs (org_id, ticket_id, equipment_id, action, notes, cost, performed_by, performed_at) VALUES
            (v_org, v_tk1, v_eq_grill, 'Inspection started', 'Initial inspection of grill temperature sensors', 0, 'Ahmed M.', now() - interval '2 hours'),
            (v_org, v_tk1, v_eq_grill, 'Parts ordered', 'Ordered replacement thermocouple sensor', 8500, 'Ahmed M.', now() - interval '1 hour'),
            (v_org, v_tk2, v_eq_cooler, 'Ticket created', 'Reported unusual compressor noise during morning shift', 0, 'Manager on Duty', now() - interval '6 hours'),
            (v_org, v_tk3, v_eq_dish, 'Repair started', 'Began filter replacement procedure', 0, 'Hassan K.', now() - interval '5 days'),
            (v_org, v_tk3, v_eq_dish, 'Parts replaced', 'Installed new water filter HF-25', 8500, 'Hassan K.', now() - interval '5 days' + interval '2 hours'),
            (v_org, v_tk3, v_eq_dish, 'Spray arms cleaned', 'Cleaned and descaled all spray arms', 0, 'Hassan K.', now() - interval '5 days' + interval '3 hours'),
            (v_org, v_tk3, v_eq_dish, 'Repair completed', 'All tests passed. Dishwasher running at full capacity.', 4000, 'Hassan K.', now() - interval '4 days'),
            (v_org, v_tk5, v_eq_fryer, 'Emergency call', 'Oil leak reported by kitchen staff. Fryer taken out of service.', 0, 'Manager on Duty', now() - interval '10 days'),
            (v_org, v_tk5, v_eq_fryer, 'Contractor dispatched', 'External contractor called for emergency repair', 15000, 'Operations Manager', now() - interval '10 days' + interval '1 hour'),
            (v_org, v_tk5, v_eq_fryer, 'Valve replaced', 'Drain valve gasket replaced and pressure tested', 27000, 'External Contractor', now() - interval '9 days'),
            (v_org, NULL, v_eq_fire, 'Annual inspection', 'Fire suppression system passed annual inspection. All nozzles clear, tank pressure normal.', 5000, 'Fire Safety Inc.', now() - interval '60 days'),
            (v_org, NULL, v_eq_pos, 'Software update', 'Updated POS terminal firmware to v4.2.1', 0, 'IT Support', now() - interval '11 days');
    END LOOP;
END $$;
