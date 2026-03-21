-- =============================================================================
-- Nimbu International/Fusion Menu - Complete Replacement
-- Replaces Peruvian menu with 20-item international fusion menu
-- =============================================================================

SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

BEGIN;

-- =============================================================================
-- STEP 1: Delete existing menu-related data (FK order)
-- =============================================================================

-- FK children of checks
DELETE FROM kds_ticket_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM kds_tickets WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM loyalty_transactions WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM guest_visits WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM check_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM payments WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM checks WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- Menu-related
DELETE FROM menu_simulations WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM menu_item_resource_profiles WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM item_id_mappings WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM recipe_explosion WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM recipe_ingredients WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM recipes WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM menu_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- Ingredient-related
DELETE FROM waste_logs WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM inventory_count_lines WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM inventory_variances WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM purchase_order_lines WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM ingredient_location_configs WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM ingredient_price_history WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM ingredients WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- =============================================================================
-- STEP 2: Insert 30 Ingredients
-- =============================================================================

INSERT INTO ingredients (ingredient_id, org_id, name, category, unit, cost_per_unit, prep_yield_factor, allergens, status) VALUES
('10000001-0001-0001-0001-000000000001','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Ribeye Steak','protein','kg',180000,0.8500,'{}','active'),
('10000001-0001-0001-0001-000000000002','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Sea Bass Fillet','protein','kg',120000,0.9000,'{"fish"}','active'),
('10000001-0001-0001-0001-000000000003','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Chicken Breast','protein','kg',8000,0.9000,'{}','active'),
('10000001-0001-0001-0001-000000000004','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Lamb Rack','protein','kg',200000,0.7500,'{}','active'),
('10000001-0001-0001-0001-000000000005','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Lobster Tail','protein','ea',95000,0.8000,'{"shellfish"}','active'),
('10000001-0001-0001-0001-000000000006','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Duck Leg','protein','ea',45000,0.8000,'{}','active'),
('10000001-0001-0001-0001-000000000007','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Wagyu Beef','protein','kg',350000,0.8500,'{}','active'),
('10000001-0001-0001-0001-000000000008','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Salmon Fillet','protein','kg',140000,0.9000,'{"fish"}','active'),
('10000001-0001-0001-0001-000000000009','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Tuna Sashimi Grade','protein','kg',160000,0.9500,'{"fish"}','active'),
('10000001-0001-0001-0001-000000000010','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Tiger Prawns','protein','kg',110000,0.8000,'{"shellfish"}','active'),
('10000001-0001-0001-0001-000000000011','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Burrata Cheese','dairy','ea',12000,1.0000,'{"dairy"}','active'),
('10000001-0001-0001-0001-000000000012','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Truffle Oil','condiment','ml',500,1.0000,'{}','active'),
('10000001-0001-0001-0001-000000000013','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Fresh Truffle','condiment','g',3000,1.0000,'{}','active'),
('10000001-0001-0001-0001-000000000014','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Linguine Pasta','grain','kg',4000,1.0000,'{"gluten"}','active'),
('10000001-0001-0001-0001-000000000015','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Japanese Rice','grain','kg',3500,1.0000,'{}','active'),
('10000001-0001-0001-0001-000000000016','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Panko Breadcrumbs','grain','kg',5000,1.0000,'{"gluten"}','active'),
('10000001-0001-0001-0001-000000000017','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Mixed Greens','produce','kg',6000,0.9000,'{}','active'),
('10000001-0001-0001-0001-000000000018','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Cherry Tomatoes','produce','kg',8000,0.9500,'{}','active'),
('10000001-0001-0001-0001-000000000019','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Avocado','produce','ea',3500,0.7500,'{}','active'),
('10000001-0001-0001-0001-000000000020','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Lemon','produce','ea',500,0.8000,'{}','active'),
('10000001-0001-0001-0001-000000000021','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Lime','produce','ea',400,0.8000,'{}','active'),
('10000001-0001-0001-0001-000000000022','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Ginger','produce','kg',7000,0.8500,'{}','active'),
('10000001-0001-0001-0001-000000000023','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Edamame','produce','kg',12000,1.0000,'{"soy"}','active'),
('10000001-0001-0001-0001-000000000024','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Chickpeas','produce','kg',3000,1.0000,'{}','active'),
('10000001-0001-0001-0001-000000000025','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Tahini','condiment','kg',8000,1.0000,'{"sesame"}','active'),
('10000001-0001-0001-0001-000000000026','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Curry Paste','condiment','kg',15000,1.0000,'{}','active'),
('10000001-0001-0001-0001-000000000027','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Soy Sauce','condiment','ltr',5000,1.0000,'{"soy","gluten"}','active'),
('10000001-0001-0001-0001-000000000028','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','White Wine','beverage','ltr',15000,1.0000,'{}','active'),
('10000001-0001-0001-0001-000000000029','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Rum','beverage','ltr',25000,1.0000,'{}','active'),
('10000001-0001-0001-0001-000000000030','3f7ef589-f499-43e3-a1c5-aaacd9d543ec','Japanese Whisky','beverage','ltr',80000,1.0000,'{}','active');

-- =============================================================================
-- STEP 3: Insert Menu Items for ALL 4 locations (20 items x 4 = 80 rows)
-- =============================================================================

-- Use deterministic UUIDs: loc_num in 4th block, item_num in last block
INSERT INTO menu_items (menu_item_id, org_id, location_id, name, category, price, available, description, source)
SELECT
  ('aa000000-0000-0000-' || lpad(loc.loc_num::text,4,'0') || '-' || lpad(mi_def.item_num::text,12,'0'))::uuid,
  '3f7ef589-f499-43e3-a1c5-aaacd9d543ec',
  loc.location_id,
  mi_def.name,
  mi_def.category,
  mi_def.price,
  true,
  mi_def.description,
  'manual'
FROM (VALUES
  -- STARTERS
  ('Truffle Burrata','starters',32000,'Creamy burrata with truffle oil, cherry tomatoes, and mixed greens',1),
  ('Tuna Tartare','starters',38000,'Sashimi-grade tuna with avocado, soy dressing, and sesame',2),
  ('Tom Yum Soup','starters',22000,'Spicy Thai soup with tiger prawns, lemongrass, and ginger',3),
  ('Wagyu Carpaccio','starters',45000,'Thinly sliced wagyu with truffle oil and mixed greens',4),
  ('Hummus Trio','starters',18000,'Classic, roasted garlic, and beetroot hummus with warm bread',5),
  ('Prawn Tempura','starters',28000,'Crispy tiger prawns with soy dipping sauce',6),
  -- MAINS
  ('Grilled Ribeye 300g','mains',68000,'Prime ribeye steak with truffle fries and grilled vegetables',7),
  ('Pan-Seared Sea Bass','mains',55000,'Mediterranean sea bass with cherry tomatoes and white wine sauce',8),
  ('Chicken Katsu Curry','mains',38000,'Panko-crusted chicken with Japanese curry and steamed rice',9),
  ('Lamb Rack','mains',72000,'French-trimmed lamb rack with herb crust and grilled vegetables',10),
  ('Lobster Linguine','mains',85000,'Whole lobster tail with linguine in white wine and cherry tomato sauce',11),
  ('Duck Confit','mains',62000,'Slow-cooked duck leg with truffle fries and mixed greens',12),
  ('Wagyu Burger','mains',48000,'Wagyu beef patty with truffle aioli, avocado, and fries',13),
  ('Grilled Salmon Teriyaki','mains',52000,'Norwegian salmon with teriyaki glaze and steamed rice',14),
  -- SIDES
  ('Truffle Fries','sides',16000,'Hand-cut fries with truffle oil and parmesan',15),
  ('Edamame','sides',12000,'Steamed edamame with sea salt',16),
  ('Grilled Vegetables','sides',14000,'Seasonal grilled vegetables with herbs',17),
  -- BEVERAGES
  ('Signature Mojito','beverages',18000,'Fresh mint, lime, rum, and soda',18),
  ('Japanese Whisky Sour','beverages',22000,'Japanese whisky with fresh lemon and ginger',19),
  ('Fresh Lemonade','beverages',8000,'Freshly squeezed lemon with mint',20)
) AS mi_def(name, category, price, description, item_num)
CROSS JOIN (VALUES
  (1, 'a1111111-1111-1111-1111-111111111111'::uuid),
  (2, 'b2222222-2222-2222-2222-222222222222'::uuid),
  (3, 'c3333333-3333-3333-3333-333333333333'::uuid),
  (4, 'd4444444-4444-4444-4444-444444444444'::uuid)
) AS loc(loc_num, location_id);

-- =============================================================================
-- STEP 4: Ingredient Location Configs for all 4 locations x 30 ingredients
-- =============================================================================

INSERT INTO ingredient_location_configs (org_id, ingredient_id, location_id, vendor_name, vendor_item_code, local_cost_per_unit, par_level, reorder_point, lead_time_days, avg_daily_usage)
SELECT
  '3f7ef589-f499-43e3-a1c5-aaacd9d543ec',
  i.ingredient_id,
  l.location_id,
  CASE
    WHEN i.name IN ('Ribeye Steak','Lamb Rack','Wagyu Beef','Duck Leg') THEN 'Premium Meats Co'
    WHEN i.name IN ('Sea Bass Fillet','Lobster Tail','Salmon Fillet','Tuna Sashimi Grade','Tiger Prawns') THEN 'Ocean Fresh Egypt'
    WHEN i.name IN ('Chicken Breast','Mixed Greens','Cherry Tomatoes','Avocado','Lemon','Lime','Ginger','Edamame','Chickpeas') THEN 'Metro Market'
    ELSE 'Specialty Imports'
  END,
  'VND-' || left(i.ingredient_id::text, 8),
  i.cost_per_unit,
  CASE
    WHEN i.category = 'protein' THEN 15.00
    WHEN i.category = 'dairy' THEN 20.00
    WHEN i.category = 'produce' THEN 25.00
    WHEN i.category = 'grain' THEN 30.00
    WHEN i.category = 'condiment' THEN 10.00
    WHEN i.category = 'beverage' THEN 8.00
    ELSE 10.00
  END,
  CASE
    WHEN i.category = 'protein' THEN 5.00
    WHEN i.category = 'dairy' THEN 8.00
    WHEN i.category = 'produce' THEN 10.00
    WHEN i.category = 'grain' THEN 12.00
    WHEN i.category = 'condiment' THEN 4.00
    WHEN i.category = 'beverage' THEN 3.00
    ELSE 5.00
  END,
  CASE
    WHEN i.category = 'protein' THEN 2
    WHEN i.category = 'produce' THEN 1
    ELSE 3
  END,
  CASE
    WHEN i.category = 'protein' THEN 3.5000
    WHEN i.category = 'dairy' THEN 5.0000
    WHEN i.category = 'produce' THEN 6.0000
    WHEN i.category = 'grain' THEN 4.0000
    WHEN i.category = 'condiment' THEN 2.0000
    WHEN i.category = 'beverage' THEN 1.5000
    ELSE 2.0000
  END
FROM ingredients i
CROSS JOIN (VALUES
  ('a1111111-1111-1111-1111-111111111111'::uuid),
  ('b2222222-2222-2222-2222-222222222222'::uuid),
  ('c3333333-3333-3333-3333-333333333333'::uuid),
  ('d4444444-4444-4444-4444-444444444444'::uuid)
) AS l(location_id)
WHERE i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- =============================================================================
-- STEP 5: Recipes (one per menu item, using El Gouna items as reference)
-- =============================================================================

-- We create one recipe per menu item name (using El Gouna location items)
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes, status)
SELECT
  gen_random_uuid(),
  '3f7ef589-f499-43e3-a1c5-aaacd9d543ec',
  mi.menu_item_id,
  mi.name || ' Recipe',
  1.00,
  'ea',
  CASE mi.category
    WHEN 'starters' THEN 15
    WHEN 'mains' THEN 25
    WHEN 'sides' THEN 10
    WHEN 'beverages' THEN 5
  END,
  'active'
FROM menu_items mi
WHERE mi.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
  AND mi.location_id = 'a1111111-1111-1111-1111-111111111111';

-- =============================================================================
-- STEP 5b: Recipe Ingredients
-- =============================================================================

-- Helper: insert recipe ingredients by matching menu item name to recipe
-- Truffle Burrata: burrata, truffle oil, cherry tomatoes, mixed greens
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Burrata Cheese', 1.0000, 'ea'),
  ('Truffle Oil', 10.0000, 'ml'),
  ('Cherry Tomatoes', 0.0800, 'kg'),
  ('Mixed Greens', 0.0500, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Truffle Burrata';

-- Tuna Tartare: tuna, avocado, soy sauce, mixed greens
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Tuna Sashimi Grade', 0.1200, 'kg'),
  ('Avocado', 0.5000, 'ea'),
  ('Soy Sauce', 0.0200, 'ltr'),
  ('Mixed Greens', 0.0300, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Tuna Tartare';

-- Tom Yum Soup: tiger prawns, ginger, lime, lemon
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Tiger Prawns', 0.0800, 'kg'),
  ('Ginger', 0.0200, 'kg'),
  ('Lime', 1.0000, 'ea'),
  ('Lemon', 0.5000, 'ea')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Tom Yum Soup';

-- Wagyu Carpaccio: wagyu, truffle oil, mixed greens, lemon
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Wagyu Beef', 0.0800, 'kg'),
  ('Truffle Oil', 5.0000, 'ml'),
  ('Mixed Greens', 0.0400, 'kg'),
  ('Lemon', 0.5000, 'ea')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Wagyu Carpaccio';

-- Hummus Trio: chickpeas, tahini, lemon, avocado
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Chickpeas', 0.2000, 'kg'),
  ('Tahini', 0.0500, 'kg'),
  ('Lemon', 1.0000, 'ea'),
  ('Cherry Tomatoes', 0.0300, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Hummus Trio';

-- Prawn Tempura: tiger prawns, panko, soy sauce
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Tiger Prawns', 0.1500, 'kg'),
  ('Panko Breadcrumbs', 0.0500, 'kg'),
  ('Soy Sauce', 0.0300, 'ltr')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Prawn Tempura';

-- Grilled Ribeye 300g: ribeye, truffle oil, mixed greens, cherry tomatoes
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Ribeye Steak', 0.3000, 'kg'),
  ('Truffle Oil', 5.0000, 'ml'),
  ('Mixed Greens', 0.0400, 'kg'),
  ('Cherry Tomatoes', 0.0500, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Grilled Ribeye 300g';

-- Pan-Seared Sea Bass: sea bass, white wine, cherry tomatoes, lemon
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Sea Bass Fillet', 0.2500, 'kg'),
  ('White Wine', 0.0500, 'ltr'),
  ('Cherry Tomatoes', 0.0800, 'kg'),
  ('Lemon', 1.0000, 'ea')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Pan-Seared Sea Bass';

-- Chicken Katsu Curry: chicken, panko, curry paste, japanese rice
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Chicken Breast', 0.2500, 'kg'),
  ('Panko Breadcrumbs', 0.0500, 'kg'),
  ('Curry Paste', 0.0300, 'kg'),
  ('Japanese Rice', 0.1500, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Chicken Katsu Curry';

-- Lamb Rack: lamb rack, mixed greens, cherry tomatoes, truffle oil
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Lamb Rack', 0.3500, 'kg'),
  ('Mixed Greens', 0.0500, 'kg'),
  ('Cherry Tomatoes', 0.0500, 'kg'),
  ('Truffle Oil', 5.0000, 'ml')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Lamb Rack';

-- Lobster Linguine: lobster tail, linguine, white wine, cherry tomatoes, lemon
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Lobster Tail', 1.0000, 'ea'),
  ('Linguine Pasta', 0.1500, 'kg'),
  ('White Wine', 0.0500, 'ltr'),
  ('Cherry Tomatoes', 0.0800, 'kg'),
  ('Lemon', 0.5000, 'ea')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Lobster Linguine';

-- Duck Confit: duck leg, truffle oil, mixed greens
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Duck Leg', 1.0000, 'ea'),
  ('Truffle Oil', 5.0000, 'ml'),
  ('Mixed Greens', 0.0500, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Duck Confit';

-- Wagyu Burger: wagyu, avocado, truffle oil, mixed greens
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Wagyu Beef', 0.2000, 'kg'),
  ('Avocado', 0.5000, 'ea'),
  ('Truffle Oil', 5.0000, 'ml'),
  ('Mixed Greens', 0.0300, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Wagyu Burger';

-- Grilled Salmon Teriyaki: salmon, soy sauce, ginger, japanese rice
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Salmon Fillet', 0.2200, 'kg'),
  ('Soy Sauce', 0.0300, 'ltr'),
  ('Ginger', 0.0100, 'kg'),
  ('Japanese Rice', 0.1500, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Grilled Salmon Teriyaki';

-- Truffle Fries: fresh truffle, truffle oil (potatoes assumed house stock)
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Fresh Truffle', 3.0000, 'g'),
  ('Truffle Oil', 10.0000, 'ml')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Truffle Fries';

-- Edamame: edamame
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Edamame', 0.1500, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Edamame';

-- Grilled Vegetables: mixed greens, cherry tomatoes, avocado
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Mixed Greens', 0.1000, 'kg'),
  ('Cherry Tomatoes', 0.0800, 'kg'),
  ('Avocado', 0.5000, 'ea')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Grilled Vegetables';

-- Signature Mojito: rum, lime, lemon
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Rum', 0.0600, 'ltr'),
  ('Lime', 2.0000, 'ea'),
  ('Lemon', 0.5000, 'ea')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Signature Mojito';

-- Japanese Whisky Sour: japanese whisky, lemon, ginger
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Japanese Whisky', 0.0600, 'ltr'),
  ('Lemon', 1.0000, 'ea'),
  ('Ginger', 0.0100, 'kg')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Japanese Whisky Sour';

-- Fresh Lemonade: lemon, lime
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', r.recipe_id, i.ingredient_id, v.qty, v.unit
FROM recipes r
JOIN menu_items mi ON r.menu_item_id = mi.menu_item_id
CROSS JOIN (VALUES
  ('Lemon', 3.0000, 'ea'),
  ('Lime', 1.0000, 'ea')
) AS v(ing_name, qty, unit)
JOIN ingredients i ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
WHERE mi.name = 'Fresh Lemonade';

-- =============================================================================
-- STEP 5c: Recipe Explosion (materialized view of ingredient usage per menu item)
-- For ALL menu items across ALL locations
-- =============================================================================

INSERT INTO recipe_explosion (org_id, menu_item_id, ingredient_id, quantity_per_unit, unit)
SELECT DISTINCT
  '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'::uuid,
  mi_all.menu_item_id,
  ri.ingredient_id,
  ri.quantity,
  ri.unit
FROM menu_items mi_all
JOIN menu_items mi_ref ON mi_ref.name = mi_all.name
  AND mi_ref.location_id = 'a1111111-1111-1111-1111-111111111111'
  AND mi_ref.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
JOIN recipes r ON r.menu_item_id = mi_ref.menu_item_id
JOIN recipe_ingredients ri ON ri.recipe_id = r.recipe_id
WHERE mi_all.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- =============================================================================
-- STEP 6: Resource Profiles for all menu items
-- =============================================================================

INSERT INTO menu_item_resource_profiles (org_id, menu_item_id, station_type, task_sequence, duration_secs, elu_required, batch_size)
SELECT
  '3f7ef589-f499-43e3-a1c5-aaacd9d543ec',
  mi.menu_item_id,
  v.station_type,
  v.task_seq,
  v.duration,
  v.elu,
  1
FROM menu_items mi
JOIN (VALUES
  ('Truffle Burrata','prep',1,180,1.0),
  ('Tuna Tartare','prep',1,240,1.0),
  ('Tom Yum Soup','saute',1,300,1.0),
  ('Wagyu Carpaccio','prep',1,200,1.0),
  ('Hummus Trio','prep',1,180,0.8),
  ('Prawn Tempura','fryer',1,240,1.0),
  ('Grilled Ribeye 300g','grill',1,480,1.5),
  ('Pan-Seared Sea Bass','saute',1,420,1.2),
  ('Chicken Katsu Curry','fryer',1,360,1.0),('Chicken Katsu Curry','saute',2,180,0.8),
  ('Lamb Rack','grill',1,540,1.5),
  ('Lobster Linguine','saute',1,480,1.5),('Lobster Linguine','grill',2,300,1.0),
  ('Duck Confit','saute',1,600,1.2),
  ('Wagyu Burger','grill',1,360,1.2),
  ('Grilled Salmon Teriyaki','grill',1,360,1.2),
  ('Truffle Fries','fryer',1,180,0.8),
  ('Edamame','prep',1,120,0.5),
  ('Grilled Vegetables','grill',1,180,0.8),
  ('Signature Mojito','bar',1,90,0.5),
  ('Japanese Whisky Sour','bar',1,120,0.6),
  ('Fresh Lemonade','bar',1,60,0.3)
) AS v(item_name, station_type, task_seq, duration, elu)
ON mi.name = v.item_name
WHERE mi.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- =============================================================================
-- STEP 8: Menu Scores (set before orders so classification is ready)
-- =============================================================================

UPDATE menu_items SET
  margin_score = v.margin,
  velocity_score = v.velocity,
  complexity_score = v.complexity,
  satisfaction_score = v.satisfaction,
  strategic_score = v.strategic,
  classification = v.class,
  classification_changed_at = now()
FROM (VALUES
  ('Truffle Burrata',       72.00, 65.00, 30.00, 82.00, 70.00, 'crowd_pleaser'),
  ('Tuna Tartare',          68.00, 55.00, 45.00, 85.00, 65.00, 'hidden_gem'),
  ('Tom Yum Soup',          60.00, 70.00, 25.00, 78.00, 62.00, 'workhorse'),
  ('Wagyu Carpaccio',       85.00, 40.00, 35.00, 90.00, 78.00, 'hidden_gem'),
  ('Hummus Trio',           55.00, 75.00, 15.00, 72.00, 58.00, 'workhorse'),
  ('Prawn Tempura',         62.00, 60.00, 40.00, 80.00, 64.00, 'crowd_pleaser'),
  ('Grilled Ribeye 300g',   78.00, 80.00, 50.00, 92.00, 82.00, 'powerhouse'),
  ('Pan-Seared Sea Bass',   70.00, 55.00, 55.00, 88.00, 72.00, 'complex_star'),
  ('Chicken Katsu Curry',   65.00, 82.00, 45.00, 85.00, 74.00, 'crowd_pleaser'),
  ('Lamb Rack',             80.00, 45.00, 65.00, 90.00, 76.00, 'complex_star'),
  ('Lobster Linguine',      82.00, 70.00, 70.00, 95.00, 85.00, 'powerhouse'),
  ('Duck Confit',           75.00, 50.00, 60.00, 87.00, 70.00, 'complex_star'),
  ('Wagyu Burger',          76.00, 78.00, 35.00, 88.00, 80.00, 'powerhouse'),
  ('Grilled Salmon Teriyaki',72.00, 68.00, 40.00, 86.00, 74.00, 'crowd_pleaser'),
  ('Truffle Fries',         58.00, 85.00, 15.00, 75.00, 60.00, 'workhorse'),
  ('Edamame',               50.00, 70.00, 10.00, 68.00, 52.00, 'workhorse'),
  ('Grilled Vegetables',    55.00, 60.00, 20.00, 72.00, 56.00, 'workhorse'),
  ('Signature Mojito',      70.00, 65.00, 20.00, 80.00, 68.00, 'crowd_pleaser'),
  ('Japanese Whisky Sour',  72.00, 50.00, 25.00, 82.00, 66.00, 'hidden_gem'),
  ('Fresh Lemonade',        45.00, 90.00, 5.00, 70.00, 55.00, 'crowd_pleaser')
) AS v(item_name, margin, velocity, complexity, satisfaction, strategic, class)
WHERE menu_items.name = v.item_name
  AND menu_items.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

COMMIT;
