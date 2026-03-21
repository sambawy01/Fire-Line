# FireLine by OpsNerve

## The AI Operating System for Multi-Location Restaurant Chains

**Version:** 1.0 | **Date:** March 2026 | **Classification:** Confidential -- For Investors, Partners, and Technical Stakeholders

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Architecture](#2-system-architecture)
3. [AI Intelligence Layer -- The Core Differentiator](#3-ai-intelligence-layer--the-core-differentiator)
4. [Department Operations -- How Each Team Uses FireLine](#4-department-operations--how-each-team-uses-fireline)
5. [Benefits of the AI Intelligence Layer](#5-benefits-of-the-ai-intelligence-layer)
6. [Technical Innovation](#6-technical-innovation)
7. [Roadmap and Future Vision](#7-roadmap-and-future-vision)

---

## 1. Executive Summary

### The Problem

Multi-location restaurant chains operate in an environment defined by thin margins, high labor turnover, perishable inventory, and unpredictable demand. A typical restaurant generates thousands of operational data points each day -- point-of-sale transactions, inventory movements, equipment sensor readings, labor clock events, guest interactions -- yet the vast majority of this data goes unanalyzed. Managers make decisions reactively, discovering problems only after they have already eroded margins. Food spoils before anyone notices the cooler temperature drifting upward. High-value guests stop returning without a single win-back attempt. Inventory variances are written off as shrinkage with no root cause attribution. Vendors are selected on relationships rather than data-driven reliability scores.

The restaurant technology landscape is fragmented: one system for POS, another for scheduling, another for inventory, another for maintenance. None of them talk to each other, and none of them think.

### The Solution

**FireLine** is an AI-first restaurant operations platform that unifies every operational domain -- inventory, finance, labor, menu engineering, vendor management, customer intelligence, kitchen operations, equipment maintenance, and marketing -- into a single system with a shared intelligence layer. Built by **OpsNerve**, FireLine does not merely collect data; it reasons about it, detects anomalies, predicts failures, attributes causes probabilistically, and generates actionable recommendations with confidence scores.

### Key Differentiators

- **Cross-module intelligence.** Inventory alerts feed financial insights. Customer churn data informs menu decisions. Equipment sensor trends trigger maintenance tickets that reference the inventory at risk. No domain operates in isolation.
- **Probabilistic reasoning.** FireLine does not simply flag "variance detected." It attributes causes -- portioning error (62% probability), unrecorded waste (28%), vendor spec change (10%) -- giving managers the context to act, not just the alert to acknowledge.
- **Tiered autonomy.** Routine decisions are handled automatically (reorder stock below PAR, escalate void spikes). Complex decisions surface as recommendations with projected impact and confidence scores, leaving the final call to managers.
- **Multi-tenant by design.** Row-Level Security at the database layer ensures zero-trust data isolation between restaurant organizations, enabling a true SaaS deployment model.

### Target Market

FireLine serves multi-location restaurant chains (5-500+ locations), ghost kitchen operators, franchise groups, and hospitality management companies. The initial demo deployment -- **Nimbu**, a 4-branch international restaurant chain operating in El Gouna, New Cairo, Sheikh Zayed, and North Coast, Egypt -- validates the platform across geographically distributed locations with distinct demand patterns.

### Business Impact

| Metric | Projected Impact |
|---|---|
| Food cost reduction | 2-5% through variance detection, portioning alerts, automated reordering |
| Labor cost optimization | 3-8% through demand-based scheduling, overtime prevention |
| VIP customer retention | 8-12% churn prevention through predictive win-back |
| Equipment uptime | 15-25% reduction in unplanned downtime |
| Waste reduction | 20-30% through expiry tracking, FIFO enforcement, usage variance alerts |

---

## 2. System Architecture

### Design Philosophy: The Modular Monolith

FireLine is built as a **modular monolith** in Go -- a deliberate architectural choice that balances the domain isolation benefits of microservices with the operational simplicity and transactional consistency of a single deployable unit. Each domain (inventory, financial, menu, labor, vendor, customer, operations, marketing, portfolio) is a self-contained Go package with its own models, business logic, and database queries, communicating with other domains exclusively through a typed in-process event bus.

This architecture was chosen for three reasons:

1. **Transactional consistency.** Restaurant operations demand ACID guarantees. When a physical inventory count triggers a variance analysis, which triggers a purchase order, which updates the financial P&L -- that entire chain must succeed or fail atomically. Distributed transactions across microservices add complexity without proportional benefit at this stage.

2. **Deployment simplicity.** A single Go binary deploys to AWS ECS Fargate or Railway with zero service mesh overhead. There are no inter-service network calls to debug, no service discovery to configure, no distributed tracing infrastructure to maintain.

3. **Future-proof boundaries.** The event bus and package boundaries are designed so that any domain can be extracted into an independent service when scale demands it. The event envelope structure is NATS-compatible, making the transition from in-process to distributed messaging a configuration change rather than a rewrite.

### Domain Modules

FireLine comprises 15+ domain packages organized into three tiers:

**Core Infrastructure**

| Package | Responsibility |
|---|---|
| `auth` | JWT authentication, refresh tokens, PIN-based tablet auth, MFA, RBAC middleware |
| `tenant` | Organization context propagation, Row-Level Security enforcement |
| `event` | In-process event bus with NATS-compatible subject naming and wildcard routing |
| `database` | Dual-pool connection management, tenant-scoped transactions, RLS lifecycle hooks |
| `adapter` | POS system integration layer (Toast adapter implemented), raw log ingestion |
| `pipeline` | ETL pipeline: raw POS logs to normalized domain models |
| `config` | Environment-based configuration with validation |
| `observability` | Structured logging with `slog` |

**Domain Intelligence**

| Package | Responsibility |
|---|---|
| `inventory` | Physical counting, waste logging, variance analysis, purchasing, expiry tracking |
| `financial` | P&L computation, budget management, cost centers, transaction anomaly detection |
| `menu` | 5-dimension scoring, 8-class classification, simulation sandbox |
| `labor` | ELU ratings, staff points, demand-based scheduling, shift swaps |
| `vendor` | Vendor Reliability Score, OTIF tracking, price intelligence |
| `customer` | Guest profiles, CLV scoring, RFM segmentation, churn prediction |
| `operations` | Kitchen capacity, KDS ticket routing, overload response, planning horizons |
| `maintenance` | Equipment registry, predictive maintenance, work order lifecycle |
| `marketing` | Campaigns, loyalty program, promotion engine |
| `portfolio` | Multi-location hierarchy, cross-location benchmarking, best practices |

**API and Presentation**

| Package | Responsibility |
|---|---|
| `api` | 120+ RESTful endpoint handlers, request validation, response formatting |
| `reporting` | Report generation, PDF export |
| `alerting` | Priority-based alert routing, notification management |
| `onboarding` | Guided setup wizard for new organizations |

### Event Bus Architecture

The event bus is the nervous system that enables cross-module intelligence without creating tight coupling between domains. Every event is wrapped in a typed `Envelope` containing:

```
EventID       -- unique identifier for tracing
EventType     -- NATS-compatible subject (e.g., "inventory.count.completed")
OrgID         -- tenant scope for RLS enforcement
LocationID    -- optional location scope
Source        -- originating module
SchemaVersion -- payload version for backward compatibility
Payload       -- event-specific data
```

Subjects use dot-separated tokens with wildcard support: `*` matches a single token, `>` matches one or more trailing tokens. This enables subscribers to listen at any granularity -- from a specific event (`inventory.count.completed`) to an entire domain (`inventory.>`).

The bus supports middleware chains for cross-cutting concerns such as logging, metrics, and error handling. Failed events are routed to a dead letter queue for investigation and replay.

**Example cross-module flow:**

1. Inventory count is submitted via tablet (`inventory.count.completed`)
2. Variance engine calculates theoretical vs. actual usage and categorizes causes
3. If a PAR breach is detected, a purchase order is auto-generated (`inventory.par.breached`)
4. Financial module recalculates COGS impact (`financial.cogs.updated`)
5. Alerting module surfaces high-severity variances to the manager dashboard

### Multi-Tenant Data Isolation

FireLine uses PostgreSQL **Row-Level Security (RLS)** for tenant data isolation -- every table row is scoped to an `org_id`, and RLS policies ensure that queries can only access rows belonging to the authenticated organization. This is not application-level filtering; it is enforced at the database engine level.

The implementation follows a fail-closed design:

1. **Before connection acquisition:** The pool validates that a tenant context exists. Connections without tenant context are logged as warnings.
2. **Transaction start:** `TenantTx` sets `SET LOCAL app.current_org_id` within each transaction, scoping all queries to the authenticated organization.
3. **After connection release:** The pool clears the `app.current_org_id` GUC. If clearing fails, the connection is destroyed rather than returned to the pool, preventing context leakage.

This architecture supports true multi-tenant SaaS deployment where multiple restaurant organizations share a single database cluster with cryptographic isolation guarantees.

### Dual-Pool Database Architecture

FireLine maintains two connection pools:

- **Application pool:** Used by all tenant-scoped operations. Every query runs within a `TenantTx` that enforces RLS via `SET LOCAL`.
- **Admin pool:** Used exclusively for schema migrations, cross-tenant analytics, and system administration tasks. Bypasses RLS for legitimate administrative operations.

### Database Schema

The schema is managed through 18 versioned migrations (001 through 018), covering:

- 001-003: Core schema (organizations, locations, users, roles, menu items, recipes, ingredients, checks), authentication tokens, menu and recipe structure
- 004: Shift management
- 005: Customer records
- 006: Inventory counting and waste logging
- 007: Purchase orders and delivery receiving
- 008: Financial budgets and cost centers
- 009: ELU ratings and staff point system
- 010: Demand-based scheduling
- 011: Kitchen operations (stations, resource profiles, KDS tickets)
- 012: Guest profiles, CLV, segmentation
- 013: Menu scoring and simulation
- 014: Vendor scoring and price tracking
- 015: Marketing campaigns and loyalty program
- 016: Multi-location portfolio hierarchy
- 017: Onboarding wizard state
- 018: Equipment maintenance and work orders

### Technology Stack

| Layer | Technology |
|---|---|
| Backend | Go 1.22+, modular monolith, `net/http` standard library router |
| Database | PostgreSQL 16 with Row-Level Security, pgx/v5 driver |
| Frontend | React 18, TypeScript, Tailwind CSS, Vite |
| Tablet App | Expo React Native (iOS/Android) |
| Infrastructure | AWS ECS Fargate (Terraform), Docker |
| Deployment | Vercel (frontend), Railway (backend), Neon/Supabase (database) |
| CI/CD | GitHub Actions |
| Schema Management | Atlas (versioned migrations) |

### API Surface

FireLine exposes 120+ RESTful endpoints organized by domain. All endpoints require JWT authentication (or PIN authentication for tablet operations) and return consistent JSON response envelopes. The API is versioned and designed for third-party integration.

---

## 3. AI Intelligence Layer -- The Core Differentiator

The AI intelligence layer is what separates FireLine from traditional restaurant management software. While conventional systems record what happened, FireLine predicts what will happen, explains why it happened, and recommends what to do next. Each intelligence module operates independently but shares data through the event bus, creating a compound intelligence effect where insights from one domain amplify the value of another.

### 3.1 Predictive Maintenance Intelligence

**Problem:** Unplanned equipment failure is one of the costliest events in restaurant operations. A failed walk-in cooler means thousands of dollars in spoiled inventory. A broken fryer during dinner service means lost revenue and guest dissatisfaction. Most restaurants practice reactive maintenance -- fixing equipment after it breaks.

**FireLine's Approach:**

FireLine's maintenance module tracks equipment health through daily kitchen reports that include sensor readings: cooler temperatures, freezer temperatures, grill surface readings, AC output, and humidity levels. The system builds per-equipment baselines and applies trend detection algorithms to identify degradation patterns before failure occurs.

- **Temperature trend detection.** The system identifies compressor degradation in coolers (gradual temperature increase over days/weeks), AC performance decline (rising ambient temperatures in service areas), and grill sensor drift (inconsistent surface temperatures affecting food quality).
- **Pattern matching against failure curves.** Historical maintenance records create equipment-specific failure probability curves. When current sensor trends match known pre-failure patterns, the system generates preventive maintenance tickets with confidence scores.
- **Risk assessment.** Each maintenance alert includes the inventory at risk (dollar value of perishable goods stored in a degrading cooler), food safety threshold proximity (hours until temperature enters the danger zone), and guest comfort impact for HVAC systems.
- **Auto-generated work orders.** The system creates maintenance tickets with root cause analysis, priority classification, and recommended actions, routing them to the appropriate maintenance staff.

**Equipment health scoring** provides a 0-100 composite score for each piece of equipment, factoring in age, maintenance history, current sensor trends, and manufacturer-recommended service intervals. Equipment below threshold scores is flagged for proactive replacement planning.

### 3.2 Financial Intelligence

**Problem:** Restaurant P&L reporting is typically retrospective -- managers learn about margin erosion weeks after it occurs. Transaction-level fraud (excessive voids, unauthorized discounts, off-hours activity) often goes undetected for months.

**FireLine's Approach:**

The financial intelligence module provides real-time P&L computation with channel-aware margins, distinguishing between dine-in (highest margin), takeout (medium margin with packaging costs), and delivery (lowest margin after platform commissions).

- **Budget vs. actual variance analysis.** Managers set budget targets for revenue, food cost percentage, and labor cost percentage. FireLine computes real-time variance with status classification: `on_track`, `over`, or `under`. Variances exceeding configurable thresholds trigger alerts.
- **Cost center breakdown.** COGS is decomposed by ingredient category (proteins, dairy, produce, dry goods, beverages), enabling managers to identify which categories are driving cost overruns.
- **P&L drill-down.** The financial module supports progressive drill-down from summary P&L to category-level analysis, to individual menu item contribution, to ingredient-level cost, to vendor-level pricing -- a five-level analytical hierarchy.
- **Transaction anomaly detection.** The system applies Z-score analysis against 30-day rolling baselines to detect:
  - **Excessive voids:** Daily void count exceeding 2 standard deviations from the 30-day mean
  - **Discount rate spikes:** Total discount as a percentage of subtotal exceeding baseline norms
  - **Off-hours transactions:** Activity outside normal operating hours
  - **Comp anomalies:** Unusually high complimentary item counts

Each anomaly includes the current value, baseline value, Z-score, and severity classification, giving managers the statistical context to distinguish genuine operational issues from normal variation.

- **Period comparison.** Revenue, costs, and margins can be compared across time periods: current day vs. same day last week, current week vs. prior week, current month vs. same month last year. Trend arrows and percentage changes surface at a glance.

### 3.3 Inventory Intelligence

**Problem:** Inventory variance -- the gap between what a restaurant should have used (theoretical) and what it actually used (actual) -- represents one of the largest controllable cost leaks. Industry averages show 5-10% variance, meaning a restaurant with EGP 1M monthly food cost is losing EGP 50,000-100,000 to waste, theft, portioning errors, and vendor discrepancies. Most systems detect variance but cannot explain it.

**FireLine's Approach:**

FireLine's inventory intelligence module combines recipe explosion (theoretical usage calculation), physical counting, waste logging, and purchase order data to compute per-ingredient variance with **probabilistic cause attribution**.

- **Theoretical vs. actual usage calculation.** For every ingredient, the system computes expected usage by exploding recipes (menu item sold x recipe quantity x number sold) and compares it against actual usage (opening stock + received - closing count).

- **Anti-anchoring physical count workflow.** When staff perform physical counts on the tablet app, FireLine deliberately hides expected quantities. This eliminates anchoring bias -- the psychological tendency to adjust counts toward an expected value rather than counting accurately. Items are grouped by storage category (walk-in, dry storage, freezer) for efficient counting workflow.

- **Variance categorization engine.** This is the core innovation. For each ingredient with significant variance, the system evaluates multiple signals and assigns probability scores to possible causes:

| Cause | Signal | Example |
|---|---|---|
| Portioning error | Shortage + portioning flag from prior audits | "Chicken breast: 62% portioning, 18% unrecorded waste" |
| Unrecorded waste | Shortage partially explained by logged waste quantities | Waste log covers 40% of variance; remainder is unrecorded |
| Theft signal | Persistent shortage with no waste logs, no portioning flags | High-value items with consistent unexplained loss |
| Vendor spec change | Surplus or shortage coinciding with vendor delivery | Received quantity matches PO but yield differs |
| Measurement error | Small variance within expected count precision range | Variance within 2% on bulk-counted items |

Probabilities are computed from the `VarianceSignals` struct, which captures the variance quantity, theoretical usage, logged waste quantity, and portioning flag status. Values do not sum to 1.0 -- multiple causes can contribute simultaneously.

- **PAR breach detection.** When ingredient stock falls below the configured PAR level, the system calculates the projected stockout date based on trailing usage rates and generates an alert with urgency classification.
- **Automated purchase order generation.** Ingredients below reorder points trigger auto-generated purchase orders, pre-populated with the optimal vendor (selected by the vendor intelligence module based on reliability score, price, and lead time) and the quantity needed to restore stock to PAR level.
- **Expiry date tracking.** Batch-level expiry tracking with shelf-life alerts enables FIFO enforcement. Items approaching expiry are surfaced in priority order with the dollar value at risk.
- **Waste logging.** Staff log waste events with reason classification (expired, damaged, overproduction, customer return, prep error) via the tablet app, feeding data back into the variance categorization engine.

### 3.4 Menu Intelligence

**Problem:** Menu engineering in most restaurants relies on the Boston Consulting Group matrix (stars, puzzles, workhorses, dogs) -- a two-dimensional analysis based only on sales volume and contribution margin. This oversimplification ignores operational complexity, customer satisfaction, and strategic value, leading to suboptimal menu decisions.

**FireLine's Approach:**

FireLine scores every menu item on **five dimensions**, producing a nuanced classification that captures the full picture:

| Dimension | Scale | What It Measures |
|---|---|---|
| Contribution Margin | 0-100 | Revenue minus ingredient cost per unit |
| Sales Velocity | 0-100 | Units sold relative to category peers |
| Operational Complexity | 0-100 | Kitchen resource consumption (stations, time, skill level) |
| Customer Satisfaction | 0-100 | Return rate, complaint frequency, social sentiment |
| Strategic Value | 0-100 | Brand identity contribution, seasonal importance, competitive differentiation |

These five scores map to an **8-class classification system** with explicit threshold logic:

| Classification | Criteria | Action Guidance |
|---|---|---|
| **Powerhouse** | High margin + high velocity + manageable complexity | Protect and promote; these are profit engines |
| **Hidden Gem** | High margin + low velocity + high satisfaction | Increase visibility; these items delight guests who find them |
| **Crowd Pleaser** | Low margin + high velocity | Volume drivers; optimize ingredient cost or accept as traffic builders |
| **Workhorse** | High margin + high velocity + high complexity | Profitable but operationally demanding; simplify prep if possible |
| **Complex Star** | High margin + low velocity + high complexity | Evaluate whether margin justifies kitchen burden |
| **Declining Star** | High margin + low velocity + mid complexity + low satisfaction | Formerly strong items losing appeal; refresh or retire |
| **Underperformer** | Low margin + low velocity | Candidates for removal; analyze dependencies first |
| **Strategic Anchor** | High strategic value regardless of economics | Kept for brand or competitive reasons despite weak economics |

**Simulation sandbox.** Before making menu changes, managers can model scenarios:

- **Price change impact:** Adjust a menu item's price and see the projected effect on margin, sales velocity (using price elasticity estimates), and overall revenue.
- **Item removal (86) analysis:** Before removing an item, the system identifies all ingredient dependencies and co-ordered items that may be affected. If an ingredient is used exclusively in the removed item, the system flags the potential vendor contract impact.
- **Ingredient cost change propagation:** When a vendor raises prices on a key ingredient, the simulation shows every menu item affected, the new margins, and which items may need price adjustments.

**Cross-sell affinity analysis** identifies co-ordered item pairs from transaction data, enabling upsell recommendations: "Guests who order the Lamb Kofta also order the Hibiscus Lemonade 43% of the time."

**Ingredient dependency graph** maps which ingredients are used in which menu items, detecting single-point-of-failure risks: "Tahini is used in 14 menu items across 3 categories. A supply disruption affects 22% of your menu."

### 3.5 Customer Intelligence

**Problem:** Most restaurants have no idea who their best customers are, when they are at risk of churning, or how much lifetime revenue they represent. Without POS-integrated loyalty programs, guest identity is invisible.

**FireLine's Approach:**

FireLine resolves guest identity through **payment token hashing** (SHA-256), creating anonymous behavioral profiles without requiring guests to enroll in a loyalty program. Privacy is maintained through a tiered system:

| Privacy Tier | Data Available | Trigger |
|---|---|---|
| Behavioral (anonymous) | Visit frequency, spend patterns, preferred items, daypart preferences | Payment token match |
| Identified (linked visits) | All behavioral data + visit history across locations | Recurring token pattern |
| Contactable | All above + email/phone | Guest opts in via loyalty program |

**Customer Lifetime Value (CLV) scoring** estimates each guest's projected future revenue based on visit frequency, average check size, and retention probability. CLV enables managers to prioritize high-value guest recovery over blanket marketing.

**RFM segmentation** classifies guests by Recency, Frequency, and Monetary value into actionable segments:

| Segment | Description | Recommended Action |
|---|---|---|
| Champion | Recent, frequent, high-spend | VIP treatment, exclusive previews |
| Loyal Regular | Frequent visitor, consistent spend | Reward consistency, upsell opportunities |
| At Risk | Previously frequent, declining visits | Targeted win-back campaign |
| New Discoverer | Recent first visit, single transaction | Welcome offer, encourage return |

**Churn prediction** uses a visit frequency decay model. The algorithm computes the average inter-visit interval from historical data and measures how many days overdue the guest is for their next expected visit:

- Days overdue <= 0: **Low risk** (5% churn probability)
- Days overdue <= 0.5x average interval: **Medium risk** (25% probability)
- Days overdue <= 1.5x average interval: **High risk** (60% probability)
- Days overdue > 1.5x average interval: **Critical risk** (90% probability)

The system requires a minimum of 3 visits to generate predictions, returning a conservative low-risk estimate for newer guests.

**Win-back alerts** are generated when high-CLV guests enter medium or higher churn risk tiers, providing the guest profile, estimated revenue at risk, and suggested win-back actions.

### 3.6 Labor Intelligence

**Problem:** Labor is the second-largest cost in restaurant operations (after food cost), yet scheduling is typically done manually based on manager intuition. Overstaffing wastes money; understaffing degrades service and burns out employees.

**FireLine's Approach:**

- **Effective Labor Unit (ELU) ratings.** Every employee is rated on a 0-5 scale per kitchen station, reflecting their speed, accuracy, and consistency at that station. A line cook rated 4.5 on grill and 2.0 on pastry is scheduled accordingly. ELU ratings update based on performance data, not just manager opinion.

- **Staff Point System.** A gamified performance tracking system with configurable earn and deduct rules. Points are awarded for positive behaviors (on-time arrival, completing prep lists, positive guest feedback) and deducted for negative ones (late arrival, food safety violations, no-show). Leaderboards create healthy competition across staff within and between locations.

- **Demand-based scheduling.** The scheduling engine generates shift recommendations by:
  1. Forecasting demand (covers per hour) based on historical patterns, day of week, seasonality, and known events
  2. Calculating required labor units per station based on forecasted demand and station capacity models
  3. Generating constraint-based shift assignments considering employee availability, ELU ratings, overtime limits, and labor law compliance
  4. Presenting the draft schedule for manager review and adjustment

- **Overtime detection and alerting.** The system tracks cumulative hours per employee across all locations and alerts managers when employees approach overtime thresholds, calculating the incremental cost impact.

- **Shift swap workflow.** Employees can request shift swaps through the tablet app, subject to manager approval. The system validates that the replacement employee has adequate ELU ratings for the assigned station.

### 3.7 Vendor Intelligence

**Problem:** Vendor selection in restaurants is often relationship-driven rather than data-driven. Price comparisons are done manually and sporadically. Delivery reliability, order accuracy, and quality consistency are tracked anecdotally if at all.

**FireLine's Approach:**

- **Vendor Reliability Score (VRS).** A composite score computed from four weighted sub-scores:

| Sub-Score | Weight | Measures |
|---|---|---|
| Price | 30% | Cost consistency vs. quoted prices |
| Delivery | 25% | On-time delivery rate |
| Quality | 25% | Rejection and return rate |
| Accuracy | 20% | Order accuracy (correct items, correct quantities) |

The overall score is calculated as: `price * 0.30 + delivery * 0.25 + quality * 0.25 + accuracy * 0.20`, rounded to two decimal places. Scores are recalculated from the trailing 90 days of received purchase order data.

- **OTIF (On-Time-In-Full) tracking.** Every delivery is compared against the original purchase order for both timeliness (on-time rate) and completeness (in-full rate). OTIF rate is the intersection -- deliveries that were both on time and complete.

- **Price intelligence.** Historical cost tracking per ingredient per vendor enables:
  - Anomaly detection: price increases that deviate from historical norms
  - Seasonal trend analysis: identifying predictable price fluctuations for budget planning
  - Cross-vendor comparison: real-time price benchmarking for the same ingredient across multiple suppliers

- **Vendor recommendation engine.** When generating purchase orders, the system recommends the optimal vendor for each ingredient based on a composite of price, reliability score, and lead time. The recommendation includes the data rationale: "Vendor A is 5% more expensive but has 98% OTIF vs. Vendor B's 71% OTIF."

- **Single-source risk detection.** The system identifies ingredients sourced from only one vendor and flags the supply chain risk, recommending secondary supplier qualification.

### 3.8 Operations Intelligence

**Problem:** Kitchen operations are managed by intuition and shouting. Station overload, ticket time degradation, and capacity bottlenecks are recognized only after service quality has already suffered.

**FireLine's Approach:**

- **Kitchen capacity model.** Each menu item has a **resource profile** defining which kitchen stations it uses, for how long, and with what skill requirements. The capacity model computes real-time station load by aggregating active ticket resource demands against station capacity.

- **KDS (Kitchen Display System).** Tickets are routed to station-specific displays based on menu item resource profiles. The KDS provides:
  - Station-specific views showing only relevant tickets
  - Color-coded urgency (green, yellow, red based on elapsed time vs. target)
  - Bump workflow for marking items as completed
  - Ticket time tracking from creation to completion

- **Overload detection with tiered autonomy.** The system classifies kitchen capacity into three states:

| Capacity | Threshold | Response |
|---|---|---|
| Normal | < 85% | No intervention |
| Elevated | 85-95% | System auto-throttles: increases quoted wait times, suggests temporary 86s for complex items |
| Critical | > 95% | Manager notification with suggested actions: pause delivery channels, activate additional staff, simplify menu to high-velocity items |

Each overload response tier specifies whether the action is auto-applied or requires manager approval, implementing tiered autonomy.

- **Operational health score.** A composite score (0-100) aggregating five sub-dimensions:
  - Kitchen health: station load distribution, equipment status
  - Ticket performance: average ticket time vs. target, completion rate
  - Staffing: scheduled vs. required labor, ELU coverage per station
  - Financial: real-time food cost vs. budget, revenue vs. forecast
  - Inventory: PAR coverage, expiry risk, variance severity

- **Five planning horizons.** FireLine structures operational awareness across temporal scales:

| Horizon | Timeframe | Key Metrics |
|---|---|---|
| Real-time | Now | Active tickets, station loads, overload status, health score |
| Shift | Next 4 hours | Forecasted covers, scheduled vs. required staff, expected revenue |
| Daily | Today | Prep items, expected deliveries, scheduled shifts, forecasted revenue |
| Weekly | This week | Total scheduled hours, pending POs, projected labor and revenue |
| Strategic | 30-day trailing | Revenue trends, COGS trends, guest count trajectory, vendor performance |

### 3.9 Marketing Intelligence

**Problem:** Restaurant marketing is often ad hoc -- a social media post here, a paper coupon there. There is no closed-loop measurement connecting marketing spend to guest behavior, and loyalty programs are typically one-size-fits-all.

**FireLine's Approach:**

- **Campaign engine.** Managers create campaigns with defined targets (customer segment, location, date range), distribution channels (email, SMS, in-app notification), and offer mechanics (percentage discount, fixed amount, free item, points multiplier). Campaign performance is tracked against redemption rates and incremental revenue.

- **Loyalty program.** A points-based system with four tier levels (Bronze, Silver, Gold, Platinum), each with configurable benefits. Points accrue based on spend and can be redeemed against menu items. Tier progression is automatic based on cumulative points within a rolling period.

- **Promotion simulation.** Before launching a campaign, managers can model the projected impact on revenue, margin, and customer behavior using historical response rates for similar promotions.

### 3.10 Multi-Location Intelligence

**Problem:** Multi-location operators need to compare performance across branches, identify outliers, and propagate best practices -- but data is typically siloed per location.

**FireLine's Approach:**

- **Portfolio hierarchy.** FireLine models organizational structure as Organization > Region > District > Location, supporting arbitrarily complex multi-location topologies including franchise groups with regional operators.

- **Cross-location benchmarking.** Every operational metric -- food cost percentage, labor cost percentage, average ticket time, guest satisfaction, equipment uptime -- is computed per location and ranked with percentile positioning. A location at the 25th percentile for ticket time immediately knows it has room to improve and can examine what the 90th percentile locations do differently.

- **Best practice detection.** When a location achieves sustained superior performance on a metric, the system identifies the operational differences (scheduling patterns, vendor choices, menu mix, prep procedures) that may explain the outperformance and surfaces them as recommendations for other locations.

- **Performance outlier identification.** The system flags locations whose metrics deviate significantly from portfolio averages, distinguishing between positive outliers (to study and replicate) and negative outliers (to investigate and remediate).

---

## 4. Department Operations -- How Each Team Uses FireLine

FireLine is designed for every role in a restaurant organization, from the CEO reviewing portfolio performance to the line cook bumping tickets on a kitchen display. Each role sees a tailored interface with the information and actions relevant to their responsibilities.

### 4.1 CEO / Owner

The CEO view provides portfolio-level oversight without requiring deep operational involvement in any single location.

**Portfolio Overview.** A single screen displays all 4 Nimbu branches with:
- Per-location health scores (composite 0-100)
- Daily revenue with trend arrows (vs. same day last week)
- Active alert count by priority (critical, high, medium, low)
- Top-level P&L summary (revenue, COGS%, labor%, gross margin)

**Executive Briefing.** An AI-generated daily briefing prioritizes the issues requiring the CEO's attention:
- Performance highlights: "Sheikh Zayed achieved 92% health score, highest in portfolio"
- Risk items: "North Coast cooler #3 showing compressor degradation trend -- maintenance ticket generated, EGP 45,000 inventory at risk"
- Strategic outlook: "Portfolio-wide food cost trending 1.2 points above budget; primary driver is protein costs at New Cairo (up 8% vs. last month)"

**AI Recommendations.** Actionable suggestions with projected impact and confidence scores:
- "Recommend adding Lamb Kofta to North Coast menu based on El Gouna performance (Powerhouse classification, EGP 85 contribution margin, 94% satisfaction). Confidence: 82%."
- "Consider renegotiating produce contract with Vendor B -- OTIF rate of 71% is 22 points below portfolio average. Estimated annual savings if switched to Vendor A: EGP 180,000."

**Action management.** The CEO can assign recommendations to specific team members, add comments, and track resolution status -- creating an accountability loop from insight to action.

### 4.2 General Manager (Per Branch)

The General Manager dashboard is the operational nerve center for a single location.

**Dashboard.** The GM's home screen displays:
- Revenue by hour (bar chart showing today vs. same day last week)
- Channel mix (dine-in / takeout / delivery split with margins per channel)
- Alert priority queue (sorted by severity, newest first)
- Kitchen pulse (current station loads, active ticket count, average ticket time)
- Top sellers (today's best-performing items by units sold and revenue)

**Financial drill-down.** From the P&L summary, the GM can drill into:
1. Summary P&L (revenue, COGS, labor, overhead, net margin)
2. Category-level COGS (proteins, dairy, produce, dry goods, beverages)
3. Item-level contribution (which menu items are driving costs)
4. Ingredient-level cost (which specific ingredients are over budget)
5. Vendor-level pricing (which vendor is causing the price increase)

**Scheduling.** The GM reviews AI-drafted schedules, adjusts based on local knowledge (staff preferences, upcoming events), approves shift swap requests, and monitors overtime accumulation across the team.

**Purchase orders.** When the system auto-generates purchase orders from PAR breaches, the GM reviews and approves them. The PO screen shows the recommended vendor, quantity, estimated cost, and the data rationale for the recommendation.

**Inventory variances.** The GM reviews variance reports with AI cause attribution, deciding which variances warrant investigation vs. write-off. High-severity variances (potential theft signals, large unexplained shortages) are flagged for immediate attention.

### 4.3 Kitchen Manager

The Kitchen Manager's interface is optimized for service flow and operational control.

**KDS (Kitchen Display System).** Station-specific ticket routing with:
- Color-coded urgency: green (within target), yellow (approaching target), red (exceeded target)
- Bump workflow: tap to mark items as completed and advance the ticket
- Aggregate view: all stations on one screen for the expeditor role

**Kitchen capacity monitoring.** Real-time station load bars showing current utilization as a percentage of capacity. When any station exceeds 85%, the display shifts to elevated mode with recommended actions (throttle incoming orders for that station, redistribute items to alternate stations if cross-trained staff are available).

**Resource profile management.** The Kitchen Manager defines how each menu item consumes kitchen resources:
- Which stations are involved (grill, saute, fry, cold, pastry)
- Estimated preparation time per station
- Required skill level (mapped to ELU ratings)
- Whether items can be prepped in advance (batch production)

**Equipment health monitoring.** The kitchen manager's dashboard includes equipment health indicators for all kitchen-critical equipment: coolers, freezers, grills, fryers, ovens, dishwashers. Degradation trends are visualized as trend lines, and the manager can view active maintenance tickets and scheduled preventive maintenance.

### 4.4 Inventory Manager

The Inventory Manager's workflow spans the tablet app (for physical operations) and the web dashboard (for analysis and decision-making).

**Physical counting workflow (tablet).** The tablet app presents ingredients grouped by storage location (walk-in cooler, dry storage, freezer, bar) for efficient walk-through counting. Key design decisions:
- Expected quantities are hidden to prevent anchoring bias
- Counts can be performed offline and synced when connectivity is restored
- Each count creates a timestamped snapshot for variance calculation
- Categories are organized to match the physical layout of the storage areas

**Waste logging (tablet).** Staff log waste events as they occur, categorizing each by reason:
- Expired
- Damaged (received or during prep)
- Overproduction
- Customer return
- Prep error

Each waste event captures the ingredient, quantity, reason, and optional photo documentation.

**Variance analysis (web dashboard).** After a count is completed, the variance engine computes per-ingredient variances with probabilistic cause attribution. The Inventory Manager reviews:
- Variance quantity (actual minus theoretical)
- Variance cost (dollar impact)
- Severity classification (low, medium, high, critical)
- Cause probabilities (portioning 62%, unrecorded waste 28%, theft signal 10%)

The manager can drill into any variance to see the underlying data: what was sold (from POS data), what was received (from purchase orders), what was wasted (from waste logs), and what was counted.

**PAR breach alerts.** Real-time alerts when ingredients fall below configured minimum stock levels, with one-tap approval to send the auto-generated purchase order.

**Expiry tracking.** A priority list of ingredients approaching expiry, sorted by expiry date, with batch numbers and dollar value at risk. Enables FIFO enforcement and proactive menu promotion of items nearing expiry.

**Delivery receiving (tablet).** Line-by-line verification of incoming deliveries against purchase orders. The tablet displays each expected line item; staff enter actual received quantities. Variances (short deliveries, substitutions, quality rejections) are flagged immediately, updating the vendor's accuracy and quality sub-scores.

### 4.5 Front-of-House Manager

**Customer intelligence.** The FOH Manager accesses guest profiles showing:
- Visit history across all locations
- CLV score and tier (Champion, Loyal Regular, At Risk, etc.)
- Preferred items and daypart patterns
- Churn risk status with recommended actions

**Marketing campaigns.** Create and manage promotional campaigns:
- Define target segment (e.g., "At Risk guests with CLV > EGP 5,000")
- Set offer mechanics (20% discount on next visit, double loyalty points)
- Schedule campaign window
- Track redemption rate and incremental revenue

**Loyalty program management.** Monitor loyalty program metrics:
- Enrollment rate by location
- Points issued vs. redeemed
- Tier distribution (percentage of guests at each tier)
- Program ROI (incremental revenue from loyalty members vs. non-members)

### 4.6 Maintenance Team

**Equipment registry.** A complete inventory of all restaurant equipment with:
- Equipment type, manufacturer, model, serial number
- Installation date and warranty status
- Current health score (0-100)
- Maintenance history (all past work orders)

**AI-generated preventive maintenance.** When sensor trends indicate degradation, the system creates maintenance tickets with:
- Root cause analysis: "Compressor duty cycle increasing 3% per week over last 4 weeks"
- Priority classification based on risk assessment
- Recommended actions and estimated time to complete
- Inventory at risk (for cold storage equipment)

**Work order lifecycle.** Maintenance tickets follow a structured workflow:
- **Open:** Created by AI or manually by staff
- **In Progress:** Assigned to a technician, work started
- **Completed:** Work finished, resolution documented, parts and cost recorded

**Maintenance schedule.** Calendar view of all scheduled preventive maintenance, with overdue tasks highlighted. The system tracks compliance rate (percentage of scheduled maintenance completed on time).

**Cost tracking.** All maintenance costs are recorded per equipment item, enabling total cost of ownership analysis and replacement timing decisions.

### 4.7 Finance / Accounting

**Budget management.** Set targets at the location level for:
- Revenue (daily, weekly, monthly)
- Food cost percentage (target and maximum acceptable)
- Labor cost percentage (target and maximum acceptable)
- Overhead and net margin targets

**Budget vs. actual variance reports.** Real-time comparison of actual performance against budget with variance calculation and status classification (on_track / over / under). Reports can be generated for any time period.

**Cost center analysis.** COGS broken down by ingredient category with trend analysis. Identify which categories are driving cost increases and drill into specific ingredients and vendors.

**Transaction anomaly detection.** The finance team receives alerts for statistically significant anomalies:
- Void counts exceeding 2 standard deviations from the 30-day baseline
- Discount rates spiking above historical norms
- Off-hours transaction activity
- Unusual comp patterns

Each alert includes the Z-score, baseline value, current value, and severity, providing the statistical rigor needed for investigation decisions.

**Vendor invoice matching.** Purchase orders, delivery receipts, and vendor invoices are compared to detect discrepancies before payment, preventing overpayment for short deliveries or rejected items.

---

## 5. Benefits of the AI Intelligence Layer

### 5.1 Quantifiable Benefits

**Food cost reduction: 2-5%.** FireLine attacks food cost from multiple angles simultaneously:
- Variance detection identifies the specific ingredients where waste is occurring and the probable causes, enabling targeted corrective action rather than blanket cost-cutting
- Portioning alerts flag ingredients where actual usage consistently exceeds theoretical usage, indicating over-portioning before it becomes a systemic problem
- Automated reordering prevents both stockouts (lost sales) and over-ordering (increased waste), maintaining optimal inventory levels
- Expiry tracking with FIFO enforcement reduces spoilage
- For a restaurant chain with EGP 5M monthly food cost, a 3% improvement represents EGP 150,000 per month in recovered margin

**Labor cost optimization: 3-8%.** Demand-based scheduling replaces intuition-based scheduling:
- Historical demand patterns and day-of-week trends generate accurate labor forecasts
- ELU ratings ensure the right employees are assigned to the right stations, reducing the need for overstaffing to compensate for skill gaps
- Overtime detection alerts managers before overtime is incurred, not after
- Shift gap analysis identifies periods where staffing exceeds demand, enabling schedule optimization
- For a chain with EGP 2M monthly labor cost, a 5% improvement represents EGP 100,000 per month

**Revenue protection: 8-12% VIP churn prevention.** The churn prediction model identifies high-CLV guests at risk of leaving before they are lost:
- A guest with EGP 50,000 annual CLV entering "high risk" churn status represents EGP 50,000 in revenue at risk
- Targeted win-back campaigns for at-risk high-CLV guests have significantly higher ROI than untargeted marketing
- Cross-location guest recognition enables consistent VIP treatment regardless of which branch the guest visits

**Equipment uptime: 15-25% reduction in unplanned downtime.** Predictive maintenance transforms equipment management from reactive to proactive:
- Early detection of compressor degradation can prevent walk-in cooler failure and the associated inventory loss (typically EGP 30,000-80,000 per incident)
- Scheduled preventive maintenance during off-hours avoids service disruption
- Equipment health scoring enables data-driven replacement timing, avoiding both premature replacement (wasted capital) and delayed replacement (increased failure risk)

**Waste reduction: 20-30%.** Multiple intelligence modules contribute to waste reduction:
- Expiry tracking with shelf-life alerts ensures FIFO compliance
- Variance categorization identifies systematic waste patterns (overproduction, prep errors) for process improvement
- Menu simulation enables managers to model the impact of removing low-velocity items that generate disproportionate waste
- Demand forecasting (future enhancement) will enable prep quantity optimization

### 5.2 Qualitative Benefits

**Decision speed: from hours to seconds.** Traditional restaurant analysis requires exporting data from multiple systems, combining it in spreadsheets, and manually identifying trends. FireLine surfaces the insight directly: "Food cost at New Cairo is 2.3 points above budget, driven by protein costs. Chicken breast usage variance is 18% (portioning probable cause: 72%). Recommend portioning audit this week."

**Consistency across locations.** Standardized processes, scoring algorithms, and reporting formats ensure that "good" means the same thing at every location. A health score of 85 at El Gouna is directly comparable to a health score of 85 at Sheikh Zayed.

**Proactive vs. reactive operations.** FireLine shifts the operational paradigm from discovering problems after they have caused damage to preventing them before they occur:
- Maintenance alerts before equipment fails
- Churn prediction before guests leave
- PAR breach alerts before stockouts affect service
- Overload detection before ticket times degrade

**Data-driven culture.** When every decision is backed by real-time analytics with confidence scores, the organizational culture shifts from opinion-based management ("I think we should...") to evidence-based management ("The data shows..."). This is particularly valuable in franchise environments where standardization across owner-operators is critical.

**Scalability.** Adding a new location to FireLine requires minimal operational overhead: create the location in the portfolio hierarchy, configure location-specific parameters (operating hours, station layout, PAR levels), and the entire intelligence layer -- from menu scoring to vendor recommendations to labor scheduling -- is immediately available. The system learns the location's patterns within weeks of operation.

### 5.3 Competitive Advantages

**Cross-module intelligence.** This is FireLine's most significant competitive moat. In a fragmented technology landscape, inventory, financial, labor, and customer systems operate independently. In FireLine, they share a common event bus and data layer:
- An inventory variance triggers a financial COGS recalculation and a vendor accuracy score update
- A customer churn alert includes the guest's preferred menu items, informing the win-back offer
- A menu item removal simulation shows the impact on ingredient demand (affecting purchase orders), kitchen station load (affecting staffing), and vendor contracts (affecting procurement)

No single-domain point solution can replicate this compound intelligence.

**Probabilistic reasoning.** Most restaurant software generates binary alerts: "variance detected" or "not detected." FireLine's probabilistic cause attribution provides the **why** alongside the **what**, enabling managers to take the correct remedial action. A 62% probability of portioning error leads to a portioning audit. A 45% probability of theft signal leads to a security review. These are fundamentally different responses to the same variance alert.

**Tiered autonomy.** FireLine applies AI autonomy proportionally to decision risk:
- Low-risk, high-frequency decisions are automated (reorder stock, adjust quoted wait times)
- Medium-risk decisions surface as recommendations with projected impact (schedule optimization, vendor selection)
- High-risk decisions require explicit manager approval (menu price changes, equipment replacement, staff disciplinary actions)

This approach builds trust by demonstrating AI competence on routine decisions while preserving human judgment where it matters most.

**Continuous learning.** Every operational cycle (count, service period, delivery, maintenance event) generates data that refines the system's models. Variance categorization becomes more accurate as the system accumulates more examples. Churn prediction improves as the visit frequency model processes more guest histories. Vendor scores become more reliable as the 90-day trailing window incorporates more deliveries.

---

## 6. Technical Innovation

### Event-Driven Cross-Module Communication

FireLine's event bus implements NATS-compatible subject naming with wildcard routing in a synchronous in-process architecture. Every domain module publishes events when state changes occur and subscribes to events from other modules that affect its computations. The event envelope carries schema versioning, enabling backward-compatible evolution of event payloads without breaking downstream subscribers.

This design achieves the composability benefits of event-driven microservices while maintaining the transactional consistency and operational simplicity of a monolithic deployment. The event bus supports middleware chains for observability, and a dead letter queue captures failed events for investigation.

### Row-Level Security for Zero-Trust Multi-Tenant Isolation

FireLine's multi-tenant isolation goes beyond application-level filtering. PostgreSQL RLS policies enforce data boundaries at the database engine level, meaning that even a bug in application code cannot cause cross-tenant data leakage. The dual-hook connection pool design (set tenant context on acquire, clear on release, destroy on failure) implements defense-in-depth with a fail-closed posture.

This is a genuine differentiator in the restaurant SaaS space, where most competitors rely on application-level tenant filtering or, worse, separate databases per tenant (which prevents cross-tenant analytics for portfolio operators).

### Probabilistic Cause Attribution in Variance Analysis

The variance categorization engine is, to our knowledge, unique in restaurant technology. Rather than presenting raw variance numbers, the engine evaluates multiple signals (variance magnitude, waste log correlation, portioning history, vendor delivery timing) and computes independent probability scores for each potential cause. This transforms variance analysis from a data dump requiring manual interpretation into an actionable diagnosis.

### Demand-Forecast-Driven Automated Purchasing

FireLine closes the loop between inventory intelligence and procurement. When stock falls below PAR levels, the system automatically generates purchase orders with optimal vendor selection (based on the vendor intelligence module's composite scoring) and optimal quantities (based on usage rates and lead times). This eliminates the delay between "we need more chicken" and "the order is placed," reducing both stockout risk and manual procurement effort.

### Anti-Anchoring Bias in Physical Counting

A subtle but important innovation: FireLine's tablet-based counting workflow deliberately hides expected quantities from the counting staff. Research in behavioral economics demonstrates that anchoring bias causes people to adjust their counts toward expected values rather than counting accurately. By removing the anchor, FireLine produces more accurate physical counts, which in turn produces more accurate variance analysis, which in turn produces more accurate cause attribution.

### Tiered Autonomy in Kitchen Overload Response

The overload response system implements a graduated autonomy model where the system's authority to act autonomously is proportional to the reversibility of the action. Auto-throttling quoted wait times (easily reversible, low impact) is automatic. Recommending temporary menu simplification (moderate impact) requires manager acknowledgment. Suggesting delivery channel shutdown (high impact, affects partner relationships) requires explicit manager approval. This design builds trust while still providing value during high-pressure service periods when managers are least available for decision-making.

---

## 7. Roadmap and Future Vision

### Near-Term (Q2-Q3 2026)

**Python ML service with Prophet demand forecasting.** The current demand estimation uses historical averages. A dedicated Python ML service will implement Facebook Prophet for time-series demand forecasting incorporating day-of-week patterns, seasonal trends, weather data, local events, and holiday calendars. This will improve both labor scheduling accuracy and inventory procurement precision.

**Real POS integration.** The current adapter framework includes a Toast integration reference implementation. Near-term priorities include production-grade integrations with Toast, Square, and Clover, enabling FireLine to ingest real-time transaction data from the most widely deployed POS systems. The adapter registry architecture is designed for this expansion -- each new POS integration is a self-contained adapter conforming to a standard interface.

**Mobile app for managers.** A native mobile application providing managers with on-the-go access to dashboards, alert management, and approval workflows. The app will support push notifications for critical alerts (equipment failure, transaction anomalies, critical churn risk).

### Medium-Term (Q4 2026 - Q1 2027)

**Computer vision for waste detection.** Camera-based waste monitoring at disposal points, using object detection models to automatically classify and quantify waste events. This eliminates reliance on manual waste logging, which suffers from underreporting. Waste images are analyzed to identify ingredient type, estimated quantity, and probable cause, feeding directly into the variance categorization engine.

**Voice-activated kitchen commands.** Hands-free KDS interaction using wake-word detection and natural language understanding. Kitchen staff can bump tickets, check order status, and report equipment issues without touching screens -- critical in a fast-paced, hands-dirty kitchen environment.

**Advanced customer sentiment analysis.** NLP-based analysis of review data (Google Reviews, TripAdvisor, social media mentions) correlated with guest profiles to enrich the customer intelligence module. Sentiment scores feed into the menu satisfaction dimension and inform win-back campaign messaging.

### Long-Term (2027+)

**Blockchain-based supply chain verification.** Immutable provenance tracking for high-value and high-risk ingredients (imported seafood, organic produce, halal certification). Each handoff in the supply chain is recorded on a distributed ledger, providing verifiable chain-of-custody from farm to table.

**SOC 2 Type II compliance.** Formal security audit and compliance certification, required for enterprise restaurant chain deployments. FireLine's existing security architecture (RLS, encrypted tokens, role-based access control) provides a strong foundation; SOC 2 adds the process and audit framework.

**Franchise-specific features.** Royalty calculation engines, franchisee performance scorecards, brand compliance monitoring, and multi-tier administrative hierarchies designed for franchise operators managing hundreds of locations across multiple brands.

**Autonomous operations mode.** For mature deployments with sufficient historical data, FireLine will support fully autonomous operation of routine tasks: inventory reordering, schedule generation, preventive maintenance scheduling, and campaign targeting -- with human oversight reduced to exception handling and strategic decisions.

---

## Appendix A: Nimbu Demo Deployment

The initial demonstration deployment of FireLine operates under the brand **Nimbu**, a 4-branch international restaurant chain in Egypt.

| Branch | Location | Characteristics |
|---|---|---|
| El Gouna | Red Sea resort town | Tourism-driven demand, seasonal variation, high seafood usage |
| New Cairo | Urban residential/commercial | Consistent weekday demand, delivery-heavy, diverse menu |
| Sheikh Zayed | Western Cairo suburb | Family-oriented, weekend peaks, dine-in focused |
| North Coast | Mediterranean coastal | Extreme seasonality (summer peak), high-volume weekends |

The Nimbu deployment validates FireLine across diverse demand patterns, geographic distribution, and operational challenges, demonstrating the platform's adaptability to different restaurant contexts within a single portfolio.

---

## Appendix B: System Metrics

| Metric | Value |
|---|---|
| Backend codebase | Go modular monolith, 15+ domain packages |
| Database migrations | 18 versioned migrations |
| API endpoints | 120+ RESTful endpoints |
| Web dashboard pages | 20 pages with tabbed sub-views |
| Tablet app screens | 5 screens (Count, Waste, Receive, KDS, Clock) |
| Test packages | 21 packages, 0 failures |
| Development sprints completed | 20 |

---

*FireLine by OpsNerve -- The Operational Nervous System for Modern Restaurant Chains.*

*For inquiries: contact OpsNerve at [info@opsnerve.com](mailto:info@opsnerve.com)*
