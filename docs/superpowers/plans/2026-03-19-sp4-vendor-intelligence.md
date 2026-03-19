# SP4: Vendor Intelligence Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build vendor directory with spend analysis, item coverage, and scoring — derived entirely from existing data with no new migrations.

**Architecture:** New Go service `internal/vendor/` queries ingredient_location_configs + ingredients + recipe_explosion + checks via TenantTx. HTTP handler exposes vendor list and summary. React frontend adds Vendor page with KPI cards and DataTable.

**Tech Stack:** Go 1.22+ (pgx/v5, TenantTx), React 19, TypeScript, Tailwind CSS 4, TanStack React Query, Lucide icons.

**Spec:** `docs/superpowers/specs/2026-03-19-sp4-vendor-intelligence-design.md`

---

Tasks: Backend service → Handler + wiring → Frontend → Smoke test
