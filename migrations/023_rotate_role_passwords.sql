-- Migration 023: Rotate hardcoded role passwords from migration 001.
--
-- Role passwords must be set via environment-specific automation, not in migrations.
-- This migration forces a password change to prevent the hardcoded defaults from being usable.

DO $$
BEGIN
    -- Rotate fireline_app password — must be reset to the real value by deployment automation
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'fireline_app') THEN
        EXECUTE format('ALTER ROLE fireline_app PASSWORD %L', gen_random_uuid()::text);
        RAISE NOTICE 'fireline_app password rotated — set real password via ALTER ROLE before deploying app';
    END IF;
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'fireline_audit') THEN
        EXECUTE format('ALTER ROLE fireline_audit PASSWORD %L', gen_random_uuid()::text);
        RAISE NOTICE 'fireline_audit password rotated — set real password via ALTER ROLE before deploying app';
    END IF;
END $$;
