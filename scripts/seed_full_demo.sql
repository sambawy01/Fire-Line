-- ============================================================================
-- FireLine Full Demo Seed Script
-- Populates ALL intelligence modules with realistic restaurant data
-- Run: docker exec -i fireline-postgres-1 psql -U fireline -d fireline < scripts/seed_full_demo.sql
-- ============================================================================

-- Disable RLS for seeding
SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

BEGIN;

-- ============================================================================
-- CONSTANTS
-- ============================================================================
-- org_id:       3f7ef589-f499-43e3-a1c5-aaacd9d543ec
-- downtown:     a1111111-1111-1111-1111-111111111111
-- airport:      b2222222-2222-2222-2222-222222222222
-- user_id:      0d55e810-1e4a-417a-8a70-08b98f4595c2
-- Downtown employees: ee111111-1111-*, ee111111-2222-*, ee111111-3333-*, ee111111-4444-*, ee111111-5555-*
-- Airport employees:  ee222222-1111-*, ee222222-2222-*, ee222222-3333-*
-- Downtown menu items: 22222222-aaaa-{1111..6666}-aaaa-111111111111
-- Airport menu items:  33333333-aaaa-{1111..5555}-aaaa-111111111111

-- ============================================================================
-- 1. ORDERS (past 30 days) - PL/pgSQL DO block
-- ============================================================================
DO $$
DECLARE
    v_org_id      uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_downtown    uuid := 'a1111111-1111-1111-1111-111111111111';
    v_airport     uuid := 'b2222222-2222-2222-2222-222222222222';
    v_day         date;
    v_check_id    uuid;
    v_check_item_id uuid;
    v_num_orders  int;
    v_channel     text;
    v_channels    text[] := ARRAY['dine_in','takeout','delivery'];
    v_dt_items    uuid[] := ARRAY[
        '22222222-aaaa-1111-aaaa-111111111111'::uuid,
        '22222222-aaaa-2222-aaaa-111111111111'::uuid,
        '22222222-aaaa-3333-aaaa-111111111111'::uuid,
        '22222222-aaaa-4444-aaaa-111111111111'::uuid,
        '22222222-aaaa-5555-aaaa-111111111111'::uuid,
        '22222222-aaaa-6666-aaaa-111111111111'::uuid
    ];
    v_dt_names    text[] := ARRAY['Classic Burger','Bacon Avocado Burger','Grilled Chicken Sandwich','Caesar Salad','Loaded Fries','Side Salad'];
    v_dt_prices   int[] := ARRAY[1495,1795,1395,1195,895,595];
    v_ap_items    uuid[] := ARRAY[
        '33333333-aaaa-1111-aaaa-111111111111'::uuid,
        '33333333-aaaa-2222-aaaa-111111111111'::uuid,
        '33333333-aaaa-3333-aaaa-111111111111'::uuid,
        '33333333-aaaa-4444-aaaa-111111111111'::uuid,
        '33333333-aaaa-5555-aaaa-111111111111'::uuid
    ];
    v_ap_names    text[] := ARRAY['Classic Burger','Bacon Avocado Burger','Grilled Chicken Sandwich','Caesar Salad','Loaded Fries'];
    v_ap_prices   int[] := ARRAY[1695,1995,1595,1395,1095];
    v_item_idx    int;
    v_num_items   int;
    v_subtotal    int;
    v_tax         int;
    v_total       int;
    v_tip         int;
    v_price       int;
    v_qty         int;
    v_opened      timestamptz;
    v_method      text;
    v_methods     text[] := ARRAY['card','card','card','cash','card'];
    v_loc_id      uuid;
    v_items_arr   uuid[];
    v_names_arr   text[];
    v_prices_arr  int[];
    v_order_num   int := 5000;
    i int; j int; k int;
BEGIN
    FOR v_day IN SELECT generate_series(CURRENT_DATE - 30, CURRENT_DATE - 1, '1 day')::date LOOP
        -- Process both locations
        FOR k IN 1..2 LOOP
            IF k = 1 THEN
                v_loc_id := v_downtown;
                v_items_arr := v_dt_items;
                v_names_arr := v_dt_names;
                v_prices_arr := v_dt_prices;
                -- More orders on weekends
                IF EXTRACT(DOW FROM v_day) IN (0,6) THEN
                    v_num_orders := 60 + floor(random()*20)::int;
                ELSE
                    v_num_orders := 40 + floor(random()*20)::int;
                END IF;
            ELSE
                v_loc_id := v_airport;
                v_items_arr := v_ap_items;
                v_names_arr := v_ap_names;
                v_prices_arr := v_ap_prices;
                IF EXTRACT(DOW FROM v_day) IN (0,6) THEN
                    v_num_orders := 30 + floor(random()*15)::int;
                ELSE
                    v_num_orders := 20 + floor(random()*15)::int;
                END IF;
            END IF;

            FOR i IN 1..v_num_orders LOOP
                v_check_id := gen_random_uuid();
                v_channel := v_channels[1 + floor(random()*3)::int];
                -- Spread orders across the day (10am-10pm)
                v_opened := v_day + (interval '10 hours') + (random() * interval '12 hours');
                v_subtotal := 0;
                v_num_items := 1 + floor(random()*3)::int;
                v_order_num := v_order_num + 1;

                INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel,
                    subtotal, tax, total, tip, discount, opened_at, closed_at, source, created_at)
                VALUES (v_check_id, v_org_id, v_loc_id, 'ORD-' || v_order_num, 'closed', v_channel,
                    0, 0, 0, 0, 0, v_opened, v_opened + interval '25 minutes' + (random() * interval '20 minutes'),
                    'manual', v_opened);

                FOR j IN 1..v_num_items LOOP
                    v_item_idx := 1 + floor(random() * array_length(v_items_arr, 1))::int;
                    v_qty := 1 + floor(random()*2)::int;
                    v_price := v_prices_arr[v_item_idx];
                    v_subtotal := v_subtotal + (v_price * v_qty);

                    INSERT INTO check_items (org_id, check_id, menu_item_id, name, quantity, unit_price, created_at)
                    VALUES (v_org_id, v_check_id, v_items_arr[v_item_idx], v_names_arr[v_item_idx], v_qty, v_price, v_opened);
                END LOOP;

                v_tax := (v_subtotal * 0.0825)::int;
                v_total := v_subtotal + v_tax;
                v_tip := CASE WHEN v_channel = 'dine_in' THEN (v_subtotal * (0.15 + random()*0.10))::int ELSE 0 END;
                v_method := v_methods[1 + floor(random()*5)::int];

                UPDATE checks SET subtotal = v_subtotal, tax = v_tax, total = v_total, tip = v_tip
                WHERE check_id = v_check_id;

                INSERT INTO payments (org_id, check_id, amount, tip, method, status, created_at)
                VALUES (v_org_id, v_check_id, v_total, v_tip, v_method, 'completed', v_opened + interval '30 minutes');
            END LOOP;
        END LOOP;
    END LOOP;
    RAISE NOTICE 'Orders seeded: % orders generated', v_order_num - 5000;
END $$;

-- ============================================================================
-- 2. INVENTORY COUNTS (3 weeks)
-- ============================================================================
INSERT INTO inventory_counts (count_id, org_id, location_id, counted_by, count_type, status, started_at, submitted_at, approved_by, approved_at, created_at, updated_at) VALUES
('cc111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ee111111-3333-3333-3333-333333333333', 'full', 'approved', CURRENT_DATE - 21, CURRENT_DATE - 21, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 21, CURRENT_DATE - 21, CURRENT_DATE - 21),
('cc222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ee111111-3333-3333-3333-333333333333', 'full', 'approved', CURRENT_DATE - 14, CURRENT_DATE - 14, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14, CURRENT_DATE - 14, CURRENT_DATE - 14),
('cc333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ee111111-4444-4444-4444-444444444444', 'full', 'approved', CURRENT_DATE - 7, CURRENT_DATE - 7, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7, CURRENT_DATE - 7, CURRENT_DATE - 7);

-- Count 1 (3 weeks ago) - some items short
INSERT INTO inventory_count_lines (org_id, count_id, location_id, ingredient_id, expected_qty, counted_qty, unit, note) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 50.0, 44.5, 'lb', 'Ground beef short'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-2222-aaaa-111111111111', 30.0, 27.0, 'lb', 'Chicken slightly under'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 15.0, 13.5, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 20.0, 17.0, 'head', 'Some wilted, discarded'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 25.0, 23.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 120.0, 110.0, 'ea', 'Missing 10 buns'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-7777-aaaa-111111111111', 40.0, 38.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-8888-aaaa-111111111111', 40.0, 36.0, 'ea', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 20.0, 17.5, 'lb', 'Bacon short'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc111111-1111-1111-1111-111111111111', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-aaaa-aaaa-111111111111', 64.0, 60.0, 'oz', NULL);

-- Count 2 (2 weeks ago) - improving
INSERT INTO inventory_count_lines (org_id, count_id, location_id, ingredient_id, expected_qty, counted_qty, unit, note) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 48.0, 46.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-2222-aaaa-111111111111', 28.0, 27.5, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 14.0, 13.5, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 18.0, 17.0, 'head', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 24.0, 23.5, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 100.0, 98.0, 'ea', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-7777-aaaa-111111111111', 38.0, 37.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-8888-aaaa-111111111111', 35.0, 34.0, 'ea', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 18.0, 17.5, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc222222-2222-2222-2222-222222222222', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-aaaa-aaaa-111111111111', 60.0, 59.0, 'oz', NULL);

-- Count 3 (1 week ago) - mostly on target
INSERT INTO inventory_count_lines (org_id, count_id, location_id, ingredient_id, expected_qty, counted_qty, unit, note) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 52.0, 51.5, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-2222-aaaa-111111111111', 32.0, 31.5, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 16.0, 16.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 22.0, 21.5, 'head', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 26.0, 26.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 110.0, 109.0, 'ea', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-7777-aaaa-111111111111', 42.0, 42.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-8888-aaaa-111111111111', 38.0, 37.5, 'ea', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 22.0, 22.0, 'lb', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc333333-3333-3333-3333-333333333333', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-aaaa-aaaa-111111111111', 64.0, 63.5, 'oz', NULL);

-- ============================================================================
-- 3. WASTE LOGS (30 entries over past 30 days)
-- ============================================================================
INSERT INTO waste_logs (org_id, location_id, ingredient_id, quantity, unit, reason, logged_by, logged_at, note) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 3.5, 'lb', 'expired', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 28 + time '09:00', 'Found in walk-in, past use-by date'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 2.0, 'head', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 27 + time '08:30', 'Wilted lettuce'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-2222-aaaa-111111111111', 1.5, 'lb', 'overcooked', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 26 + time '13:15', 'Chicken dried out on grill'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-7777-aaaa-111111111111', 2.0, 'lb', 'dropped', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 25 + time '12:00', 'Basket fell off counter'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 3.0, 'lb', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 24 + time '09:30', 'Tomatoes gone soft'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-8888-aaaa-111111111111', 4.0, 'ea', 'expired', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 23 + time '10:00', 'Avocados overripe'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 2.0, 'lb', 'overcooked', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 22 + time '14:00', 'Burnt patties during rush'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 1.0, 'lb', 'contaminated', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 21 + time '11:00', 'Cross-contamination concern'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 8.0, 'ea', 'expired', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 20 + time '09:00', 'Stale buns'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 0.5, 'lb', 'dropped', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 19 + time '12:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-7777-aaaa-111111111111', 1.5, 'lb', 'overcooked', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 18 + time '13:45', 'Fries burnt'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 1.0, 'lb', 'overproduction', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 17 + time '21:00', 'End of night excess prep'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-2222-aaaa-111111111111', 1.0, 'lb', 'overproduction', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 16 + time '21:30', 'Too much chicken prepped'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-aaaa-aaaa-111111111111', 8.0, 'oz', 'expired', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 15 + time '09:00', 'Dressing past date'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-8888-aaaa-111111111111', 3.0, 'ea', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 14 + time '10:00', 'Overripe avocados'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 1.0, 'head', 'expired', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 13 + time '08:45', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 1.5, 'lb', 'dropped', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 12 + time '12:15', 'Patty dropped on floor'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 2.0, 'lb', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 11 + time '09:15', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-7777-aaaa-111111111111', 1.0, 'lb', 'overcooked', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 10 + time '13:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 0.5, 'lb', 'overcooked', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 9 + time '14:00', 'Bacon burned'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 5.0, 'ea', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 8 + time '09:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-2222-aaaa-111111111111', 0.75, 'lb', 'dropped', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 7 + time '12:45', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 0.25, 'lb', 'dropped', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 6 + time '11:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 2.0, 'lb', 'overproduction', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 5 + time '21:15', 'Slow night'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-8888-aaaa-111111111111', 2.0, 'ea', 'expired', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 4 + time '10:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 1.0, 'head', 'expired', 'ee111111-5555-5555-5555-555555555555', CURRENT_DATE - 3 + time '08:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-7777-aaaa-111111111111', 1.5, 'lb', 'dropped', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 2 + time '12:00', 'Spilled near fryer'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 1.5, 'lb', 'expired', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 1 + time '09:00', 'Soft tomatoes');

-- ============================================================================
-- 4. PURCHASE ORDERS
-- ============================================================================
-- PO1: US Foods - received with some discrepancies
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, received_by, received_at, total_estimated, total_actual, notes, created_at, updated_at) VALUES
('dd111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'US Foods', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 18, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 16, 52500, 54800, 'Weekly protein order', CURRENT_DATE - 20, CURRENT_DATE - 16);

INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost, received_qty, received_unit_cost, variance_qty, variance_flag, received_at) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 60.0, 'lb', 450, 58.0, 460, -2.0, 'short', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd111111-1111-1111-1111-111111111111', '11111111-aaaa-2222-aaaa-111111111111', 40.0, 'lb', 375, 40.0, 385, 0.0, 'exact', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 25.0, 'lb', 680, 24.0, 695, -1.0, 'short', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd111111-1111-1111-1111-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 20.0, 'lb', 520, 20.0, 520, 0.0, 'exact', CURRENT_DATE - 16);

-- PO2: Sysco - received
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, received_by, received_at, total_estimated, total_actual, notes, created_at, updated_at) VALUES
('dd222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 12, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 10, 38200, 39100, 'Dry goods and produce', CURRENT_DATE - 14, CURRENT_DATE - 10);

INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost, received_qty, received_unit_cost, variance_qty, variance_flag, received_at) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd222222-2222-2222-2222-222222222222', '11111111-aaaa-6666-aaaa-111111111111', 200.0, 'ea', 85, 200.0, 85, 0.0, 'exact', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd222222-2222-2222-2222-222222222222', '11111111-aaaa-7777-aaaa-111111111111', 60.0, 'lb', 195, 55.0, 200, -5.0, 'short', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd222222-2222-2222-2222-222222222222', '11111111-aaaa-aaaa-aaaa-111111111111', 128.0, 'oz', 25, 128.0, 25, 0.0, 'exact', CURRENT_DATE - 10);

-- PO3: Local Farm Co - approved, awaiting delivery
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, total_estimated, total_actual, notes, created_at, updated_at) VALUES
('dd333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Local Farm Co', 'approved', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1, 18500, 0, 'Fresh produce delivery', CURRENT_DATE - 2, CURRENT_DATE - 1);

INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd333333-3333-3333-3333-333333333333', '11111111-aaaa-4444-aaaa-111111111111', 30.0, 'head', 185),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd333333-3333-3333-3333-333333333333', '11111111-aaaa-5555-aaaa-111111111111', 35.0, 'lb', 260),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd333333-3333-3333-3333-333333333333', '11111111-aaaa-8888-aaaa-111111111111', 50.0, 'ea', 150);

-- PO4: System recommended - draft
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, suggested_at, total_estimated, total_actual, notes, created_at, updated_at) VALUES
('dd444444-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'US Foods', 'draft', 'system_recommended', NOW(), 45600, 0, 'Auto-generated based on projected demand and current inventory levels', CURRENT_DATE, CURRENT_DATE);

INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd444444-4444-4444-4444-444444444444', '11111111-aaaa-1111-aaaa-111111111111', 55.0, 'lb', 450),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd444444-4444-4444-4444-444444444444', '11111111-aaaa-2222-aaaa-111111111111', 35.0, 'lb', 375),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd444444-4444-4444-4444-444444444444', '11111111-aaaa-9999-aaaa-111111111111', 20.0, 'lb', 680);

-- ============================================================================
-- 5. INVENTORY VARIANCES
-- ============================================================================
INSERT INTO inventory_variances (org_id, location_id, ingredient_id, count_id, period_start, period_end, theoretical_usage, actual_usage, variance_qty, variance_cents, cause_probabilities, severity) VALUES
-- Count 1 variances (worst)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 'cc111111-1111-1111-1111-111111111111', CURRENT_DATE - 28, CURRENT_DATE - 21, 45.0, 50.5, 5.5, 2475, '{"waste":0.35,"theft":0.15,"portioning":0.40,"recording_error":0.10}', 'warning'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 'cc111111-1111-1111-1111-111111111111', CURRENT_DATE - 28, CURRENT_DATE - 21, 17.0, 20.0, 3.0, 600, '{"waste":0.60,"spoilage":0.30,"recording_error":0.10}', 'warning'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 'cc111111-1111-1111-1111-111111111111', CURRENT_DATE - 28, CURRENT_DATE - 21, 16.0, 18.5, 2.5, 1700, '{"portioning":0.50,"waste":0.30,"recording_error":0.20}', 'warning'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 'cc111111-1111-1111-1111-111111111111', CURRENT_DATE - 28, CURRENT_DATE - 21, 100.0, 110.0, 10.0, 850, '{"waste":0.45,"portioning":0.35,"recording_error":0.20}', 'info'),
-- Count 2 variances (improving)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 'cc222222-2222-2222-2222-222222222222', CURRENT_DATE - 21, CURRENT_DATE - 14, 44.0, 46.0, 2.0, 900, '{"portioning":0.50,"waste":0.30,"recording_error":0.20}', 'info'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 'cc222222-2222-2222-2222-222222222222', CURRENT_DATE - 21, CURRENT_DATE - 14, 16.0, 17.0, 1.0, 200, '{"waste":0.50,"recording_error":0.50}', 'info'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-9999-aaaa-111111111111', 'cc222222-2222-2222-2222-222222222222', CURRENT_DATE - 21, CURRENT_DATE - 14, 17.0, 17.5, 0.5, 340, '{"portioning":0.60,"recording_error":0.40}', 'info'),
-- Count 3 variances (on target)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 'cc333333-3333-3333-3333-333333333333', CURRENT_DATE - 14, CURRENT_DATE - 7, 50.0, 50.5, 0.5, 225, '{"recording_error":0.70,"portioning":0.30}', 'info'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 'cc333333-3333-3333-3333-333333333333', CURRENT_DATE - 14, CURRENT_DATE - 7, 21.0, 21.5, 0.5, 100, '{"recording_error":0.80,"waste":0.20}', 'info');

-- ============================================================================
-- 6. BUDGETS
-- ============================================================================
INSERT INTO budgets (org_id, location_id, period_type, period_start, period_end, revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target, created_by) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'monthly', '2026-03-01', '2026-03-31', 5000000, 30.00, 28.00, 1500000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'weekly', date_trunc('week', CURRENT_DATE)::date, (date_trunc('week', CURRENT_DATE) + interval '6 days')::date, 1200000, 30.00, 28.00, 360000, '0d55e810-1e4a-417a-8a70-08b98f4595c2');

-- ============================================================================
-- 7. SHIFTS (past 30 days for Downtown employees)
-- ============================================================================
DO $$
DECLARE
    v_org_id    uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_loc_id    uuid := 'a1111111-1111-1111-1111-111111111111';
    v_day       date;
    v_emps      uuid[] := ARRAY[
        'ee111111-1111-1111-1111-111111111111'::uuid,
        'ee111111-2222-2222-2222-222222222222'::uuid,
        'ee111111-3333-3333-3333-333333333333'::uuid,
        'ee111111-4444-4444-4444-444444444444'::uuid,
        'ee111111-5555-5555-5555-555555555555'::uuid
    ];
    v_roles     text[] := ARRAY['gm','shift_manager','staff','staff','staff'];
    v_rates     int[] := ARRAY[2800, 2200, 1700, 1600, 1650];
    v_emp_id    uuid;
    v_dow       int;
    v_clock_in  timestamptz;
    v_clock_out timestamptz;
    v_hours     numeric;
    i int;
BEGIN
    FOR v_day IN SELECT generate_series(CURRENT_DATE - 30, CURRENT_DATE - 1, '1 day')::date LOOP
        v_dow := EXTRACT(DOW FROM v_day)::int;
        FOR i IN 1..5 LOOP
            v_emp_id := v_emps[i];
            -- Everyone works Mon-Fri; staff 3,4,5 also work some weekends
            IF v_dow BETWEEN 1 AND 5 THEN
                -- Weekday shift
                IF i <= 2 THEN
                    -- GM and shift manager: morning shift
                    v_clock_in := v_day + interval '8 hours' + (random() * interval '15 minutes');
                    -- Some overtime days for manager
                    IF i = 2 AND v_dow IN (4,5) THEN
                        v_clock_out := v_clock_in + interval '9 hours' + (random() * interval '1 hour');
                    ELSE
                        v_clock_out := v_clock_in + interval '7 hours' + (random() * interval '1 hour');
                    END IF;
                ELSE
                    -- Staff: afternoon/evening shift
                    v_clock_in := v_day + interval '14 hours' + (random() * interval '30 minutes');
                    v_clock_out := v_clock_in + interval '6 hours' + (random() * interval '2 hours');
                END IF;

                INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status, created_at, updated_at)
                VALUES (v_org_id, v_loc_id, v_emp_id, v_roles[i], v_clock_in, v_clock_out, v_rates[i], 'completed', v_clock_in, v_clock_out);

            ELSIF v_dow IN (0, 6) AND i >= 3 THEN
                -- Weekend shifts for staff only
                IF random() > 0.3 THEN
                    v_clock_in := v_day + interval '10 hours' + (random() * interval '30 minutes');
                    v_clock_out := v_clock_in + interval '7 hours' + (random() * interval '2 hours');

                    INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status, created_at, updated_at)
                    VALUES (v_org_id, v_loc_id, v_emp_id, v_roles[i], v_clock_in, v_clock_out, v_rates[i], 'completed', v_clock_in, v_clock_out);
                END IF;
            END IF;
        END LOOP;
    END LOOP;
    RAISE NOTICE 'Shifts seeded';
END $$;

-- ============================================================================
-- 8. STAFF POINT EVENTS (35 events over past 30 days)
-- ============================================================================
INSERT INTO staff_point_events (org_id, employee_id, points, reason, description, created_at) VALUES
-- Maria Santos (GM) - mostly bonuses
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 15.00, 'task_completion', 'Completed weekly inventory count ahead of schedule', CURRENT_DATE - 28),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 10.00, 'accuracy_bonus', 'Zero variance on cash drawer', CURRENT_DATE - 20),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 20.00, 'attendance', 'Perfect attendance - March week 1', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 10.00, 'task_completion', 'Staff training session completed', CURRENT_DATE - 7),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 15.00, 'peer_nominated', 'Nominated by team for leadership', CURRENT_DATE - 3),
-- Jake Thompson (Shift Manager) - good but one late
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 10.00, 'speed_bonus', 'Fastest table turn during Friday rush', CURRENT_DATE - 25),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', -5.00, 'late', 'Arrived 12 minutes late', CURRENT_DATE - 22),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 15.00, 'task_completion', 'Trained 2 new staff members', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 20.00, 'attendance', 'Perfect attendance recovery', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 10.00, 'accuracy_bonus', 'Accurate closing report 5 days straight', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 15.00, 'speed_bonus', 'Record ticket times during Saturday rush', CURRENT_DATE - 2),
-- Sarah Chen (Staff) - top performer
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 20.00, 'speed_bonus', 'Fastest prep time this month', CURRENT_DATE - 27),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 15.00, 'task_completion', 'Deep clean of walk-in completed', CURRENT_DATE - 21),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 25.00, 'peer_nominated', 'Team MVP nomination', CURRENT_DATE - 15),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 10.00, 'accuracy_bonus', 'Perfect inventory count', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 20.00, 'attendance', 'Perfect attendance March', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 10.00, 'task_completion', 'Organized dry storage', CURRENT_DATE - 1),
-- Marcus Brown (Staff) - average with a couple penalties
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 10.00, 'task_completion', 'Restocked all stations', CURRENT_DATE - 26),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', -5.00, 'late', '15 minutes late', CURRENT_DATE - 23),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 10.00, 'speed_bonus', 'Fast prep for catering order', CURRENT_DATE - 17),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', -10.00, 'incomplete_task', 'Did not complete closing checklist', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 15.00, 'task_completion', 'Caught up on all tasks', CURRENT_DATE - 6),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 10.00, 'attendance', 'On time all week', CURRENT_DATE - 2),
-- Lily Nguyen (Staff) - solid
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 15.00, 'task_completion', 'Organized prep station', CURRENT_DATE - 29),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 10.00, 'speed_bonus', 'Quick turnaround on expo', CURRENT_DATE - 24),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', -5.00, 'late', '10 min late due to traffic', CURRENT_DATE - 19),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 20.00, 'peer_nominated', 'Best team player', CURRENT_DATE - 13),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 15.00, 'accuracy_bonus', 'Zero waste on prep', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 10.00, 'task_completion', 'Cross-trained on grill', CURRENT_DATE - 3),
-- Airport staff
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-1111-1111-1111-111111111111', 20.00, 'attendance', 'Perfect attendance', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-2222-2222-2222-222222222222', 15.00, 'speed_bonus', 'Fast service during delay rush', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-3333-3333-3333-333333333333', 10.00, 'task_completion', 'Cleaned entire FOH', CURRENT_DATE - 7),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-1111-1111-1111-111111111111', 10.00, 'manager_adjustment', 'Bonus for handling staffing emergency', CURRENT_DATE - 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-2222-2222-2222-222222222222', -5.00, 'late', '8 minutes late', CURRENT_DATE - 2);

-- Update staff_points totals on employees
UPDATE employees SET staff_points = COALESCE((
    SELECT SUM(points) FROM staff_point_events WHERE staff_point_events.employee_id = employees.employee_id
), 0);

-- ============================================================================
-- 9. GUEST PROFILES (20 guests)
-- ============================================================================
INSERT INTO guest_profiles (guest_id, org_id, payment_token_hash, privacy_tier, first_name, email, phone, total_visits, total_spend, avg_check, preferred_channel, favorite_items, clv_score, segment, churn_risk, churn_probability, next_visit_predicted, last_visit_at) VALUES
-- Champions (3)
('aa111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_champ_001', 3, 'James Mitchell', 'james.m@email.com', '512-555-0101', 28, 84500, 3018, 'dine_in', '["Classic Burger","Loaded Fries"]', 1250.00, 'champion', 'low', 0.0500, CURRENT_DATE + 3, CURRENT_DATE - 2),
('aa111111-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_champ_002', 3, 'Elena Rodriguez', 'elena.r@email.com', '512-555-0102', 22, 72600, 3300, 'dine_in', '["Bacon Avocado Burger","Caesar Salad"]', 1180.00, 'champion', 'low', 0.0800, CURRENT_DATE + 5, CURRENT_DATE - 4),
('aa111111-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_champ_003', 3, 'Robert Chen', 'robert.c@email.com', '512-555-0103', 30, 95000, 3167, 'delivery', '["Grilled Chicken Sandwich","Side Salad"]', 1400.00, 'champion', 'low', 0.0300, CURRENT_DATE + 2, CURRENT_DATE - 1),
-- Loyal Regulars (5)
('aa222222-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_loyal_001', 2, 'Sarah Thompson', 'sarah.t@email.com', '512-555-0201', 15, 42000, 2800, 'dine_in', '["Classic Burger"]', 680.00, 'loyal_regular', 'low', 0.1200, CURRENT_DATE + 7, CURRENT_DATE - 5),
('aa222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_loyal_002', 2, 'David Kim', NULL, '512-555-0202', 12, 38400, 3200, 'takeout', '["Bacon Avocado Burger","Loaded Fries"]', 590.00, 'loyal_regular', 'low', 0.1500, CURRENT_DATE + 6, CURRENT_DATE - 6),
('aa222222-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_loyal_003', 2, 'Lisa Nguyen', 'lisa.n@email.com', '512-555-0203', 14, 35000, 2500, 'dine_in', '["Caesar Salad","Side Salad"]', 620.00, 'loyal_regular', 'medium', 0.2000, CURRENT_DATE + 10, CURRENT_DATE - 9),
('aa222222-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_loyal_004', 2, 'Michael Park', 'mike.p@email.com', NULL, 10, 28500, 2850, 'delivery', '["Grilled Chicken Sandwich"]', 520.00, 'loyal_regular', 'low', 0.1000, CURRENT_DATE + 4, CURRENT_DATE - 3),
('aa222222-5555-5555-5555-555555555555', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_loyal_005', 2, 'Amanda Jones', 'amanda.j@email.com', '512-555-0205', 11, 31000, 2818, 'dine_in', '["Classic Burger","Caesar Salad"]', 545.00, 'loyal_regular', 'low', 0.1100, CURRENT_DATE + 8, CURRENT_DATE - 7),
-- At Risk (4) - high CLV but lapsed
('aa333333-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_risk_001', 2, 'Chris Williams', 'chris.w@email.com', '512-555-0301', 18, 58000, 3222, 'dine_in', '["Bacon Avocado Burger"]', 950.00, 'at_risk', 'high', 0.6500, CURRENT_DATE - 5, CURRENT_DATE - 35),
('aa333333-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_risk_002', 2, 'Jennifer Lee', 'jen.lee@email.com', '512-555-0302', 14, 46200, 3300, 'dine_in', '["Classic Burger","Loaded Fries"]', 780.00, 'at_risk', 'high', 0.7000, CURRENT_DATE - 10, CURRENT_DATE - 42),
('aa333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_risk_003', 1, 'Mark Davis', NULL, '512-555-0303', 16, 51200, 3200, 'takeout', '["Grilled Chicken Sandwich","Caesar Salad"]', 850.00, 'at_risk', 'critical', 0.8200, CURRENT_DATE - 15, CURRENT_DATE - 55),
('aa333333-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_risk_004', 2, 'Rachel Green', 'rachel.g@email.com', '512-555-0304', 12, 39600, 3300, 'delivery', '["Loaded Fries"]', 710.00, 'at_risk', 'medium', 0.4500, CURRENT_DATE + 2, CURRENT_DATE - 28),
-- New Discoverers (3)
('aa444444-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_new_001', 1, 'Tyler Brown', NULL, NULL, 2, 5200, 2600, 'dine_in', '["Classic Burger"]', 85.00, 'new_discoverer', 'medium', 0.3500, CURRENT_DATE + 12, CURRENT_DATE - 8),
('aa444444-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_new_002', 1, 'Mia Johnson', NULL, NULL, 1, 2800, 2800, 'takeout', '["Bacon Avocado Burger"]', 45.00, 'new_discoverer', 'medium', 0.4000, CURRENT_DATE + 20, CURRENT_DATE - 12),
('aa444444-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_new_003', 1, 'Alex Rivera', NULL, NULL, 3, 7500, 2500, 'delivery', '["Loaded Fries","Side Salad"]', 120.00, 'new_discoverer', 'low', 0.2500, CURRENT_DATE + 6, CURRENT_DATE - 5),
-- Casual (5)
('aa555555-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_cas_001', 1, NULL, NULL, NULL, 5, 12500, 2500, 'dine_in', '["Classic Burger"]', 200.00, 'casual', 'medium', 0.3000, CURRENT_DATE + 15, CURRENT_DATE - 18),
('aa555555-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_cas_002', 1, NULL, NULL, NULL, 4, 9800, 2450, 'takeout', '["Grilled Chicken Sandwich"]', 160.00, 'casual', 'medium', 0.3500, CURRENT_DATE + 20, CURRENT_DATE - 22),
('aa555555-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_cas_003', 1, NULL, NULL, NULL, 6, 16200, 2700, 'dine_in', '["Bacon Avocado Burger"]', 250.00, 'casual', 'low', 0.2000, CURRENT_DATE + 10, CURRENT_DATE - 12),
('aa555555-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_cas_004', 1, NULL, NULL, NULL, 3, 7200, 2400, 'delivery', '["Loaded Fries"]', 115.00, 'casual', 'high', 0.5500, CURRENT_DATE + 25, CURRENT_DATE - 30),
('aa555555-5555-5555-5555-555555555555', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_cas_005', 1, NULL, NULL, NULL, 4, 10800, 2700, 'dine_in', '["Caesar Salad"]', 175.00, 'casual', 'medium', 0.3200, CURRENT_DATE + 18, CURRENT_DATE - 15);

-- ============================================================================
-- 10. GUEST VISITS (linked to existing checks where possible)
-- ============================================================================
DO $$
DECLARE
    v_org_id     uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_downtown   uuid := 'a1111111-1111-1111-1111-111111111111';
    v_guest_ids  uuid[] := ARRAY[
        'aa111111-1111-1111-1111-111111111111'::uuid,
        'aa111111-2222-2222-2222-222222222222'::uuid,
        'aa111111-3333-3333-3333-333333333333'::uuid,
        'aa222222-1111-1111-1111-111111111111'::uuid,
        'aa222222-2222-2222-2222-222222222222'::uuid,
        'aa222222-3333-3333-3333-333333333333'::uuid,
        'aa222222-4444-4444-4444-444444444444'::uuid,
        'aa222222-5555-5555-5555-555555555555'::uuid,
        'aa333333-1111-1111-1111-111111111111'::uuid,
        'aa333333-2222-2222-2222-222222222222'::uuid,
        'aa333333-3333-3333-3333-333333333333'::uuid,
        'aa333333-4444-4444-4444-444444444444'::uuid,
        'aa444444-1111-1111-1111-111111111111'::uuid,
        'aa444444-2222-2222-2222-222222222222'::uuid,
        'aa444444-3333-3333-3333-333333333333'::uuid,
        'aa555555-1111-1111-1111-111111111111'::uuid,
        'aa555555-2222-2222-2222-222222222222'::uuid,
        'aa555555-3333-3333-3333-333333333333'::uuid,
        'aa555555-4444-4444-4444-444444444444'::uuid,
        'aa555555-5555-5555-5555-555555555555'::uuid
    ];
    v_visit_counts int[] := ARRAY[15,12,15,8,7,8,6,6,4,3,3,3,2,1,3,5,4,6,3,4];
    v_channels     text[] := ARRAY['dine_in','dine_in','delivery','dine_in','takeout','dine_in','delivery','dine_in','dine_in','dine_in','takeout','delivery','dine_in','takeout','delivery','dine_in','takeout','dine_in','delivery','dine_in'];
    v_spends       int[] := ARRAY[3018,3300,3167,2800,3200,2500,2850,2818,3222,3300,3200,3300,2600,2800,2500,2500,2450,2700,2400,2700];
    v_guest_id     uuid;
    v_num_visits   int;
    v_visit_date   timestamptz;
    i int; j int;
BEGIN
    FOR i IN 1..20 LOOP
        v_guest_id := v_guest_ids[i];
        v_num_visits := v_visit_counts[i];
        FOR j IN 1..v_num_visits LOOP
            -- Spread visits over the past 90 days
            v_visit_date := CURRENT_DATE - (random() * 90)::int + interval '12 hours' + (random() * interval '8 hours');
            INSERT INTO guest_visits (org_id, guest_id, location_id, channel, spend, item_count, party_size, visited_at)
            VALUES (v_org_id, v_guest_id, v_downtown, v_channels[i],
                (v_spends[i] * (0.8 + random()*0.4))::int,
                1 + floor(random()*3)::int,
                CASE WHEN v_channels[i] = 'dine_in' THEN 1 + floor(random()*4)::int ELSE 1 END,
                v_visit_date);
        END LOOP;
    END LOOP;
    RAISE NOTICE 'Guest visits seeded';
END $$;

-- ============================================================================
-- 11. KITCHEN STATIONS (Downtown)
-- ============================================================================
INSERT INTO kitchen_stations (station_id, org_id, location_id, name, station_type, max_concurrent, status) VALUES
('ad111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Grill Station', 'grill', 4, 'active'),
('ad222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Fryer Station', 'fryer', 3, 'active'),
('ad333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Saute Station', 'saute', 3, 'active'),
('ad444444-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Prep Station', 'prep', 4, 'active'),
('ad555555-5555-5555-5555-555555555555', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Expo Station', 'expo', 2, 'active'),
('ad666666-6666-6666-6666-666666666666', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Dish Pit', 'dish', 2, 'active');

-- ============================================================================
-- 12. MENU ITEM RESOURCE PROFILES (Downtown items mapped to stations)
-- ============================================================================
INSERT INTO menu_item_resource_profiles (org_id, menu_item_id, station_type, task_sequence, duration_secs, elu_required, batch_size) VALUES
-- Classic Burger: prep -> grill -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-1111-aaaa-111111111111', 'prep', 1, 60, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-1111-aaaa-111111111111', 'grill', 2, 300, 1.00, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-1111-aaaa-111111111111', 'expo', 3, 45, 0.50, 1),
-- Bacon Avocado Burger: prep -> grill -> saute(bacon) -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-2222-aaaa-111111111111', 'prep', 1, 90, 0.75, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-2222-aaaa-111111111111', 'grill', 2, 300, 1.00, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-2222-aaaa-111111111111', 'saute', 3, 120, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-2222-aaaa-111111111111', 'expo', 4, 45, 0.50, 1),
-- Grilled Chicken Sandwich: prep -> grill -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-3333-aaaa-111111111111', 'prep', 1, 60, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-3333-aaaa-111111111111', 'grill', 2, 360, 1.00, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-3333-aaaa-111111111111', 'expo', 3, 45, 0.50, 1),
-- Caesar Salad: prep -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-4444-aaaa-111111111111', 'prep', 1, 120, 1.00, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-4444-aaaa-111111111111', 'expo', 2, 30, 0.25, 1),
-- Loaded Fries: fryer -> saute(toppings) -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-5555-aaaa-111111111111', 'fryer', 1, 240, 1.00, 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-5555-aaaa-111111111111', 'saute', 2, 90, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-5555-aaaa-111111111111', 'expo', 3, 30, 0.25, 1),
-- Side Salad: prep -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-6666-aaaa-111111111111', 'prep', 1, 60, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '22222222-aaaa-6666-aaaa-111111111111', 'expo', 2, 20, 0.25, 1);

-- ============================================================================
-- 13. MENU ITEM SCORES (pre-calculated 5D scores)
-- ============================================================================
UPDATE menu_items SET
    margin_score = 75, velocity_score = 85, complexity_score = 70,
    satisfaction_score = 80, strategic_score = 60, classification = 'powerhouse',
    classification_changed_at = CURRENT_DATE - 7, updated_at = NOW()
WHERE menu_item_id = '22222222-aaaa-1111-aaaa-111111111111';

UPDATE menu_items SET
    margin_score = 82, velocity_score = 55, complexity_score = 45,
    satisfaction_score = 75, strategic_score = 50, classification = 'hidden_gem',
    classification_changed_at = CURRENT_DATE - 7, updated_at = NOW()
WHERE menu_item_id = '22222222-aaaa-2222-aaaa-111111111111';

UPDATE menu_items SET
    margin_score = 60, velocity_score = 70, complexity_score = 75,
    satisfaction_score = 70, strategic_score = 50, classification = 'workhorse',
    classification_changed_at = CURRENT_DATE - 7, updated_at = NOW()
WHERE menu_item_id = '22222222-aaaa-3333-aaaa-111111111111';

UPDATE menu_items SET
    margin_score = 85, velocity_score = 40, complexity_score = 90,
    satisfaction_score = 65, strategic_score = 40, classification = 'hidden_gem',
    classification_changed_at = CURRENT_DATE - 7, updated_at = NOW()
WHERE menu_item_id = '22222222-aaaa-4444-aaaa-111111111111';

UPDATE menu_items SET
    margin_score = 45, velocity_score = 80, complexity_score = 85,
    satisfaction_score = 60, strategic_score = 30, classification = 'crowd_pleaser',
    classification_changed_at = CURRENT_DATE - 7, updated_at = NOW()
WHERE menu_item_id = '22222222-aaaa-5555-aaaa-111111111111';

UPDATE menu_items SET
    margin_score = 70, velocity_score = 25, complexity_score = 95,
    satisfaction_score = 50, strategic_score = 20, classification = 'declining_star',
    classification_changed_at = CURRENT_DATE - 7, updated_at = NOW()
WHERE menu_item_id = '22222222-aaaa-6666-aaaa-111111111111';

-- ============================================================================
-- 14. VENDOR SCORES
-- ============================================================================
INSERT INTO vendor_scores (org_id, location_id, vendor_name, overall_score, price_score, delivery_score, quality_score, accuracy_score, total_orders, otif_rate, on_time_rate, in_full_rate, avg_lead_days, calculated_at) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'US Foods', 78.50, 72.00, 85.00, 80.00, 77.00, 24, 82.50, 88.00, 85.00, 2.50, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 82.00, 68.00, 90.00, 85.00, 85.00, 18, 88.00, 92.00, 90.00, 2.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Local Farm Co', 88.00, 75.00, 82.00, 95.00, 90.00, 12, 90.00, 85.00, 95.00, 1.00, NOW());

-- ============================================================================
-- 15. INGREDIENT PRICE HISTORY (past 6 months)
-- ============================================================================
INSERT INTO ingredient_price_history (org_id, ingredient_id, vendor_name, unit_cost, quantity, source, recorded_at) VALUES
-- Ground Beef trending up
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-1111-aaaa-111111111111', 'US Foods', 420, 60.0, 'po_received', CURRENT_DATE - 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-1111-aaaa-111111111111', 'US Foods', 430, 55.0, 'po_received', CURRENT_DATE - 150),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-1111-aaaa-111111111111', 'US Foods', 440, 60.0, 'po_received', CURRENT_DATE - 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-1111-aaaa-111111111111', 'US Foods', 445, 58.0, 'po_received', CURRENT_DATE - 90),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-1111-aaaa-111111111111', 'US Foods', 450, 60.0, 'po_received', CURRENT_DATE - 60),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-1111-aaaa-111111111111', 'US Foods', 460, 58.0, 'po_received', CURRENT_DATE - 16),
-- Chicken stable
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-2222-aaaa-111111111111', 'US Foods', 370, 40.0, 'po_received', CURRENT_DATE - 150),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-2222-aaaa-111111111111', 'US Foods', 375, 40.0, 'po_received', CURRENT_DATE - 90),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-2222-aaaa-111111111111', 'US Foods', 385, 40.0, 'po_received', CURRENT_DATE - 16),
-- Avocado seasonal spike
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-8888-aaaa-111111111111', 'Local Farm Co', 150, 50.0, 'po_received', CURRENT_DATE - 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-8888-aaaa-111111111111', 'Local Farm Co', 160, 45.0, 'po_received', CURRENT_DATE - 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-8888-aaaa-111111111111', 'Local Farm Co', 175, 40.0, 'po_received', CURRENT_DATE - 60),
-- Bacon up
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-9999-aaaa-111111111111', 'US Foods', 650, 25.0, 'po_received', CURRENT_DATE - 150),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-9999-aaaa-111111111111', 'US Foods', 680, 25.0, 'po_received', CURRENT_DATE - 60),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11111111-aaaa-9999-aaaa-111111111111', 'US Foods', 695, 24.0, 'po_received', CURRENT_DATE - 16);

-- ============================================================================
-- 16. CAMPAIGNS
-- ============================================================================
INSERT INTO campaigns (campaign_id, org_id, location_id, name, campaign_type, status, target_segment, channel, discount_type, discount_value, min_purchase, start_at, end_at, recurring, recurrence_rule, redemptions, revenue_attributed, cost_of_promotion) VALUES
('ca111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Happy Hour Special', 'happy_hour', 'active', 'casual', 'all', 'percentage', 15.00, 2000, CURRENT_DATE - 14, CURRENT_DATE + 14, true, 'FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR', 45, 225000, 33750),
('ca222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Weekend BOGO', 'bogo', 'completed', 'loyal_regular', 'email', 'bogo', NULL, 1500, CURRENT_DATE - 45, CURRENT_DATE - 15, false, NULL, 120, 840000, 126000),
('ca333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Loyalty Launch', 'loyalty_reward', 'draft', 'champion', 'push', 'dollar_off', 5.00, 2500, NULL, NULL, false, NULL, 0, 0, 0);

-- ============================================================================
-- 17. LOYALTY MEMBERS (linked to guest profiles)
-- ============================================================================
INSERT INTO loyalty_members (member_id, org_id, guest_id, points_balance, lifetime_points, tier, joined_at) VALUES
('ae111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa111111-1111-1111-1111-111111111111', 845.00, 2800.00, 'platinum', CURRENT_DATE - 180),
('ae222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa111111-2222-2222-2222-222222222222', 620.00, 2400.00, 'gold', CURRENT_DATE - 150),
('ae333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa111111-3333-3333-3333-333333333333', 1100.00, 3200.00, 'platinum', CURRENT_DATE - 200),
('ae444444-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa222222-1111-1111-1111-111111111111', 320.00, 1400.00, 'gold', CURRENT_DATE - 120),
('ae555555-5555-5555-5555-555555555555', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa222222-2222-2222-2222-222222222222', 180.00, 1100.00, 'silver', CURRENT_DATE - 100),
('ae666666-6666-6666-6666-666666666666', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa222222-3333-3333-3333-333333333333', 450.00, 1200.00, 'gold', CURRENT_DATE - 90),
('ae777777-7777-7777-7777-777777777777', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa333333-1111-1111-1111-111111111111', 200.00, 1800.00, 'gold', CURRENT_DATE - 160),
('ae888888-8888-8888-8888-888888888888', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa444444-3333-3333-3333-333333333333', 75.00, 75.00, 'bronze', CURRENT_DATE - 30);

-- ============================================================================
-- 18. LOYALTY TRANSACTIONS (earn/redeem)
-- ============================================================================
INSERT INTO loyalty_transactions (org_id, member_id, type, points, description, created_at) VALUES
-- Platinum member James - lots of earns
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae111111-1111-1111-1111-111111111111', 'earn', 150.00, 'Purchase - Classic Burger combo', CURRENT_DATE - 28),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae111111-1111-1111-1111-111111111111', 'earn', 180.00, 'Purchase - dinner for two', CURRENT_DATE - 21),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae111111-1111-1111-1111-111111111111', 'redeem', -500.00, 'Redeemed $5 reward', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae111111-1111-1111-1111-111111111111', 'earn', 130.00, 'Purchase - lunch', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae111111-1111-1111-1111-111111111111', 'earn', 165.00, 'Purchase - takeout order', CURRENT_DATE - 7),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae111111-1111-1111-1111-111111111111', 'earn', 120.00, 'Purchase - quick lunch', CURRENT_DATE - 2),
-- Gold member Elena
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae222222-2222-2222-2222-222222222222', 'earn', 170.00, 'Purchase - Bacon Avocado combo', CURRENT_DATE - 25),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae222222-2222-2222-2222-222222222222', 'earn', 140.00, 'Purchase - salad lunch', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae222222-2222-2222-2222-222222222222', 'redeem', -300.00, 'Redeemed $3 reward', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae222222-2222-2222-2222-222222222222', 'earn', 160.00, 'Purchase - dinner', CURRENT_DATE - 4),
-- Platinum Robert
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae333333-3333-3333-3333-333333333333', 'earn', 200.00, 'Purchase - family delivery', CURRENT_DATE - 22),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae333333-3333-3333-3333-333333333333', 'earn', 180.00, 'Purchase - delivery order', CURRENT_DATE - 15),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae333333-3333-3333-3333-333333333333', 'redeem', -1000.00, 'Redeemed $10 reward', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae333333-3333-3333-3333-333333333333', 'earn', 220.00, 'Purchase - large delivery', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae333333-3333-3333-3333-333333333333', 'earn', 190.00, 'Purchase - delivery', CURRENT_DATE - 1),
-- Other members
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae444444-4444-4444-4444-444444444444', 'earn', 140.00, 'Purchase', CURRENT_DATE - 20),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae444444-4444-4444-4444-444444444444', 'earn', 130.00, 'Purchase', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae555555-5555-5555-5555-555555555555', 'earn', 160.00, 'Purchase', CURRENT_DATE - 15),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae555555-5555-5555-5555-555555555555', 'redeem', -200.00, 'Redeemed $2 reward', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae666666-6666-6666-6666-666666666666', 'earn', 125.00, 'Purchase', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae666666-6666-6666-6666-666666666666', 'earn', 150.00, 'Purchase', CURRENT_DATE - 9),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae777777-7777-7777-7777-777777777777', 'earn', 160.00, 'Purchase', CURRENT_DATE - 35),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae777777-7777-7777-7777-777777777777', 'redeem', -500.00, 'Redeemed $5 reward', CURRENT_DATE - 34),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae888888-8888-8888-8888-888888888888', 'earn', 75.00, 'Welcome bonus + first purchase', CURRENT_DATE - 5);

-- ============================================================================
-- 19. PORTFOLIO NODES (Portfolio -> Region -> Locations)
-- ============================================================================
INSERT INTO portfolio_nodes (node_id, org_id, parent_node_id, name, node_type, location_id, sort_order) VALUES
('af111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', NULL, 'Bistro Cloud Portfolio', 'org', NULL, 0),
('af222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af111111-1111-1111-1111-111111111111', 'Austin Metro', 'region', NULL, 1),
('af333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af222222-2222-2222-2222-222222222222', 'Downtown', 'location', 'a1111111-1111-1111-1111-111111111111', 1),
('af444444-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af222222-2222-2222-2222-222222222222', 'Airport', 'location', 'b2222222-2222-2222-2222-222222222222', 2);

-- ============================================================================
-- 20. LOCATION BENCHMARKS (current month)
-- ============================================================================
INSERT INTO location_benchmarks (org_id, location_id, period_start, period_end, revenue, food_cost_pct, labor_cost_pct, avg_check_cents, check_count, revenue_percentile, food_cost_percentile, labor_cost_percentile, avg_check_percentile) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '2026-03-01', '2026-03-31', 3250000, 28.500, 27.200, 2850, 1140, 72.000, 68.000, 65.000, 70.000),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '2026-03-01', '2026-03-31', 1850000, 31.200, 29.800, 3150, 587, 58.000, 45.000, 50.000, 78.000);

-- ============================================================================
-- 21. BEST PRACTICES
-- ============================================================================
INSERT INTO best_practices (org_id, title, description, metric, source_location_id, impact_pct, status, detected_at) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Optimized Prep Scheduling Reduces Food Waste', 'Downtown location reduced food waste by 18% after implementing demand-based prep scheduling. Morning prep quantities are now adjusted based on day-of-week sales patterns and weather forecasts.', 'food_waste_pct', 'a1111111-1111-1111-1111-111111111111', 18.000, 'suggested', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cross-Training Program Improves Labor Efficiency', 'Cross-training staff on multiple stations at Downtown reduced overtime hours by 12% while maintaining ticket times. Staff certified on 2+ stations can flex between positions during demand spikes.', 'labor_cost_pct', 'a1111111-1111-1111-1111-111111111111', 12.000, 'adopted', CURRENT_DATE - 10);

-- ============================================================================
-- 22. SCHEDULED SHIFTS (this week at Downtown)
-- ============================================================================
INSERT INTO schedules (schedule_id, org_id, location_id, week_start, status, created_by, published_at) VALUES
('5a111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', date_trunc('week', CURRENT_DATE)::date, 'published', '0d55e810-1e4a-417a-8a70-08b98f4595c2', NOW() - interval '2 days');

DO $$
DECLARE
    v_org_id  uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_sched   uuid := '5a111111-1111-1111-1111-111111111111';
    v_week_start date := date_trunc('week', CURRENT_DATE)::date;
    v_emps    uuid[] := ARRAY[
        'ee111111-1111-1111-1111-111111111111'::uuid,
        'ee111111-2222-2222-2222-222222222222'::uuid,
        'ee111111-3333-3333-3333-333333333333'::uuid,
        'ee111111-4444-4444-4444-444444444444'::uuid,
        'ee111111-5555-5555-5555-555555555555'::uuid
    ];
    v_stations text[] := ARRAY['expo', 'grill', 'prep', 'fryer', 'saute'];
    v_day date;
    v_dow int;
    i int;
BEGIN
    -- Mon-Fri for all, Sat for staff
    FOR i IN 0..6 LOOP
        v_day := v_week_start + i;
        v_dow := EXTRACT(DOW FROM v_day)::int;

        -- GM: Mon-Fri 8am-4pm
        IF v_dow BETWEEN 1 AND 5 THEN
            INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
            VALUES (v_org_id, v_sched, v_emps[1], v_day, '08:00', '16:00', 'expo', 'confirmed');
        END IF;

        -- Shift Manager: Mon-Fri 10am-7pm
        IF v_dow BETWEEN 1 AND 5 THEN
            INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
            VALUES (v_org_id, v_sched, v_emps[2], v_day, '10:00', '19:00', 'grill', 'confirmed');
        END IF;

        -- Staff 3: Mon-Sat 2pm-10pm
        IF v_dow BETWEEN 1 AND 6 THEN
            INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
            VALUES (v_org_id, v_sched, v_emps[3], v_day, '14:00', '22:00', 'prep', 'scheduled');
        END IF;

        -- Staff 4: Mon-Fri 3pm-10pm, Sat 11am-7pm
        IF v_dow BETWEEN 1 AND 5 THEN
            INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
            VALUES (v_org_id, v_sched, v_emps[4], v_day, '15:00', '22:00', 'fryer', 'scheduled');
        ELSIF v_dow = 6 THEN
            INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
            VALUES (v_org_id, v_sched, v_emps[4], v_day, '11:00', '19:00', 'fryer', 'scheduled');
        END IF;

        -- Staff 5: Tue-Sat 4pm-11pm
        IF v_dow BETWEEN 2 AND 6 THEN
            INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
            VALUES (v_org_id, v_sched, v_emps[5], v_day, '16:00', '23:00', 'saute', 'scheduled');
        END IF;
    END LOOP;
    RAISE NOTICE 'Scheduled shifts seeded';
END $$;

-- ============================================================================
-- 23. LABOR DEMAND FORECAST (today and tomorrow)
-- ============================================================================
INSERT INTO labor_demand_forecast (org_id, location_id, forecast_date, time_block, forecasted_covers, required_elu, required_headcount) VALUES
-- Today
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '08:00', 5, 2.00, 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '09:00', 8, 3.00, 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '10:00', 12, 4.50, 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '11:00', 25, 8.50, 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '12:00', 40, 14.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '13:00', 35, 12.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '14:00', 18, 6.50, 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '15:00', 10, 4.00, 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '16:00', 12, 4.50, 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '17:00', 20, 7.00, 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '18:00', 38, 13.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '19:00', 42, 15.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '20:00', 30, 10.50, 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '21:00', 15, 5.50, 3),
-- Tomorrow
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '08:00', 4, 1.50, 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '09:00', 7, 2.50, 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '10:00', 10, 3.50, 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '11:00', 22, 7.50, 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '12:00', 38, 13.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '13:00', 32, 11.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '14:00', 15, 5.50, 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '15:00', 8, 3.00, 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '16:00', 10, 3.50, 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '17:00', 18, 6.50, 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '18:00', 35, 12.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '19:00', 40, 14.00, 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '20:00', 28, 9.50, 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '21:00', 12, 4.50, 3);

-- ============================================================================
-- 24. KDS TICKETS (5 active tickets with items)
-- ============================================================================
-- First create 5 open checks for the KDS tickets
INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel, subtotal, tax, total, tip, discount, opened_at, source) VALUES
('ab111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'KDS-001', 'open', 'dine_in', 2890, 238, 3128, 0, 0, NOW() - interval '8 minutes', 'manual'),
('ab222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'KDS-002', 'open', 'dine_in', 4085, 337, 4422, 0, 0, NOW() - interval '5 minutes', 'manual'),
('ab333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'KDS-003', 'open', 'takeout', 1495, 123, 1618, 0, 0, NOW() - interval '3 minutes', 'manual'),
('ab444444-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'KDS-004', 'open', 'delivery', 3585, 296, 3881, 0, 0, NOW() - interval '12 minutes', 'manual'),
('ab555555-5555-5555-5555-555555555555', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'KDS-005', 'open', 'dine_in', 2390, 197, 2587, 0, 0, NOW() - interval '1 minute', 'manual');

-- KDS Tickets
INSERT INTO kds_tickets (ticket_id, org_id, location_id, check_id, order_number, channel, status, priority, estimated_ready_at, created_at, updated_at) VALUES
('ac111111-1111-1111-1111-111111111111', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ab111111-1111-1111-1111-111111111111', 'KDS-001', 'dine_in', 'in_progress', 0, NOW() + interval '8 minutes', NOW() - interval '8 minutes', NOW()),
('ac222222-2222-2222-2222-222222222222', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ab222222-2222-2222-2222-222222222222', 'KDS-002', 'dine_in', 'in_progress', 0, NOW() + interval '12 minutes', NOW() - interval '5 minutes', NOW()),
('ac333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ab333333-3333-3333-3333-333333333333', 'KDS-003', 'takeout', 'new', 0, NOW() + interval '15 minutes', NOW() - interval '3 minutes', NOW()),
('ac444444-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ab444444-4444-4444-4444-444444444444', 'KDS-004', 'delivery', 'in_progress', 1, NOW() + interval '5 minutes', NOW() - interval '12 minutes', NOW()),
('ac555555-5555-5555-5555-555555555555', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ab555555-5555-5555-5555-555555555555', 'KDS-005', 'dine_in', 'new', 0, NOW() + interval '18 minutes', NOW() - interval '1 minute', NOW());

-- KDS Ticket Items
INSERT INTO kds_ticket_items (org_id, ticket_id, menu_item_id, item_name, quantity, station_type, status, fire_at, started_at, completed_at, duration_secs) VALUES
-- Ticket 1: Classic Burger (grill cooking) + Loaded Fries (ready)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac111111-1111-1111-1111-111111111111', '22222222-aaaa-1111-aaaa-111111111111', 'Classic Burger', 1, 'grill', 'cooking', NOW() - interval '8 minutes', NOW() - interval '5 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac111111-1111-1111-1111-111111111111', '22222222-aaaa-5555-aaaa-111111111111', 'Loaded Fries', 1, 'fryer', 'ready', NOW() - interval '8 minutes', NOW() - interval '6 minutes', NOW() - interval '2 minutes', 240),
-- Ticket 2: Bacon Avocado Burger (cooking) + Caesar Salad (ready) + Side Salad (ready)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac222222-2222-2222-2222-222222222222', '22222222-aaaa-2222-aaaa-111111111111', 'Bacon Avocado Burger', 1, 'grill', 'cooking', NOW() - interval '5 minutes', NOW() - interval '3 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac222222-2222-2222-2222-222222222222', '22222222-aaaa-4444-aaaa-111111111111', 'Caesar Salad', 1, 'prep', 'ready', NOW() - interval '5 minutes', NOW() - interval '4 minutes', NOW() - interval '2 minutes', 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac222222-2222-2222-2222-222222222222', '22222222-aaaa-6666-aaaa-111111111111', 'Side Salad', 1, 'prep', 'ready', NOW() - interval '5 minutes', NOW() - interval '4 minutes', NOW() - interval '3 minutes', 60),
-- Ticket 3: Classic Burger (pending/new)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac333333-3333-3333-3333-333333333333', '22222222-aaaa-1111-aaaa-111111111111', 'Classic Burger', 1, 'grill', 'pending', NULL, NULL, NULL, NULL),
-- Ticket 4: Grilled Chicken Sandwich (cooking - almost done) + Loaded Fries (cooking)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac444444-4444-4444-4444-444444444444', '22222222-aaaa-3333-aaaa-111111111111', 'Grilled Chicken Sandwich', 1, 'grill', 'cooking', NOW() - interval '12 minutes', NOW() - interval '10 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac444444-4444-4444-4444-444444444444', '22222222-aaaa-5555-aaaa-111111111111', 'Loaded Fries', 1, 'fryer', 'cooking', NOW() - interval '12 minutes', NOW() - interval '4 minutes', NULL, NULL),
-- Ticket 5: Caesar Salad + Classic Burger (both pending/new)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac555555-5555-5555-5555-555555555555', '22222222-aaaa-4444-aaaa-111111111111', 'Caesar Salad', 1, 'prep', 'pending', NULL, NULL, NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac555555-5555-5555-5555-555555555555', '22222222-aaaa-1111-aaaa-111111111111', 'Classic Burger', 1, 'grill', 'pending', NULL, NULL, NULL, NULL);

COMMIT;

-- Final summary
SELECT 'checks' AS table_name, COUNT(*) AS row_count FROM checks WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'check_items', COUNT(*) FROM check_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'payments', COUNT(*) FROM payments WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'inventory_counts', COUNT(*) FROM inventory_counts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'inventory_count_lines', COUNT(*) FROM inventory_count_lines WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'waste_logs', COUNT(*) FROM waste_logs WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'purchase_orders', COUNT(*) FROM purchase_orders WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'purchase_order_lines', COUNT(*) FROM purchase_order_lines WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'inventory_variances', COUNT(*) FROM inventory_variances WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'budgets', COUNT(*) FROM budgets WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'shifts', COUNT(*) FROM shifts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'staff_point_events', COUNT(*) FROM staff_point_events WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'guest_profiles', COUNT(*) FROM guest_profiles WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'guest_visits', COUNT(*) FROM guest_visits WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'kitchen_stations', COUNT(*) FROM kitchen_stations WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'menu_item_resource_profiles', COUNT(*) FROM menu_item_resource_profiles WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'vendor_scores', COUNT(*) FROM vendor_scores WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'ingredient_price_history', COUNT(*) FROM ingredient_price_history WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'campaigns', COUNT(*) FROM campaigns WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'loyalty_members', COUNT(*) FROM loyalty_members WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'loyalty_transactions', COUNT(*) FROM loyalty_transactions WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'portfolio_nodes', COUNT(*) FROM portfolio_nodes WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'location_benchmarks', COUNT(*) FROM location_benchmarks WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'best_practices', COUNT(*) FROM best_practices WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'schedules', COUNT(*) FROM schedules WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'scheduled_shifts', COUNT(*) FROM scheduled_shifts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'labor_demand_forecast', COUNT(*) FROM labor_demand_forecast WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'kds_tickets', COUNT(*) FROM kds_tickets WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'kds_ticket_items', COUNT(*) FROM kds_ticket_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
ORDER BY table_name;
