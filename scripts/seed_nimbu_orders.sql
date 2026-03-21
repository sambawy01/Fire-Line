-- =============================================================================
-- Nimbu International Menu - Order Generation (30 days, 4 locations)
-- Target: ~250 orders/day weekday, ~300 weekend, ~100K EGP/day per branch
-- =============================================================================

SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

DO $$
DECLARE
  v_org_id uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
  v_locations uuid[] := ARRAY[
    'a1111111-1111-1111-1111-111111111111',
    'b2222222-2222-2222-2222-222222222222',
    'c3333333-3333-3333-3333-333333333333',
    'd4444444-4444-4444-4444-444444444444'
  ];
  v_loc uuid;
  v_day date;
  v_day_offset int;
  v_num_orders int;
  v_dow int;
  v_check_id uuid;
  v_order_time timestamptz;
  v_channel text;
  v_channel_roll float;
  v_num_items int;
  v_item_idx int;
  v_menu_item record;
  v_subtotal int;
  v_tax int;
  v_total int;
  v_order_num int;
  v_items_arr uuid[];
  v_item_count int;
  v_method text;
BEGIN
  -- Collect menu items per location into a temp table for fast access
  CREATE TEMP TABLE tmp_menu_items AS
  SELECT menu_item_id, location_id, name, category, price,
         row_number() OVER (PARTITION BY location_id ORDER BY menu_item_id) as rn,
         count(*) OVER (PARTITION BY location_id) as total_items
  FROM menu_items
  WHERE org_id = v_org_id;

  FOREACH v_loc IN ARRAY v_locations LOOP
    v_order_num := 0;

    FOR v_day_offset IN 0..29 LOOP
      v_day := (current_date - v_day_offset)::date;
      v_dow := extract(dow from v_day);  -- 0=Sun, 5=Fri, 6=Sat

      -- Determine number of orders (weekend = Fri/Sat in Egypt)
      IF v_dow IN (5, 6) THEN
        v_num_orders := 280 + floor(random() * 41)::int;  -- 280-320
      ELSE
        v_num_orders := 220 + floor(random() * 31)::int;  -- 220-250
      END IF;

      FOR i IN 1..v_num_orders LOOP
        v_order_num := v_order_num + 1;

        -- Order time: for today use now() - random hours, otherwise random time in day
        IF v_day_offset = 0 THEN
          v_order_time := now() - (random() * interval '18 hours');
        ELSE
          -- Orders between 10:00 and 01:00 next day (15 hours), weighted toward lunch/dinner
          v_order_time := v_day + interval '10 hours' + (random() * interval '15 hours');
        END IF;

        -- Channel mix: 55% dine_in, 25% takeout, 20% delivery
        v_channel_roll := random();
        IF v_channel_roll < 0.55 THEN
          v_channel := 'dine_in';
        ELSIF v_channel_roll < 0.80 THEN
          v_channel := 'takeout';
        ELSE
          v_channel := 'delivery';
        END IF;

        -- Number of items: 1-3, weighted toward 2
        v_num_items := CASE
          WHEN random() < 0.25 THEN 1
          WHEN random() < 0.70 THEN 2
          ELSE 3
        END;

        v_check_id := gen_random_uuid();
        v_subtotal := 0;

        -- Create the check first with zero totals
        INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel, subtotal, tax, total, tip, discount, opened_at, closed_at, source)
        VALUES (v_check_id, v_org_id, v_loc,
                'ORD-' || to_char(v_day, 'YYMMDD') || '-' || lpad(v_order_num::text, 4, '0'),
                'closed', v_channel, 0, 0, 0, 0, 0,
                v_order_time,
                v_order_time + interval '20 minutes' + (random() * interval '40 minutes'),
                'manual');

        -- Add items
        FOR j IN 1..v_num_items LOOP
          -- Pick random item for this location
          SELECT menu_item_id, name, price INTO v_menu_item
          FROM tmp_menu_items
          WHERE location_id = v_loc
            AND rn = (1 + floor(random() * 20))::int;

          IF v_menu_item.menu_item_id IS NOT NULL THEN
            INSERT INTO check_items (org_id, check_id, menu_item_id, name, quantity, unit_price, fired_at)
            VALUES (v_org_id, v_check_id, v_menu_item.menu_item_id, v_menu_item.name, 1, v_menu_item.price, v_order_time);

            v_subtotal := v_subtotal + v_menu_item.price;
          END IF;
        END LOOP;

        -- Update check totals (14% tax)
        v_tax := (v_subtotal * 0.14)::int;
        v_total := v_subtotal + v_tax;

        UPDATE checks SET subtotal = v_subtotal, tax = v_tax, total = v_total
        WHERE check_id = v_check_id;

        -- Payment
        v_method := CASE WHEN random() < 0.65 THEN 'card' ELSE 'cash' END;
        INSERT INTO payments (org_id, check_id, amount, tip, method, status)
        VALUES (v_org_id, v_check_id, v_total, 0, v_method, 'completed');

      END LOOP;
    END LOOP;
  END LOOP;

  DROP TABLE tmp_menu_items;
END $$;
