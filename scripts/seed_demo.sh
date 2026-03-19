#!/usr/bin/env bash
# Seed demo data: 1 org "Bistro Cloud" with 2 locations + orders/menu/inventory
set -euo pipefail

API="http://localhost:8080/api/v1"
DB="postgres://fireline:fireline@localhost:5432/fireline?sslmode=disable"

echo "=== Step 1: Signup owner account ==="
SIGNUP=$(curl -s -X POST "$API/auth/signup" \
  -H "Content-Type: application/json" \
  -d '{
    "org_name": "Bistro Cloud",
    "org_slug": "bistro-cloud",
    "email": "owner@bistrocloud.com",
    "password": "DemoPassword1234!",
    "display_name": "Alex Rivera"
  }')

echo "$SIGNUP" | python3 -m json.tool 2>/dev/null || echo "$SIGNUP"

ORG_ID=$(echo "$SIGNUP" | python3 -c "import sys,json; print(json.load(sys.stdin)['org_id'])")
USER_ID=$(echo "$SIGNUP" | python3 -c "import sys,json; print(json.load(sys.stdin)['user_id'])")
TOKEN=$(echo "$SIGNUP" | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

echo ""
echo "org_id:  $ORG_ID"
echo "user_id: $USER_ID"
echo ""

echo "=== Step 2: Create 2 locations + seed data via SQL ==="
docker exec -i fireline-postgres-1 psql -U fireline -d fireline <<SQL

-- Two restaurant locations
INSERT INTO locations (location_id, org_id, name, address, timezone, status) VALUES
  ('a1111111-1111-1111-1111-111111111111', '$ORG_ID', 'Downtown Flagship', '123 Main St, Austin TX 78701', 'America/Chicago', 'active'),
  ('b2222222-2222-2222-2222-222222222222', '$ORG_ID', 'Airport Terminal 4', '3600 Presidential Blvd, Austin TX 78719', 'America/Chicago', 'active');

-- Grant user access to both locations
INSERT INTO user_location_access (user_id, location_id, org_id) VALUES
  ('$USER_ID', 'a1111111-1111-1111-1111-111111111111', '$ORG_ID'),
  ('$USER_ID', 'b2222222-2222-2222-2222-222222222222', '$ORG_ID');

-- === INGREDIENTS ===
INSERT INTO ingredients (ingredient_id, org_id, name, category, unit, cost_per_unit, prep_yield_factor) VALUES
  ('11111111-aaaa-1111-aaaa-111111111111', '$ORG_ID', 'Ground Beef (80/20)', 'protein', 'lb', 450, 0.9500),
  ('11111111-aaaa-2222-aaaa-111111111111', '$ORG_ID', 'Chicken Breast', 'protein', 'lb', 375, 0.8500),
  ('11111111-aaaa-3333-aaaa-111111111111', '$ORG_ID', 'Cheddar Cheese', 'dairy', 'lb', 520, 1.0000),
  ('11111111-aaaa-4444-aaaa-111111111111', '$ORG_ID', 'Romaine Lettuce', 'produce', 'head', 200, 0.7500),
  ('11111111-aaaa-5555-aaaa-111111111111', '$ORG_ID', 'Tomatoes', 'produce', 'lb', 280, 0.9000),
  ('11111111-aaaa-6666-aaaa-111111111111', '$ORG_ID', 'Brioche Buns', 'bakery', 'ea', 85, 1.0000),
  ('11111111-aaaa-7777-aaaa-111111111111', '$ORG_ID', 'French Fries (frozen)', 'frozen', 'lb', 195, 1.0000),
  ('11111111-aaaa-8888-aaaa-111111111111', '$ORG_ID', 'Avocado', 'produce', 'ea', 175, 0.6500),
  ('11111111-aaaa-9999-aaaa-111111111111', '$ORG_ID', 'Bacon', 'protein', 'lb', 680, 0.5000),
  ('11111111-aaaa-aaaa-aaaa-111111111111', '$ORG_ID', 'House Dressing', 'sauce', 'oz', 25, 1.0000);

-- === INGREDIENT LOCATION CONFIGS (PAR levels) ===
-- Downtown (higher volume)
INSERT INTO ingredient_location_configs (org_id, ingredient_id, location_id, vendor_name, local_cost_per_unit, par_level, reorder_point) VALUES
  ('$ORG_ID', '11111111-aaaa-1111-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'US Foods', 440, 50.00, 25.00),
  ('$ORG_ID', '11111111-aaaa-2222-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'US Foods', 365, 40.00, 20.00),
  ('$ORG_ID', '11111111-aaaa-3333-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 510, 25.00, 12.00),
  ('$ORG_ID', '11111111-aaaa-4444-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'Local Farm Co', 180, 20.00, 10.00),
  ('$ORG_ID', '11111111-aaaa-5555-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'Local Farm Co', 260, 30.00, 15.00),
  ('$ORG_ID', '11111111-aaaa-6666-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'Artisan Bakery', 80, 100.00, 50.00),
  ('$ORG_ID', '11111111-aaaa-7777-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 185, 60.00, 30.00),
  ('$ORG_ID', '11111111-aaaa-8888-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'Local Farm Co', 165, 40.00, 20.00),
  ('$ORG_ID', '11111111-aaaa-9999-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'US Foods', 660, 20.00, 10.00),
  ('$ORG_ID', '11111111-aaaa-aaaa-aaaa-111111111111', 'a1111111-1111-1111-1111-111111111111', 'House Made', 20, 128.00, 64.00);

-- Airport (lower volume, higher costs)
INSERT INTO ingredient_location_configs (org_id, ingredient_id, location_id, vendor_name, local_cost_per_unit, par_level, reorder_point) VALUES
  ('$ORG_ID', '11111111-aaaa-1111-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 480, 30.00, 15.00),
  ('$ORG_ID', '11111111-aaaa-2222-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 395, 25.00, 12.00),
  ('$ORG_ID', '11111111-aaaa-3333-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 540, 15.00, 8.00),
  ('$ORG_ID', '11111111-aaaa-4444-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 220, 12.00, 6.00),
  ('$ORG_ID', '11111111-aaaa-5555-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 310, 18.00, 9.00),
  ('$ORG_ID', '11111111-aaaa-6666-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Airport Supply', 110, 60.00, 30.00),
  ('$ORG_ID', '11111111-aaaa-7777-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 210, 35.00, 18.00),
  ('$ORG_ID', '11111111-aaaa-8888-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 195, 24.00, 12.00),
  ('$ORG_ID', '11111111-aaaa-9999-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'Sysco', 710, 12.00, 6.00),
  ('$ORG_ID', '11111111-aaaa-aaaa-aaaa-111111111111', 'b2222222-2222-2222-2222-222222222222', 'House Made', 25, 80.00, 40.00);

-- === MENU ITEMS ===
-- Downtown menu
INSERT INTO menu_items (menu_item_id, org_id, location_id, name, category, price, source) VALUES
  ('22222222-aaaa-1111-aaaa-111111111111', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Classic Burger', 'burgers', 1495, 'toast'),
  ('22222222-aaaa-2222-aaaa-111111111111', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Bacon Avocado Burger', 'burgers', 1795, 'toast'),
  ('22222222-aaaa-3333-aaaa-111111111111', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Grilled Chicken Sandwich', 'sandwiches', 1395, 'toast'),
  ('22222222-aaaa-4444-aaaa-111111111111', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Caesar Salad', 'salads', 1195, 'toast'),
  ('22222222-aaaa-5555-aaaa-111111111111', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Loaded Fries', 'sides', 895, 'toast'),
  ('22222222-aaaa-6666-aaaa-111111111111', '$ORG_ID', 'a1111111-1111-1111-1111-111111111111', 'Side Salad', 'sides', 595, 'toast');

-- Airport menu (same items, higher prices)
INSERT INTO menu_items (menu_item_id, org_id, location_id, name, category, price, source) VALUES
  ('33333333-aaaa-1111-aaaa-111111111111', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'Classic Burger', 'burgers', 1695, 'toast'),
  ('33333333-aaaa-2222-aaaa-111111111111', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'Bacon Avocado Burger', 'burgers', 1995, 'toast'),
  ('33333333-aaaa-3333-aaaa-111111111111', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'Grilled Chicken Sandwich', 'sandwiches', 1595, 'toast'),
  ('33333333-aaaa-4444-aaaa-111111111111', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'Caesar Salad', 'salads', 1395, 'toast'),
  ('33333333-aaaa-5555-aaaa-111111111111', '$ORG_ID', 'b2222222-2222-2222-2222-222222222222', 'Loaded Fries', 'sides', 1095, 'toast');

-- === RECIPES + RECIPE INGREDIENTS ===
-- Classic Burger recipe (Downtown)
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit) VALUES
  ('44444444-aaaa-1111-aaaa-111111111111', '$ORG_ID', '22222222-aaaa-1111-aaaa-111111111111', 'Classic Burger', 1.00, 'ea');
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('$ORG_ID', '44444444-aaaa-1111-aaaa-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 0.3333, 'lb'),
  ('$ORG_ID', '44444444-aaaa-1111-aaaa-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 1.0000, 'ea'),
  ('$ORG_ID', '44444444-aaaa-1111-aaaa-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 0.1250, 'lb'),
  ('$ORG_ID', '44444444-aaaa-1111-aaaa-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 0.1000, 'head'),
  ('$ORG_ID', '44444444-aaaa-1111-aaaa-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 0.1250, 'lb');

-- Materialize recipe explosion for Classic Burger
INSERT INTO recipe_explosion (org_id, menu_item_id, ingredient_id, quantity_per_unit, unit) VALUES
  ('$ORG_ID', '22222222-aaaa-1111-aaaa-111111111111', '11111111-aaaa-1111-aaaa-111111111111', 0.333300, 'lb'),
  ('$ORG_ID', '22222222-aaaa-1111-aaaa-111111111111', '11111111-aaaa-6666-aaaa-111111111111', 1.000000, 'ea'),
  ('$ORG_ID', '22222222-aaaa-1111-aaaa-111111111111', '11111111-aaaa-3333-aaaa-111111111111', 0.125000, 'lb'),
  ('$ORG_ID', '22222222-aaaa-1111-aaaa-111111111111', '11111111-aaaa-4444-aaaa-111111111111', 0.100000, 'head'),
  ('$ORG_ID', '22222222-aaaa-1111-aaaa-111111111111', '11111111-aaaa-5555-aaaa-111111111111', 0.125000, 'lb');

-- === CHECKS (orders) - Downtown: busy lunch day ===
-- Generate 45 dine-in orders, 20 takeout, 15 delivery for Downtown
DO \$\$
DECLARE
  i INT;
  ch_id UUID;
  ch_channel TEXT;
  ch_items INT;
  item_price INT;
  item_name TEXT;
  item_id UUID;
  ch_subtotal INT;
  ch_tax INT;
  ch_tip INT;
BEGIN
  FOR i IN 1..80 LOOP
    ch_id := gen_random_uuid();

    -- Assign channel
    IF i <= 45 THEN ch_channel := 'dine_in';
    ELSIF i <= 65 THEN ch_channel := 'takeout';
    ELSE ch_channel := 'delivery';
    END IF;

    -- Pick a menu item
    CASE (i % 6)
      WHEN 0 THEN item_id := '22222222-aaaa-1111-aaaa-111111111111'; item_name := 'Classic Burger'; item_price := 1495;
      WHEN 1 THEN item_id := '22222222-aaaa-2222-aaaa-111111111111'; item_name := 'Bacon Avocado Burger'; item_price := 1795;
      WHEN 2 THEN item_id := '22222222-aaaa-3333-aaaa-111111111111'; item_name := 'Grilled Chicken Sandwich'; item_price := 1395;
      WHEN 3 THEN item_id := '22222222-aaaa-4444-aaaa-111111111111'; item_name := 'Caesar Salad'; item_price := 1195;
      WHEN 4 THEN item_id := '22222222-aaaa-5555-aaaa-111111111111'; item_name := 'Loaded Fries'; item_price := 895;
      WHEN 5 THEN item_id := '22222222-aaaa-6666-aaaa-111111111111'; item_name := 'Side Salad'; item_price := 595;
    END CASE;

    -- Random quantity 1-3
    ch_items := 1 + (random() * 2)::INT;
    ch_subtotal := item_price * ch_items;
    ch_tax := (ch_subtotal * 0.0825)::INT;
    ch_tip := CASE WHEN ch_channel = 'dine_in' THEN (ch_subtotal * 0.18)::INT ELSE 0 END;

    INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel, subtotal, tax, total, tip, opened_at, closed_at, source)
    VALUES (ch_id, '$ORG_ID', 'a1111111-1111-1111-1111-111111111111',
            'DT-' || LPAD(i::TEXT, 4, '0'), 'closed', ch_channel,
            ch_subtotal, ch_tax, ch_subtotal + ch_tax, ch_tip,
            now() - INTERVAL '8 hours' + (i * INTERVAL '5 minutes'),
            now() - INTERVAL '8 hours' + (i * INTERVAL '5 minutes') + INTERVAL '25 minutes',
            'toast');

    INSERT INTO check_items (org_id, check_id, menu_item_id, name, quantity, unit_price, fired_at)
    VALUES ('$ORG_ID', ch_id, item_id, item_name, ch_items, item_price,
            now() - INTERVAL '8 hours' + (i * INTERVAL '5 minutes') + INTERVAL '2 minutes');

    -- Payment
    INSERT INTO payments (org_id, check_id, amount, tip, method)
    VALUES ('$ORG_ID', ch_id, ch_subtotal + ch_tax, ch_tip,
            CASE WHEN random() > 0.3 THEN 'card' ELSE 'cash' END);
  END LOOP;
END;
\$\$;

-- === CHECKS (orders) - Airport: moderate volume ===
DO \$\$
DECLARE
  i INT;
  ch_id UUID;
  ch_channel TEXT;
  item_price INT;
  item_name TEXT;
  item_id UUID;
  ch_items INT;
  ch_subtotal INT;
  ch_tax INT;
  ch_tip INT;
BEGIN
  FOR i IN 1..45 LOOP
    ch_id := gen_random_uuid();

    IF i <= 30 THEN ch_channel := 'dine_in';
    ELSE ch_channel := 'takeout';
    END IF;

    CASE (i % 5)
      WHEN 0 THEN item_id := '33333333-aaaa-1111-aaaa-111111111111'; item_name := 'Classic Burger'; item_price := 1695;
      WHEN 1 THEN item_id := '33333333-aaaa-2222-aaaa-111111111111'; item_name := 'Bacon Avocado Burger'; item_price := 1995;
      WHEN 2 THEN item_id := '33333333-aaaa-3333-aaaa-111111111111'; item_name := 'Grilled Chicken Sandwich'; item_price := 1595;
      WHEN 3 THEN item_id := '33333333-aaaa-4444-aaaa-111111111111'; item_name := 'Caesar Salad'; item_price := 1395;
      WHEN 4 THEN item_id := '33333333-aaaa-5555-aaaa-111111111111'; item_name := 'Loaded Fries'; item_price := 1095;
    END CASE;

    ch_items := 1 + (random() * 1.5)::INT;
    ch_subtotal := item_price * ch_items;
    ch_tax := (ch_subtotal * 0.0825)::INT;
    ch_tip := CASE WHEN ch_channel = 'dine_in' THEN (ch_subtotal * 0.15)::INT ELSE 0 END;

    INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel, subtotal, tax, total, tip, opened_at, closed_at, source)
    VALUES (ch_id, '$ORG_ID', 'b2222222-2222-2222-2222-222222222222',
            'AP-' || LPAD(i::TEXT, 4, '0'), 'closed', ch_channel,
            ch_subtotal, ch_tax, ch_subtotal + ch_tax, ch_tip,
            now() - INTERVAL '6 hours' + (i * INTERVAL '7 minutes'),
            now() - INTERVAL '6 hours' + (i * INTERVAL '7 minutes') + INTERVAL '20 minutes',
            'toast');

    INSERT INTO check_items (org_id, check_id, menu_item_id, name, quantity, unit_price, fired_at)
    VALUES ('$ORG_ID', ch_id, item_id, item_name, ch_items, item_price,
            now() - INTERVAL '6 hours' + (i * INTERVAL '7 minutes') + INTERVAL '2 minutes');

    INSERT INTO payments (org_id, check_id, amount, tip, method)
    VALUES ('$ORG_ID', ch_id, ch_subtotal + ch_tax, ch_tip,
            CASE WHEN random() > 0.2 THEN 'card' ELSE 'cash' END);
  END LOOP;
END;
\$\$;

SQL

echo ""
echo "=== Seed complete! ==="
echo ""
echo "Organization: Bistro Cloud ($ORG_ID)"
echo "Locations:"
echo "  1. Downtown Flagship  (a1111111-...) - 80 orders, 6 menu items"
echo "  2. Airport Terminal 4  (b2222222-...) - 45 orders, 5 menu items"
echo ""
echo "Login credentials:"
echo "  Email:    owner@bistrocloud.com"
echo "  Password: DemoPassword1234!"
echo ""
echo "Open http://localhost:3000/login to test"
