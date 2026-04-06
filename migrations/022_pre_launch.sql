-- 022_pre_launch.sql
-- Pre-launch fixes: ticket sequence, org-scoped email, GDPR guest tables prep

-- Fix 1: Replace COUNT(*)-based ticket numbering with a proper sequence.
-- Sequences are concurrency-safe and gap-free under normal operation.
-- START WITH 1000 avoids collisions with any existing tickets created via COUNT(*).
CREATE SEQUENCE IF NOT EXISTS maintenance_ticket_seq START WITH 1000;

-- Fix 2: Change users.email from globally unique to org-scoped unique.
-- Multi-tenant platforms need the same email to exist across different orgs
-- (e.g., a consultant who manages two restaurants).
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_org_email ON users(org_id, email);

-- Fix 3: Ensure guest_profiles and guest_visits have the columns referenced
-- by the GDPR deletion/export handler. These tables were created in 012_guest_profiles.sql;
-- no schema changes needed here, but we add an index for the GDPR lookup pattern.
CREATE INDEX IF NOT EXISTS idx_guest_visits_guest_id ON guest_visits(guest_id);
