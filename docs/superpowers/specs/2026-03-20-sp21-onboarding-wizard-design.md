# SP21: Onboarding Wizard & Day-One Experience

**Date:** 2026-03-20
**Status:** Approved
**Scope:** Multi-step onboarding wizard, POS connection flow, First Insights dashboard, concept type inference, module activation, Day 1 checklist
**Maps to:** Build Plan Sprint 53

---

## 1. Database — Migration 017

```sql
CREATE TABLE onboarding_sessions (
    session_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    user_id        UUID NOT NULL REFERENCES users(user_id),
    current_step   TEXT NOT NULL DEFAULT 'profile' CHECK (current_step IN ('profile', 'pos_connect', 'importing', 'first_insights', 'concept_type', 'priorities', 'modules', 'checklist', 'complete')),
    profile_data   JSONB NOT NULL DEFAULT '{}',
    concept_type   TEXT,
    priorities     TEXT[] NOT NULL DEFAULT '{}',
    active_modules TEXT[] NOT NULL DEFAULT '{}',
    insights_data  JSONB NOT NULL DEFAULT '{}',
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE onboarding_checklist_items (
    item_id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id         UUID NOT NULL REFERENCES organizations(org_id),
    title          TEXT NOT NULL,
    description    TEXT,
    category       TEXT NOT NULL,
    priority       INT NOT NULL DEFAULT 0,
    completed      BOOLEAN NOT NULL DEFAULT false,
    completed_at   TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

RLS on both. Indexes.

---

## 2. Onboarding Service

New package: `internal/onboarding/`

### Step Flow

1. **Profile**: org name, location details (address, timezone, cuisine type, seating capacity)
2. **POS Connect**: select POS type → OAuth or API key entry → test connection
3. **Importing**: real-time progress of data import (menu items, employees, transactions)
4. **First Insights**: auto-generated dashboard from imported data
5. **Concept Type**: inferred from menu/check data, user confirms (fast_casual, full_service, fine_dining, qsr, cafe, bar, ghost_kitchen)
6. **Priorities**: select top 3 from: food_cost_control, labor_optimization, revenue_growth, waste_reduction, customer_retention, operational_efficiency
7. **Modules**: recommended module activation based on priorities
8. **Checklist**: personalized Day 1 tasks

### Methods

- `StartOnboarding(ctx, orgID, userID)` — create session
- `UpdateStep(ctx, orgID, sessionID, step, data)` — advance to next step
- `GetSession(ctx, orgID, sessionID)` — current state
- `InferConceptType(ctx, orgID, locationID)` — analyze menu prices + check patterns
- `GenerateFirstInsights(ctx, orgID, locationID)` — compute revenue, top sellers, peak hours, avg check, void rate
- `RecommendModules(priorities []string) []string` — priority → module mapping
- `GenerateChecklist(ctx, orgID, conceptType string, modules []string)` — create personalized checklist items
- `GetChecklist(ctx, orgID)` — list items
- `CompleteChecklistItem(ctx, orgID, itemID)` — mark done

### Concept Type Inference

- avg_check > 5000 cents → fine_dining
- avg_check > 2500 → full_service
- avg_check > 1200 → fast_casual
- avg_check > 800 → qsr
- menu has "espresso" or "latte" → cafe
- Default: fast_casual

### Module Recommendations

| Priority | Recommended Modules |
|----------|-------------------|
| food_cost_control | Inventory, POs, Menu Scoring |
| labor_optimization | Scheduling, ELU, Points |
| revenue_growth | Menu Simulation, Marketing, Loyalty |
| waste_reduction | Counting, Waste Logging, Variance |
| customer_retention | Guest Profiles, Churn, Loyalty |
| operational_efficiency | KDS, Kitchen Capacity, Overload |

### Day 1 Checklist Generation

Based on concept type + modules:
- "Review your imported menu items" (always)
- "Set PAR levels for top 10 ingredients" (if inventory active)
- "Set your weekly revenue budget" (if financial active)
- "Assign ELU ratings for your staff" (if labor active)
- "Create your first schedule" (if scheduling active)
- "Set up kitchen stations" (if KDS active)
- "Invite your team members" (always)

---

## 3. API Endpoints

```
POST   /api/v1/onboarding/start              — Start onboarding session
GET    /api/v1/onboarding/session             — Get current session
PUT    /api/v1/onboarding/step                — Update step (body: {step, data})
GET    /api/v1/onboarding/first-insights      — Generate insights for location
POST   /api/v1/onboarding/infer-concept       — Infer concept type
POST   /api/v1/onboarding/recommend-modules   — Get module recommendations
GET    /api/v1/onboarding/checklist           — Get checklist
POST   /api/v1/onboarding/checklist/{id}/complete — Complete item
```

---

## 4. Web — Onboarding Flow

New page: `web/src/pages/OnboardingPage.tsx` (route `/onboarding`)

After signup, redirect to `/onboarding` if no onboarding session completed.

### Multi-step wizard UI

- Progress bar at top showing current step (1-8)
- Each step is a full-screen section

**Step 1: Profile** — form: restaurant name, address, timezone, cuisine, seating capacity
**Step 2: POS Connect** — POS logo grid (Toast, Square, Clover, CSV), connection form
**Step 3: Importing** — animated progress feed: "Importing menu items... 64 found", "Importing transactions... syncing"
**Step 4: First Insights** — auto-generated cards: daily revenue avg, top sellers, peak hour, avg check, void rate, staff count
**Step 5: Concept Type** — inferred badge with "Is this correct?" confirmation, option to override
**Step 6: Priorities** — selectable priority cards (max 3), each with icon and description
**Step 7: Modules** — recommended modules shown as toggleable cards based on priorities
**Step 8: Checklist** — personalized todo list with checkboxes, "Go to Dashboard" button

---

## 5. Testing
- Concept type inference: known avg_check → correct type
- Module recommendations: known priorities → correct modules
- Checklist generation: known concept + modules → correct items
