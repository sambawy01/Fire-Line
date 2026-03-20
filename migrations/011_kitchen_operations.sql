-- Kitchen stations, resource profiles, and KDS tickets

CREATE TABLE kitchen_stations (
    station_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    name           TEXT NOT NULL,
    station_type   TEXT NOT NULL,
    max_concurrent INT NOT NULL DEFAULT 4,
    status         TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE menu_item_resource_profiles (
    profile_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    menu_item_id   UUID NOT NULL REFERENCES menu_items(menu_item_id),
    station_type   TEXT NOT NULL,
    task_sequence  INT NOT NULL DEFAULT 1,
    duration_secs  INT NOT NULL DEFAULT 300,
    elu_required   NUMERIC(4,2) NOT NULL DEFAULT 1.0,
    batch_size     INT NOT NULL DEFAULT 1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(org_id, menu_item_id, station_type, task_sequence)
);

CREATE TABLE kds_tickets (
    ticket_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    location_id    UUID NOT NULL REFERENCES locations(location_id),
    check_id       UUID REFERENCES checks(check_id),
    order_number   TEXT,
    channel        TEXT,
    status         TEXT NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'in_progress', 'ready', 'delivered', 'cancelled')),
    priority       INT NOT NULL DEFAULT 0,
    estimated_ready_at TIMESTAMPTZ,
    actual_ready_at    TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE kds_ticket_items (
    ticket_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    ticket_id      UUID NOT NULL REFERENCES kds_tickets(ticket_id),
    menu_item_id   UUID NOT NULL REFERENCES menu_items(menu_item_id),
    item_name      TEXT NOT NULL,
    quantity       INT NOT NULL DEFAULT 1,
    station_type   TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'fired', 'cooking', 'ready', 'cancelled')),
    fire_at        TIMESTAMPTZ,
    started_at     TIMESTAMPTZ,
    completed_at   TIMESTAMPTZ,
    duration_secs  INT,
    notes          TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_stations_location ON kitchen_stations(org_id, location_id);
CREATE INDEX idx_resource_profiles_item ON menu_item_resource_profiles(org_id, menu_item_id);
CREATE INDEX idx_kds_tickets_location ON kds_tickets(org_id, location_id, status);
CREATE INDEX idx_kds_tickets_check ON kds_tickets(check_id);
CREATE INDEX idx_kds_items_ticket ON kds_ticket_items(ticket_id);
CREATE INDEX idx_kds_items_station ON kds_ticket_items(org_id, station_type, status);

-- RLS
ALTER TABLE kitchen_stations ENABLE ROW LEVEL SECURITY;
ALTER TABLE kitchen_stations FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON kitchen_stations USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON kitchen_stations TO fireline_app;

ALTER TABLE menu_item_resource_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE menu_item_resource_profiles FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON menu_item_resource_profiles USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON menu_item_resource_profiles TO fireline_app;

ALTER TABLE kds_tickets ENABLE ROW LEVEL SECURITY;
ALTER TABLE kds_tickets FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON kds_tickets USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON kds_tickets TO fireline_app;

ALTER TABLE kds_ticket_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE kds_ticket_items FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON kds_ticket_items USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON kds_ticket_items TO fireline_app;
