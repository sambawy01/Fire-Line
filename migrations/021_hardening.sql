-- Hardening: RLS enforcement, missing GRANTs, audit partitions, index optimization

-- 1. Force RLS on portfolio tables (they only have ENABLE, not FORCE)
ALTER TABLE portfolio_nodes FORCE ROW LEVEL SECURITY;
ALTER TABLE location_benchmarks FORCE ROW LEVEL SECURITY;
ALTER TABLE best_practices FORCE ROW LEVEL SECURITY;

-- 2. Missing GRANTs for portfolio tables
GRANT SELECT, INSERT, UPDATE, DELETE ON portfolio_nodes TO fireline_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON location_benchmarks TO fireline_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON best_practices TO fireline_app;

-- 3. Missing GRANTs for maintenance tables (018)
GRANT SELECT, INSERT, UPDATE, DELETE ON equipment TO fireline_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON maintenance_tickets TO fireline_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON maintenance_logs TO fireline_app;

-- 4. Unique constraint on guest_visits to prevent duplicate visit tracking
CREATE UNIQUE INDEX idx_guest_visits_check ON guest_visits(guest_id, check_id);

-- 5. Audit log partitions through end of 2026
CREATE TABLE IF NOT EXISTS audit_log_2026_05 PARTITION OF audit_log FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_06 PARTITION OF audit_log FOR VALUES FROM ('2026-06-01') TO ('2026-07-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_07 PARTITION OF audit_log FOR VALUES FROM ('2026-07-01') TO ('2026-08-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_08 PARTITION OF audit_log FOR VALUES FROM ('2026-08-01') TO ('2026-09-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_09 PARTITION OF audit_log FOR VALUES FROM ('2026-09-01') TO ('2026-10-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_10 PARTITION OF audit_log FOR VALUES FROM ('2026-10-01') TO ('2026-11-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_11 PARTITION OF audit_log FOR VALUES FROM ('2026-11-01') TO ('2026-12-01');
CREATE TABLE IF NOT EXISTS audit_log_2026_12 PARTITION OF audit_log FOR VALUES FROM ('2026-12-01') TO ('2027-01-01');

-- 6. Composite indexes with org_id prefix for RLS performance

-- Maintenance indexes (org_id composites)
CREATE INDEX idx_equipment_org_location ON equipment(org_id, location_id);
CREATE INDEX idx_equipment_org_status ON equipment(org_id, status);
CREATE INDEX idx_maint_tickets_org_location ON maintenance_tickets(org_id, location_id);
CREATE INDEX idx_maint_tickets_org_status ON maintenance_tickets(org_id, status);
CREATE INDEX idx_maint_logs_org_equipment ON maintenance_logs(org_id, equipment_id);

-- Task indexes (org_id composites)
CREATE INDEX idx_tasks_org_location ON tasks(org_id, location_id);
CREATE INDEX idx_tasks_org_status ON tasks(org_id, status);
CREATE INDEX idx_tasks_org_assigned ON tasks(org_id, assigned_to);
CREATE INDEX idx_announcements_org_location ON announcements(org_id, location_id);
CREATE INDEX idx_anomalies_org_location ON anomalies(org_id, location_id);
CREATE INDEX idx_anomalies_org_status ON anomalies(org_id, status);
