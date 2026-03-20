-- ============================================================
-- CHICHA: Modern Peruvian Restaurant - Complete Demo Seed
-- Replaces ALL demo data for org 3f7ef589-f499-43e3-a1c5-aaacd9d543ec
-- ============================================================

BEGIN;

-- Constants
\set org_id '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
\set loc_dt 'a1111111-1111-1111-1111-111111111111'
\set loc_ap 'b2222222-2222-2222-2222-222222222222'
\set owner_user_id '0d55e810-1e4a-417a-8a70-08b98f4595c2'

-- ============================================================
-- A. UPDATE ORG + LOCATIONS
-- ============================================================
UPDATE organizations SET name = 'Chicha', slug = 'chicha', updated_at = now()
  WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
UPDATE locations SET name = 'Chicha Downtown', address = '456 Main St, Austin TX 78701', timezone = 'America/Chicago', updated_at = now()
  WHERE location_id = 'a1111111-1111-1111-1111-111111111111';
UPDATE locations SET name = 'Chicha Domain', address = '11410 Century Oaks Terrace, Austin TX 78758', timezone = 'America/Chicago', updated_at = now()
  WHERE location_id = 'b2222222-2222-2222-2222-222222222222';

-- ============================================================
-- B. DELETE ALL EXISTING DEMO DATA (FK order)
-- ============================================================
DELETE FROM kds_ticket_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM kds_tickets WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM loyalty_transactions WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM loyalty_members WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM campaigns WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM staff_point_events WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM shift_swap_requests WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM scheduled_shifts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM schedules WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM labor_demand_forecast WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM best_practices WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM location_benchmarks WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM portfolio_nodes WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM onboarding_checklist_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM onboarding_sessions WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM menu_simulations WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM ingredient_price_history WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM vendor_scores WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM guest_visits WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM guest_profiles WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM inventory_variances WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM inventory_count_lines WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM inventory_counts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM waste_logs WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM purchase_order_lines WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM purchase_orders WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM budgets WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM payments WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM check_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM checks WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM menu_item_resource_profiles WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM recipe_explosion WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM recipe_ingredients WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM recipes WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM kitchen_stations WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM menu_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM ingredient_location_configs WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM ingredients WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
DELETE FROM shifts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- ============================================================
-- H. UPDATE EMPLOYEE NAMES + ELU
-- ============================================================
UPDATE employees SET display_name = 'Isabella Reyes',
  elu_ratings = '{"grill": 4.5, "saute": 4.2, "prep": 4.0, "expo": 4.8, "dish": 3.5}',
  staff_points = 285, certifications = ARRAY['servsafe_manager', 'food_handler']
  WHERE employee_id = 'ee111111-1111-1111-1111-111111111111';

UPDATE employees SET display_name = 'Carlos Mendez',
  elu_ratings = '{"grill": 4.2, "saute": 3.8, "prep": 3.5, "expo": 4.0, "dish": 3.0}',
  staff_points = 210, certifications = ARRAY['food_handler']
  WHERE employee_id = 'ee111111-2222-2222-2222-222222222222';

UPDATE employees SET display_name = 'Ana Lucia Torres',
  elu_ratings = '{"grill": 3.8, "saute": 4.5, "prep": 4.2, "expo": 3.2, "dish": 3.5}',
  staff_points = 175, certifications = ARRAY['food_handler']
  WHERE employee_id = 'ee111111-3333-3333-3333-333333333333';

UPDATE employees SET display_name = 'Diego Vargas',
  elu_ratings = '{"grill": 2.5, "saute": 3.0, "prep": 4.8, "expo": 2.0, "dish": 4.0}',
  staff_points = 140, certifications = ARRAY['food_handler']
  WHERE employee_id = 'ee111111-4444-4444-4444-444444444444';

UPDATE employees SET display_name = 'Sofia Herrera',
  elu_ratings = '{"grill": 1.5, "saute": 1.5, "prep": 2.5, "expo": 4.5, "dish": 3.0}',
  staff_points = 195, certifications = ARRAY['food_handler', 'tips_certified']
  WHERE employee_id = 'ee111111-5555-5555-5555-555555555555';

-- Airport staff
UPDATE employees SET display_name = 'Valentina Cruz',
  elu_ratings = '{"grill": 4.0, "saute": 3.8, "prep": 3.5, "expo": 4.2, "dish": 3.0}',
  staff_points = 160, certifications = ARRAY['servsafe_manager', 'food_handler']
  WHERE employee_id = 'ee222222-1111-1111-1111-111111111111';

UPDATE employees SET display_name = 'Marco Villanueva',
  elu_ratings = '{"grill": 3.5, "saute": 4.0, "prep": 3.8, "expo": 3.5, "dish": 3.0}',
  staff_points = 145, certifications = ARRAY['food_handler']
  WHERE employee_id = 'ee222222-2222-2222-2222-222222222222';

UPDATE employees SET display_name = 'Camila Rojas',
  elu_ratings = '{"grill": 3.0, "saute": 3.5, "prep": 4.0, "expo": 3.2, "dish": 3.5}',
  staff_points = 120, certifications = ARRAY['food_handler']
  WHERE employee_id = 'ee222222-3333-3333-3333-333333333333';

-- ============================================================
-- C. INGREDIENTS (20 items)
-- ============================================================
INSERT INTO ingredients (ingredient_id, org_id, name, category, unit, cost_per_unit, prep_yield_factor, allergens) VALUES
  ('11aaaaaa-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Aji Amarillo Paste', 'sauce', 'oz', 35, 1.0, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Beef Tenderloin', 'protein', 'lb', 1800, 0.85, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Chicken Thighs', 'protein', 'lb', 350, 0.90, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Sea Bass Fillet', 'protein', 'lb', 2200, 0.88, '{fish}'),
  ('11aaaaaa-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Jumbo Shrimp', 'protein', 'lb', 1400, 0.82, '{shellfish}'),
  ('11aaaaaa-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Purple Potatoes', 'produce', 'lb', 280, 0.85, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Sweet Potato', 'produce', 'lb', 150, 0.88, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Corn (Choclo)', 'produce', 'ea', 85, 0.75, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Red Onion', 'produce', 'lb', 120, 0.90, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Lime', 'produce', 'ea', 25, 0.90, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cilantro', 'produce', 'bunch', 90, 0.70, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Avocado', 'produce', 'ea', 175, 0.68, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Quinoa', 'grain', 'lb', 450, 1.0, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cancha Corn', 'grain', 'lb', 320, 1.0, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Huacatay (Black Mint)', 'herb', 'oz', 55, 0.65, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Rocoto Pepper', 'produce', 'lb', 600, 0.90, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Pisco', 'beverage', 'oz', 95, 1.0, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Leche de Tigre Base', 'sauce', 'oz', 40, 1.0, '{fish}'),
  ('11aaaaaa-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Plantain', 'produce', 'lb', 180, 0.80, '{}'),
  ('11aaaaaa-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cotija Cheese', 'dairy', 'lb', 680, 1.0, '{dairy}');

-- ============================================================
-- INGREDIENT LOCATION CONFIGS (Downtown)
-- ============================================================
INSERT INTO ingredient_location_configs (org_id, ingredient_id, location_id, vendor_name, vendor_item_code, local_cost_per_unit, par_level, reorder_point, lead_time_days, avg_daily_usage) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000001', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'SI-AA100', 35, 64, 24, 3, 8.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000002', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'SYS-BT200', 1800, 30, 12, 2, 5.5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000003', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'SYS-CT300', 350, 40, 15, 2, 7.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000004', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'SYS-SB400', 2200, 20, 8, 1, 4.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000005', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'SYS-JS500', 1400, 25, 10, 1, 3.5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000006', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-PP600', 280, 35, 12, 1, 6.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000007', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-SP700', 150, 30, 10, 1, 5.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000008', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-CC800', 85, 50, 20, 1, 10.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000009', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-RO900', 120, 20, 8, 1, 4.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000010', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-LI010', 25, 80, 30, 1, 15.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000011', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-CI011', 90, 15, 5, 1, 3.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000012', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-AV012', 175, 30, 10, 1, 6.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000013', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'SI-QN013', 450, 20, 8, 3, 3.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000014', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'SI-CC014', 320, 15, 5, 3, 2.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000015', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'SI-HU015', 55, 24, 8, 3, 3.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000016', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-RP016', 600, 10, 4, 1, 2.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000017', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'SI-PS017', 95, 128, 48, 5, 16.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000018', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'SI-LT018', 40, 64, 24, 3, 8.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000019', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'LM-PL019', 180, 20, 8, 1, 4.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000020', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'SI-CQ020', 680, 10, 4, 3, 2.0);

-- ============================================================
-- MENU ITEMS - DOWNTOWN (12 items)
-- ============================================================
INSERT INTO menu_items (menu_item_id, org_id, location_id, name, category, price, available, description, source) VALUES
  ('dd000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ceviche Clasico', 'appetizers', 1895, true, 'Fresh sea bass cured in leche de tigre, red onion, cilantro, cancha corn', 'manual'),
  ('dd000001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Tiradito de Corvina', 'appetizers', 2195, true, 'Thinly sliced sea bass, aji amarillo sauce, crispy shallots', 'manual'),
  ('dd000001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Causa Limena', 'appetizers', 1495, true, 'Layered purple potato terrine with avocado and shrimp', 'manual'),
  ('dd000001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Anticuchos de Corazon', 'appetizers', 1695, true, 'Grilled beef heart skewers with aji panca, roasted potatoes', 'manual'),
  ('dd000001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Lomo Saltado', 'mains', 2895, true, 'Stir-fried beef tenderloin with onions, tomatoes, soy, served with fries and rice', 'manual'),
  ('dd000001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Aji de Gallina', 'mains', 2295, true, 'Shredded chicken in creamy aji amarillo walnut sauce, rice, olives', 'manual'),
  ('dd000001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Seco de Cordero', 'mains', 2695, true, 'Slow-braised lamb in cilantro and beer sauce, canario beans', 'manual'),
  ('dd000001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Arroz con Mariscos', 'mains', 3195, true, 'Peruvian seafood rice with shrimp, calamari, mussels, aji amarillo', 'manual'),
  ('dd000001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Pollo a la Brasa', 'mains', 2495, true, 'Rotisserie chicken marinated in aji panca and spices, green sauce', 'manual'),
  ('dd000001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Papas a la Huancaina', 'sides', 995, true, 'Boiled potatoes in creamy aji amarillo cheese sauce', 'manual'),
  ('dd000001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Choclo con Queso', 'sides', 795, true, 'Giant Peruvian corn with fresh cotija cheese', 'manual'),
  ('dd000001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Pisco Sour', 'beverages', 1495, true, 'Classic pisco sour with lime, simple syrup, egg white, Angostura', 'manual');

-- ============================================================
-- MENU ITEMS - AIRPORT (same items, 15-20% higher)
-- ============================================================
INSERT INTO menu_items (menu_item_id, org_id, location_id, name, category, price, available, description, source) VALUES
  ('dd000002-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Ceviche Clasico', 'appetizers', 2195, true, 'Fresh sea bass cured in leche de tigre, red onion, cilantro, cancha corn', 'manual'),
  ('dd000002-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Tiradito de Corvina', 'appetizers', 2495, true, 'Thinly sliced sea bass, aji amarillo sauce, crispy shallots', 'manual'),
  ('dd000002-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Causa Limena', 'appetizers', 1795, true, 'Layered purple potato terrine with avocado and shrimp', 'manual'),
  ('dd000002-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Anticuchos de Corazon', 'appetizers', 1995, true, 'Grilled beef heart skewers with aji panca, roasted potatoes', 'manual'),
  ('dd000002-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Lomo Saltado', 'mains', 3395, true, 'Stir-fried beef tenderloin with onions, tomatoes, soy, served with fries and rice', 'manual'),
  ('dd000002-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Aji de Gallina', 'mains', 2695, true, 'Shredded chicken in creamy aji amarillo walnut sauce, rice, olives', 'manual'),
  ('dd000002-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Seco de Cordero', 'mains', 3095, true, 'Slow-braised lamb in cilantro and beer sauce, canario beans', 'manual'),
  ('dd000002-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Arroz con Mariscos', 'mains', 3695, true, 'Peruvian seafood rice with shrimp, calamari, mussels, aji amarillo', 'manual'),
  ('dd000002-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Pollo a la Brasa', 'mains', 2895, true, 'Rotisserie chicken marinated in aji panca and spices, green sauce', 'manual'),
  ('dd000002-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Papas a la Huancaina', 'sides', 1195, true, 'Boiled potatoes in creamy aji amarillo cheese sauce', 'manual'),
  ('dd000002-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Choclo con Queso', 'sides', 995, true, 'Giant Peruvian corn with fresh cotija cheese', 'manual'),
  ('dd000002-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Pisco Sour', 'beverages', 1795, true, 'Classic pisco sour with lime, simple syrup, egg white, Angostura', 'manual');

-- ============================================================
-- RECIPES + RECIPE INGREDIENTS (Downtown)
-- ============================================================
-- Ceviche Clasico
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000001', 'Ceviche Clasico', 1, 'ea', 15);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa001', '11aaaaaa-0001-0001-0001-000000000004', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa001', '11aaaaaa-0001-0001-0001-000000000018', 3.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa001', '11aaaaaa-0001-0001-0001-000000000009', 0.125, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa001', '11aaaaaa-0001-0001-0001-000000000010', 2, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa001', '11aaaaaa-0001-0001-0001-000000000014', 0.0625, 'lb');

-- Tiradito de Corvina
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000002', 'Tiradito de Corvina', 1, 'ea', 12);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa002', '11aaaaaa-0001-0001-0001-000000000004', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa002', '11aaaaaa-0001-0001-0001-000000000001', 2.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa002', '11aaaaaa-0001-0001-0001-000000000010', 1.5, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa002', '11aaaaaa-0001-0001-0001-000000000011', 0.25, 'bunch');

-- Causa Limena
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', 'Causa Limena', 1, 'ea', 20);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa003', '11aaaaaa-0001-0001-0001-000000000006', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa003', '11aaaaaa-0001-0001-0001-000000000012', 1.0, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa003', '11aaaaaa-0001-0001-0001-000000000005', 0.1875, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa003', '11aaaaaa-0001-0001-0001-000000000001', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa003', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea');

-- Anticuchos de Corazon
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000004', 'Anticuchos de Corazon', 1, 'ea', 25);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa004', '11aaaaaa-0001-0001-0001-000000000002', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa004', '11aaaaaa-0001-0001-0001-000000000001', 1.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa004', '11aaaaaa-0001-0001-0001-000000000006', 0.25, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa004', '11aaaaaa-0001-0001-0001-000000000016', 0.0625, 'lb');

-- Lomo Saltado
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000005', 'Lomo Saltado', 1, 'ea', 18);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa005', '11aaaaaa-0001-0001-0001-000000000002', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa005', '11aaaaaa-0001-0001-0001-000000000009', 0.1875, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa005', '11aaaaaa-0001-0001-0001-000000000001', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa005', '11aaaaaa-0001-0001-0001-000000000006', 0.25, 'lb');

-- Aji de Gallina
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000006', 'Aji de Gallina', 1, 'ea', 30);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa006', '11aaaaaa-0001-0001-0001-000000000003', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa006', '11aaaaaa-0001-0001-0001-000000000001', 2.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa006', '11aaaaaa-0001-0001-0001-000000000006', 0.25, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa006', '11aaaaaa-0001-0001-0001-000000000020', 0.0625, 'lb');

-- Seco de Cordero
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000007', 'Seco de Cordero', 1, 'ea', 90);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa007', '11aaaaaa-0001-0001-0001-000000000002', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa007', '11aaaaaa-0001-0001-0001-000000000011', 0.50, 'bunch'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa007', '11aaaaaa-0001-0001-0001-000000000015', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa007', '11aaaaaa-0001-0001-0001-000000000009', 0.1875, 'lb');

-- Arroz con Mariscos
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', 'Arroz con Mariscos', 1, 'ea', 25);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa008', '11aaaaaa-0001-0001-0001-000000000005', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa008', '11aaaaaa-0001-0001-0001-000000000001', 2.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa008', '11aaaaaa-0001-0001-0001-000000000016', 0.0625, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa008', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa008', '11aaaaaa-0001-0001-0001-000000000011', 0.25, 'bunch');

-- Pollo a la Brasa
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000009', 'Pollo a la Brasa', 1, 'ea', 60);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa009', '11aaaaaa-0001-0001-0001-000000000003', 0.75, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa009', '11aaaaaa-0001-0001-0001-000000000001', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa009', '11aaaaaa-0001-0001-0001-000000000015', 0.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa009', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea');

-- Papas a la Huancaina
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000010', 'Papas a la Huancaina', 1, 'ea', 15);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa010', '11aaaaaa-0001-0001-0001-000000000006', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa010', '11aaaaaa-0001-0001-0001-000000000001', 1.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa010', '11aaaaaa-0001-0001-0001-000000000020', 0.125, 'lb');

-- Choclo con Queso
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000011', 'Choclo con Queso', 1, 'ea', 10);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa011', '11aaaaaa-0001-0001-0001-000000000008', 1.0, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa011', '11aaaaaa-0001-0001-0001-000000000020', 0.125, 'lb');

-- Pisco Sour
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes) VALUES
  ('aa000001-0001-0001-0001-aaaaaaaaa012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000012', 'Pisco Sour', 1, 'ea', 5);
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa012', '11aaaaaa-0001-0001-0001-000000000017', 3.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa000001-0001-0001-0001-aaaaaaaaa012', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea');

-- ============================================================
-- RECIPE EXPLOSION (materialized cost data)
-- ============================================================
INSERT INTO recipe_explosion (org_id, menu_item_id, ingredient_id, quantity_per_unit, unit) VALUES
  -- Ceviche: sea bass 0.375lb, leche 3oz, onion 0.125lb, lime 2ea, cancha 0.0625lb
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000004', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000018', 3.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000009', 0.125, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000010', 2.0, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000014', 0.0625, 'lb'),
  -- Tiradito: sea bass 0.375lb, aji 2oz, lime 1.5ea, cilantro 0.25bunch
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000004', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000001', 2.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000010', 1.5, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000011', 0.25, 'bunch'),
  -- Causa: potato 0.50lb, avo 1ea, shrimp 0.1875lb, aji 1oz, lime 1ea
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000006', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000012', 1.0, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000005', 0.1875, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000001', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea'),
  -- Anticuchos: beef 0.375lb, aji 1.5oz, potato 0.25lb, rocoto 0.0625lb
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000004', '11aaaaaa-0001-0001-0001-000000000002', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000004', '11aaaaaa-0001-0001-0001-000000000001', 1.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000004', '11aaaaaa-0001-0001-0001-000000000006', 0.25, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000004', '11aaaaaa-0001-0001-0001-000000000016', 0.0625, 'lb'),
  -- Lomo Saltado: beef 0.50lb, onion 0.1875lb, aji 1oz, potato 0.25lb
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000005', '11aaaaaa-0001-0001-0001-000000000002', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000005', '11aaaaaa-0001-0001-0001-000000000009', 0.1875, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000005', '11aaaaaa-0001-0001-0001-000000000001', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000005', '11aaaaaa-0001-0001-0001-000000000006', 0.25, 'lb'),
  -- Aji de Gallina: chicken 0.50lb, aji 2.5oz, potato 0.25lb, cheese 0.0625lb
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000006', '11aaaaaa-0001-0001-0001-000000000003', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000006', '11aaaaaa-0001-0001-0001-000000000001', 2.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000006', '11aaaaaa-0001-0001-0001-000000000006', 0.25, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000006', '11aaaaaa-0001-0001-0001-000000000020', 0.0625, 'lb'),
  -- Seco: beef 0.50lb, cilantro 0.50bunch, huacatay 1oz, onion 0.1875lb
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000007', '11aaaaaa-0001-0001-0001-000000000002', 0.50, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000007', '11aaaaaa-0001-0001-0001-000000000011', 0.50, 'bunch'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000007', '11aaaaaa-0001-0001-0001-000000000015', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000007', '11aaaaaa-0001-0001-0001-000000000009', 0.1875, 'lb'),
  -- Arroz con Mariscos: shrimp 0.375lb, aji 2oz, rocoto 0.0625lb, lime 1ea, cilantro 0.25bunch
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', '11aaaaaa-0001-0001-0001-000000000005', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', '11aaaaaa-0001-0001-0001-000000000001', 2.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', '11aaaaaa-0001-0001-0001-000000000016', 0.0625, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', '11aaaaaa-0001-0001-0001-000000000011', 0.25, 'bunch'),
  -- Pollo a la Brasa: chicken 0.75lb, aji 1oz, huacatay 0.5oz, lime 1ea
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000009', '11aaaaaa-0001-0001-0001-000000000003', 0.75, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000009', '11aaaaaa-0001-0001-0001-000000000001', 1.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000009', '11aaaaaa-0001-0001-0001-000000000015', 0.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000009', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea'),
  -- Papas Huancaina: potato 0.375lb, aji 1.5oz, cheese 0.125lb
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000010', '11aaaaaa-0001-0001-0001-000000000006', 0.375, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000010', '11aaaaaa-0001-0001-0001-000000000001', 1.5, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000010', '11aaaaaa-0001-0001-0001-000000000020', 0.125, 'lb'),
  -- Choclo: corn 1ea, cheese 0.125lb
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000011', '11aaaaaa-0001-0001-0001-000000000008', 1.0, 'ea'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000011', '11aaaaaa-0001-0001-0001-000000000020', 0.125, 'lb'),
  -- Pisco Sour: pisco 3oz, lime 1ea
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000012', '11aaaaaa-0001-0001-0001-000000000017', 3.0, 'oz'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000012', '11aaaaaa-0001-0001-0001-000000000010', 1.0, 'ea');

-- ============================================================
-- K. MENU SCORES (pre-calculated)
-- ============================================================
UPDATE menu_items SET margin_score=82, velocity_score=88, complexity_score=30, satisfaction_score=92, strategic_score=85, classification='powerhouse'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000001';
UPDATE menu_items SET margin_score=78, velocity_score=55, complexity_score=25, satisfaction_score=88, strategic_score=72, classification='hidden_gem'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000002';
UPDATE menu_items SET margin_score=70, velocity_score=50, complexity_score=45, satisfaction_score=80, strategic_score=62, classification='workhorse'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000003';
UPDATE menu_items SET margin_score=65, velocity_score=45, complexity_score=50, satisfaction_score=82, strategic_score=60, classification='workhorse'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000004';
UPDATE menu_items SET margin_score=75, velocity_score=92, complexity_score=40, satisfaction_score=95, strategic_score=88, classification='powerhouse'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000005';
UPDATE menu_items SET margin_score=72, velocity_score=60, complexity_score=55, satisfaction_score=78, strategic_score=65, classification='workhorse'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000006';
UPDATE menu_items SET margin_score=80, velocity_score=35, complexity_score=85, satisfaction_score=90, strategic_score=68, classification='complex_star'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000007';
UPDATE menu_items SET margin_score=60, velocity_score=42, complexity_score=65, satisfaction_score=85, strategic_score=58, classification='workhorse'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000008';
UPDATE menu_items SET margin_score=68, velocity_score=72, complexity_score=60, satisfaction_score=88, strategic_score=72, classification='crowd_pleaser'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000009';
UPDATE menu_items SET margin_score=85, velocity_score=65, complexity_score=15, satisfaction_score=75, strategic_score=70, classification='workhorse'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000010';
UPDATE menu_items SET margin_score=88, velocity_score=58, complexity_score=10, satisfaction_score=72, strategic_score=68, classification='hidden_gem'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000011';
UPDATE menu_items SET margin_score=90, velocity_score=85, complexity_score=8, satisfaction_score=92, strategic_score=88, classification='crowd_pleaser'
  WHERE menu_item_id = 'dd000001-0001-0001-0001-000000000012';

-- ============================================================
-- D. GENERATE 30 DAYS OF ORDERS (via generate_series)
-- Downtown: 50-90/day, Airport: 30-50/day
-- ============================================================

-- Create a temp table for Downtown menu items with weights
CREATE TEMP TABLE dt_menu AS
SELECT menu_item_id, name, price, category,
  CASE
    WHEN name = 'Lomo Saltado' THEN 18
    WHEN name = 'Ceviche Clasico' THEN 16
    WHEN name = 'Pisco Sour' THEN 15
    WHEN name = 'Pollo a la Brasa' THEN 12
    WHEN name = 'Aji de Gallina' THEN 10
    WHEN name = 'Papas a la Huancaina' THEN 8
    WHEN name = 'Tiradito de Corvina' THEN 6
    WHEN name = 'Anticuchos de Corazon' THEN 5
    WHEN name = 'Causa Limena' THEN 4
    WHEN name = 'Choclo con Queso' THEN 7
    WHEN name = 'Arroz con Mariscos' THEN 5
    WHEN name = 'Seco de Cordero' THEN 4
  END AS weight
FROM menu_items WHERE location_id = 'a1111111-1111-1111-1111-111111111111';

CREATE TEMP TABLE ap_menu AS
SELECT menu_item_id, name, price, category,
  CASE
    WHEN name = 'Lomo Saltado' THEN 15
    WHEN name = 'Ceviche Clasico' THEN 14
    WHEN name = 'Pisco Sour' THEN 12
    WHEN name = 'Pollo a la Brasa' THEN 14
    WHEN name = 'Aji de Gallina' THEN 10
    WHEN name = 'Papas a la Huancaina' THEN 8
    WHEN name = 'Tiradito de Corvina' THEN 5
    WHEN name = 'Anticuchos de Corazon' THEN 5
    WHEN name = 'Causa Limena' THEN 4
    WHEN name = 'Choclo con Queso' THEN 6
    WHEN name = 'Arroz con Mariscos' THEN 4
    WHEN name = 'Seco de Cordero' THEN 3
  END AS weight
FROM menu_items WHERE location_id = 'b2222222-2222-2222-2222-222222222222';

-- Generate Downtown orders
DO $$
DECLARE
  d DATE;
  day_of_week INT;
  num_orders INT;
  i INT;
  check_uuid UUID;
  item_count INT;
  j INT;
  selected_item RECORD;
  order_time TIMESTAMPTZ;
  hour_rand FLOAT;
  hour_val INT;
  minute_val INT;
  v_subtotal INT;
  v_tax INT;
  v_tip INT;
  channel TEXT;
  channel_rand FLOAT;
  item_subtotal INT;
BEGIN
  FOR d IN SELECT generate_series(CURRENT_DATE - 30, CURRENT_DATE - 1, '1 day'::interval)::date LOOP
    day_of_week := EXTRACT(DOW FROM d);
    -- Fri=5, Sat=6 get more orders
    num_orders := CASE
      WHEN day_of_week IN (5, 6) THEN 75 + floor(random() * 16)::int
      WHEN day_of_week IN (0) THEN 55 + floor(random() * 11)::int
      ELSE 50 + floor(random() * 21)::int
    END;

    FOR i IN 1..num_orders LOOP
      check_uuid := gen_random_uuid();
      item_count := 1 + floor(random() * 3)::int; -- 1-3 items

      -- Time: bimodal lunch/dinner
      hour_rand := random();
      IF hour_rand < 0.45 THEN
        hour_val := 11 + floor(random() * 3)::int; -- 11-13
      ELSIF hour_rand < 0.90 THEN
        hour_val := 18 + floor(random() * 3)::int; -- 18-20
      ELSE
        hour_val := 14 + floor(random() * 4)::int; -- 14-17 slow period
      END IF;
      minute_val := floor(random() * 60)::int;
      order_time := (d || ' ' || lpad(hour_val::text, 2, '0') || ':' || lpad(minute_val::text, 2, '0') || ':00')::timestamptz;

      -- Channel
      channel_rand := random();
      IF channel_rand < 0.60 THEN channel := 'dine_in';
      ELSIF channel_rand < 0.85 THEN channel := 'takeout';
      ELSE channel := 'delivery';
      END IF;

      v_subtotal := 0;

      INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel, subtotal, tax, total, tip, opened_at, closed_at, source)
      VALUES (check_uuid, '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111',
        'DT-' || to_char(d, 'MMDD') || '-' || lpad(i::text, 3, '0'), 'closed', channel, 0, 0, 0, 0,
        order_time, order_time + interval '35 minutes', 'manual');

      FOR j IN 1..item_count LOOP
        SELECT * INTO selected_item FROM dt_menu ORDER BY random() * (1.0 / weight) LIMIT 1;
        item_subtotal := selected_item.price * (CASE WHEN random() < 0.1 THEN 2 ELSE 1 END);
        v_subtotal := v_subtotal + item_subtotal;

        INSERT INTO check_items (org_id, check_id, menu_item_id, name, quantity, unit_price, created_at)
        VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', check_uuid, selected_item.menu_item_id, selected_item.name,
          CASE WHEN random() < 0.1 THEN 2 ELSE 1 END, selected_item.price, order_time);
      END LOOP;

      v_tax := (v_subtotal * 0.0825)::int;
      v_tip := CASE WHEN channel = 'dine_in' THEN (v_subtotal * (0.15 + random() * 0.10))::int ELSE (v_subtotal * random() * 0.05)::int END;

      UPDATE checks SET subtotal = v_subtotal, tax = v_tax, total = v_subtotal + v_tax, tip = v_tip
        WHERE check_id = check_uuid;

      INSERT INTO payments (org_id, check_id, amount, tip, method, status, created_at)
      VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', check_uuid, v_subtotal + v_tax, v_tip,
        CASE WHEN random() < 0.8 THEN 'card' ELSE 'cash' END, 'completed', order_time + interval '35 minutes');
    END LOOP;
  END LOOP;
END $$;

-- Generate Airport orders
DO $$
DECLARE
  d DATE;
  day_of_week INT;
  num_orders INT;
  i INT;
  check_uuid UUID;
  item_count INT;
  j INT;
  selected_item RECORD;
  order_time TIMESTAMPTZ;
  hour_rand FLOAT;
  hour_val INT;
  minute_val INT;
  v_subtotal INT;
  v_tax INT;
  v_tip INT;
  channel TEXT;
  channel_rand FLOAT;
  item_subtotal INT;
BEGIN
  FOR d IN SELECT generate_series(CURRENT_DATE - 30, CURRENT_DATE - 1, '1 day'::interval)::date LOOP
    day_of_week := EXTRACT(DOW FROM d);
    num_orders := 30 + floor(random() * 21)::int;

    FOR i IN 1..num_orders LOOP
      check_uuid := gen_random_uuid();
      item_count := 1 + floor(random() * 2)::int;

      hour_rand := random();
      IF hour_rand < 0.35 THEN
        hour_val := 6 + floor(random() * 4)::int; -- early morning 6-9
      ELSIF hour_rand < 0.65 THEN
        hour_val := 11 + floor(random() * 3)::int; -- 11-13
      ELSIF hour_rand < 0.90 THEN
        hour_val := 17 + floor(random() * 3)::int; -- 17-19
      ELSE
        hour_val := 14 + floor(random() * 3)::int;
      END IF;
      minute_val := floor(random() * 60)::int;
      order_time := (d || ' ' || lpad(hour_val::text, 2, '0') || ':' || lpad(minute_val::text, 2, '0') || ':00')::timestamptz;

      channel_rand := random();
      IF channel_rand < 0.50 THEN channel := 'takeout';
      ELSIF channel_rand < 0.90 THEN channel := 'dine_in';
      ELSE channel := 'delivery';
      END IF;

      v_subtotal := 0;

      INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel, subtotal, tax, total, tip, opened_at, closed_at, source)
      VALUES (check_uuid, '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222',
        'AP-' || to_char(d, 'MMDD') || '-' || lpad(i::text, 3, '0'), 'closed', channel, 0, 0, 0, 0,
        order_time, order_time + interval '25 minutes', 'manual');

      FOR j IN 1..item_count LOOP
        SELECT * INTO selected_item FROM ap_menu ORDER BY random() * (1.0 / weight) LIMIT 1;
        item_subtotal := selected_item.price;
        v_subtotal := v_subtotal + item_subtotal;

        INSERT INTO check_items (org_id, check_id, menu_item_id, name, quantity, unit_price, created_at)
        VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', check_uuid, selected_item.menu_item_id, selected_item.name,
          1, selected_item.price, order_time);
      END LOOP;

      v_tax := (v_subtotal * 0.0825)::int;
      v_tip := CASE WHEN channel = 'dine_in' THEN (v_subtotal * (0.12 + random() * 0.08))::int ELSE 0 END;

      UPDATE checks SET subtotal = v_subtotal, tax = v_tax, total = v_subtotal + v_tax, tip = v_tip
        WHERE check_id = check_uuid;

      INSERT INTO payments (org_id, check_id, amount, tip, method, status, created_at)
      VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', check_uuid, v_subtotal + v_tax, v_tip,
        CASE WHEN random() < 0.9 THEN 'card' ELSE 'cash' END, 'completed', order_time + interval '25 minutes');
    END LOOP;
  END LOOP;
END $$;

DROP TABLE dt_menu;
DROP TABLE ap_menu;

-- ============================================================
-- E. INVENTORY DATA
-- ============================================================

-- 3 inventory counts over past 3 weeks
INSERT INTO inventory_counts (count_id, org_id, location_id, counted_by, count_type, status, started_at, submitted_at, approved_by, approved_at) VALUES
  ('cc000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ee111111-4444-4444-4444-444444444444', 'full', 'approved', CURRENT_DATE - 21, CURRENT_DATE - 21 + interval '2 hours', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 21 + interval '3 hours'),
  ('cc000001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ee111111-4444-4444-4444-444444444444', 'full', 'approved', CURRENT_DATE - 14, CURRENT_DATE - 14 + interval '2 hours', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14 + interval '3 hours'),
  ('cc000001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ee111111-4444-4444-4444-444444444444', 'full', 'approved', CURRENT_DATE - 7, CURRENT_DATE - 7 + interval '2 hours', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7 + interval '3 hours');

-- Count lines for each count (key ingredients)
INSERT INTO inventory_count_lines (org_id, count_id, location_id, ingredient_id, expected_qty, counted_qty, unit) VALUES
  -- Count 1 (3 weeks ago)
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000001', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 22.0, 20.5, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000001', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 15.0, 13.8, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000001', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000005', 18.0, 17.2, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000001', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000003', 28.0, 27.0, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000001', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000017', 90.0, 88.5, 'oz'),
  -- Count 2 (2 weeks ago)
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000002', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 25.0, 23.0, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000002', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 16.0, 15.0, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000002', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000005', 20.0, 18.5, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000002', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000003', 30.0, 29.0, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000002', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000017', 95.0, 92.0, 'oz'),
  -- Count 3 (1 week ago)
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000003', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 24.0, 21.5, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000003', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 18.0, 16.2, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000003', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000005', 22.0, 20.0, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000003', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000003', 32.0, 30.5, 'lb'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000001-0001-0001-0001-000000000003', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000017', 100.0, 96.0, 'oz');

-- Waste logs (20 entries over 30 days)
INSERT INTO waste_logs (org_id, location_id, ingredient_id, quantity, unit, reason, logged_by, logged_at, note) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 1.5, 'lb', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 28, 'Sea bass past prime'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000005', 0.75, 'lb', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 26, 'Shrimp discolored'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 0.50, 'lb', 'overcooked', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 25, 'Overcooked lomo saltado batch'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000012', 3.0, 'ea', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 24, 'Avocados overripe'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 0.80, 'lb', 'dropped', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 22, 'Dropped sea bass fillet'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000011', 2.0, 'bunch', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 21, 'Wilted cilantro'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000003', 0.75, 'lb', 'overcooked', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 19, 'Dry chicken, overcooked'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 0.25, 'lb', 'dropped', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 18, 'Beef fell off cutting board'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000005', 1.0, 'lb', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 16, 'Shrimp past 2-day mark'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000018', 8.0, 'oz', 'overproduction', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 15, 'Over-prepped leche de tigre'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 1.0, 'lb', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 13, 'Sunday leftover sea bass'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000012', 2.0, 'ea', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 11, 'Brown avocados'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 0.375, 'lb', 'overcooked', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 10, 'Anticuchos burnt'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000015', 3.0, 'oz', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 9, 'Huacatay past freshness'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000005', 0.50, 'lb', 'dropped', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 7, 'Shrimp container dropped'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000009', 0.50, 'lb', 'contaminated', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 6, 'Cross-contamination concern'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 0.60, 'lb', 'overcooked', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 4, 'Tiradito cut too thick, re-did'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000003', 0.50, 'lb', 'overcooked', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 3, 'Chicken dried out'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000010', 5.0, 'ea', 'expired', 'ee111111-4444-4444-4444-444444444444', CURRENT_DATE - 2, 'Limes dried up'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 0.30, 'lb', 'dropped', 'ee111111-3333-3333-3333-333333333333', CURRENT_DATE - 1, 'Beef trimmings fell');

-- Inventory variances
INSERT INTO inventory_variances (org_id, location_id, ingredient_id, count_id, period_start, period_end, theoretical_usage, actual_usage, variance_qty, variance_cents, cause_probabilities, severity) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000002', 'cc000001-0001-0001-0001-000000000003', CURRENT_DATE - 14, CURRENT_DATE - 7, 35.0, 37.5, -2.5, -4500, '{"waste": 0.4, "portioning": 0.3, "theft": 0.1, "counting_error": 0.2}', 'warning'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000004', 'cc000001-0001-0001-0001-000000000003', CURRENT_DATE - 14, CURRENT_DATE - 7, 28.0, 29.8, -1.8, -3960, '{"waste": 0.5, "portioning": 0.25, "counting_error": 0.25}', 'warning'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000005', 'cc000001-0001-0001-0001-000000000003', CURRENT_DATE - 14, CURRENT_DATE - 7, 22.0, 24.0, -2.0, -2800, '{"waste": 0.6, "portioning": 0.2, "counting_error": 0.2}', 'critical'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000003', 'cc000001-0001-0001-0001-000000000003', CURRENT_DATE - 14, CURRENT_DATE - 7, 45.0, 46.5, -1.5, -525, '{"portioning": 0.5, "waste": 0.3, "counting_error": 0.2}', 'info'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '11aaaaaa-0001-0001-0001-000000000017', 'cc000001-0001-0001-0001-000000000003', CURRENT_DATE - 14, CURRENT_DATE - 7, 95.0, 99.0, -4.0, -380, '{"overpouring": 0.6, "waste": 0.2, "counting_error": 0.2}', 'warning');

-- ============================================================
-- F. PURCHASE ORDERS
-- ============================================================

-- PO1: Received from Sysco (short delivery)
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, received_by, received_at, total_estimated, total_actual, notes) VALUES
  ('bb000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 12, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 10, 62000, 58500, 'Weekly protein order');
INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost, received_qty, received_unit_cost, variance_qty, variance_flag, received_at) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000002', 15.0, 'lb', 1800, 14.0, 1850, -1.0, 'short', CURRENT_DATE - 10),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000004', 10.0, 'lb', 2200, 10.0, 2200, 0, 'exact', CURRENT_DATE - 10),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000005', 12.0, 'lb', 1400, 11.0, 1400, -1.0, 'short', CURRENT_DATE - 10),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000001', '11aaaaaa-0001-0001-0001-000000000003', 20.0, 'lb', 350, 20.0, 350, 0, 'exact', CURRENT_DATE - 10);

-- PO2: Received from Specialty Imports (all good)
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, received_by, received_at, total_estimated, total_actual) VALUES
  ('bb000001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 10, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7, 18500, 18500);
INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost, received_qty, received_unit_cost, variance_qty, variance_flag, received_at) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000001', 32.0, 'oz', 35, 32.0, 35, 0, 'exact', CURRENT_DATE - 7),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000013', 10.0, 'lb', 450, 10.0, 450, 0, 'exact', CURRENT_DATE - 7),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000017', 64.0, 'oz', 95, 64.0, 95, 0, 'exact', CURRENT_DATE - 7),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000002', '11aaaaaa-0001-0001-0001-000000000015', 16.0, 'oz', 55, 16.0, 55, 0, 'exact', CURRENT_DATE - 7);

-- PO3: Approved, awaiting delivery from Local Market
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, total_estimated) VALUES
  ('bb000001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 'approved', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1, 8200);
INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000009', 10.0, 'lb', 120),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000010', 40.0, 'ea', 25),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000011', 8.0, 'bunch', 90),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000003', '11aaaaaa-0001-0001-0001-000000000012', 15.0, 'ea', 175);

-- PO4: Draft system_recommended
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, suggested_at, total_estimated, notes) VALUES
  ('bb000001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'draft', 'system_recommended', now(), 55000, 'AI-recommended: protein levels approaching reorder point');
INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000004', '11aaaaaa-0001-0001-0001-000000000002', 12.0, 'lb', 1800),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000004', '11aaaaaa-0001-0001-0001-000000000004', 8.0, 'lb', 2200),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'bb000001-0001-0001-0001-000000000004', '11aaaaaa-0001-0001-0001-000000000005', 10.0, 'lb', 1400);

-- PO5: Cancelled
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, total_estimated, notes, created_at) VALUES
  ('bb000001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 'cancelled', 'manual', 42000, 'Duplicate order, cancelled', CURRENT_DATE - 15);

-- ============================================================
-- G. FINANCIAL DATA (Budgets)
-- ============================================================
INSERT INTO budgets (org_id, location_id, period_type, period_start, period_end, revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target, created_by) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'monthly', '2026-03-01', '2026-03-31', 8500000, 32.00, 26.00, 2720000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'weekly', '2026-03-16', '2026-03-22', 2100000, 32.00, 26.00, 672000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'monthly', '2026-03-01', '2026-03-31', 5500000, 30.00, 24.00, 1650000, '0d55e810-1e4a-417a-8a70-08b98f4595c2');

-- ============================================================
-- H. LABOR DATA - Shifts (30 days, 5 employees DT, 5 shifts/week each)
-- ============================================================
DO $$
DECLARE
  d DATE;
  dow INT;
  emp_ids UUID[] := ARRAY[
    'ee111111-1111-1111-1111-111111111111',
    'ee111111-2222-2222-2222-222222222222',
    'ee111111-3333-3333-3333-333333333333',
    'ee111111-4444-4444-4444-444444444444',
    'ee111111-5555-5555-5555-555555555555'
  ];
  emp UUID;
  shift_start TIMESTAMPTZ;
  shift_end TIMESTAMPTZ;
  rate INT;
  idx INT;
BEGIN
  FOR d IN SELECT generate_series(CURRENT_DATE - 30, CURRENT_DATE - 1, '1 day'::interval)::date LOOP
    dow := EXTRACT(DOW FROM d);
    FOREACH emp IN ARRAY emp_ids LOOP
      idx := array_position(emp_ids, emp);
      -- Each employee gets ~5 shifts/week; skip 2 days based on position
      IF (dow + idx) % 7 < 5 THEN
        -- AM or PM shift
        IF idx <= 3 THEN
          shift_start := (d || ' 08:00:00')::timestamptz;
          shift_end := (d || ' 16:00:00')::timestamptz;
        ELSE
          shift_start := (d || ' 14:00:00')::timestamptz;
          shift_end := (d || ' 22:00:00')::timestamptz;
        END IF;

        rate := CASE idx
          WHEN 1 THEN 2800
          WHEN 2 THEN 2200
          WHEN 3 THEN 1800
          WHEN 4 THEN 1600
          WHEN 5 THEN 1500
        END;

        INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status)
        VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', emp,
          CASE idx WHEN 1 THEN 'gm' WHEN 2 THEN 'shift_manager' ELSE 'staff' END,
          shift_start, shift_end, rate, 'completed');
      END IF;
    END LOOP;
  END LOOP;
END $$;

-- Airport shifts
DO $$
DECLARE
  d DATE;
  dow INT;
  emp_ids UUID[] := ARRAY[
    'ee222222-1111-1111-1111-111111111111',
    'ee222222-2222-2222-2222-222222222222',
    'ee222222-3333-3333-3333-333333333333'
  ];
  emp UUID;
  shift_start TIMESTAMPTZ;
  shift_end TIMESTAMPTZ;
  rate INT;
  idx INT;
BEGIN
  FOR d IN SELECT generate_series(CURRENT_DATE - 30, CURRENT_DATE - 1, '1 day'::interval)::date LOOP
    dow := EXTRACT(DOW FROM d);
    FOREACH emp IN ARRAY emp_ids LOOP
      idx := array_position(emp_ids, emp);
      IF (dow + idx) % 7 < 5 THEN
        shift_start := (d || ' 07:00:00')::timestamptz;
        shift_end := (d || ' 15:00:00')::timestamptz;
        rate := CASE idx WHEN 1 THEN 2600 WHEN 2 THEN 2000 WHEN 3 THEN 1700 END;

        INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status)
        VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', emp,
          CASE idx WHEN 1 THEN 'gm' WHEN 2 THEN 'shift_manager' ELSE 'staff' END,
          shift_start, shift_end, rate, 'completed');
      END IF;
    END LOOP;
  END LOOP;
END $$;

-- Staff point events (40+ events)
INSERT INTO staff_point_events (org_id, employee_id, points, reason, description, awarded_by, created_at) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 15, 'task_completion', 'Weekly inventory count completed on time', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 28),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 10, 'speed_bonus', 'Fastest ticket times during Friday rush', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 27),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 12, 'accuracy_bonus', 'Zero send-backs on ceviche station', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 26),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 8, 'task_completion', 'All prep completed before service', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 25),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 15, 'peer_nominated', 'Team nominated for guest experience excellence', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 24),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 10, 'attendance', 'Perfect attendance this week', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 22),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', -5, 'late', '15 minutes late to shift', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 21),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 12, 'task_completion', 'Trained new prep cook on ceviche station', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 20),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 10, 'speed_bonus', 'Prep speed improved 20% this week', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 19),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 8, 'accuracy_bonus', 'No order errors during Saturday service', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 18),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 12, 'task_completion', 'Schedule optimization saved 8 labor hours', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 17),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 15, 'speed_bonus', 'Fastest saute ticket avg under 8 min', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 16),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 10, 'attendance', 'Perfect attendance this week', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 15),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', -3, 'incomplete_task', 'Prep list not fully completed', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 10, 'task_completion', 'Upsold 15 Pisco Sours during happy hour', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 13),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 8, 'manager_adjustment', 'Handled vendor issue proactively', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 12),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 10, 'task_completion', 'Developed new ceviche prep method', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 11),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 8, 'speed_bonus', 'Expo station running smoothly', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 10),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 12, 'accuracy_bonus', 'Perfect portioning on inventory check', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 9),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 12, 'peer_nominated', 'Best server of the week', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 8),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 15, 'task_completion', 'Monthly P&L review completed early', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 8, 'attendance', 'On time every day this week', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 6),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 15, 'task_completion', 'Reduced waste 12% through better rotation', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 10, 'speed_bonus', 'Prep work done 30 min early', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 4),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', -5, 'late', '10 minutes late', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 10, 'task_completion', 'Weekly team meeting well-organized', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 12, 'accuracy_bonus', 'Perfect ceviche consistency all week', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1),
  -- Airport staff events
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-1111-1111-1111-111111111111', 12, 'task_completion', 'Smooth opening every day', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 20),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-2222-2222-2222-222222222222', 10, 'speed_bonus', 'Fast turnover during rush', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 15),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-3333-3333-3333-333333333333', 8, 'accuracy_bonus', 'No mistakes on takeout orders', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 10),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-1111-1111-1111-111111111111', 10, 'attendance', 'Perfect attendance month', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-2222-2222-2222-222222222222', 12, 'task_completion', 'Trained new hire effectively', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 3),
  -- More Downtown events for 40+ total
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 10, 'accuracy_bonus', 'Food cost under budget this week', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 29),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', 8, 'task_completion', 'Deep clean completed ahead of schedule', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 23),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 15, 'peer_nominated', 'Helped coworker learn grill station', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 17),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-5555-5555-5555-555555555555', 10, 'task_completion', 'Table turnover improved by 10%', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 11),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-3333-3333-3333-333333333333', 8, 'speed_bonus', 'Covered extra station during call-out', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 8),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-1111-1111-1111-111111111111', 12, 'peer_nominated', 'Team morale consistently high', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 4),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-2222-2222-2222-222222222222', -8, 'no_show', 'Missed Sunday shift', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee111111-4444-4444-4444-444444444444', 10, 'task_completion', 'Mise en place perfect for dinner service', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ee222222-3333-3333-3333-333333333333', 10, 'speed_bonus', 'Handled solo rush effectively', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1);

-- ============================================================
-- I. GUEST PROFILES (25 guests)
-- ============================================================
INSERT INTO guest_profiles (guest_id, org_id, payment_token_hash, privacy_tier, first_name, email, phone, total_visits, total_spend, avg_check, preferred_channel, favorite_items, clv_score, segment, churn_risk, churn_probability, next_visit_predicted, last_visit_at) VALUES
  -- Champions (4)
  ('cc000002-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_champ_01', 3, 'Alejandro', 'alejandro.ruiz@email.com', '512-555-0101', 24, 192000, 8000, 'dine_in', '["Lomo Saltado", "Pisco Sour"]', 4800.00, 'champion', 'low', 0.05, CURRENT_DATE + 3, CURRENT_DATE - 2),
  ('cc000002-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_champ_02', 3, 'Patricia', 'patricia.vega@email.com', '512-555-0102', 20, 180000, 9000, 'dine_in', '["Ceviche Clasico", "Tiradito de Corvina"]', 4500.00, 'champion', 'low', 0.08, CURRENT_DATE + 4, CURRENT_DATE - 3),
  ('cc000002-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_champ_03', 3, 'Roberto', 'roberto.luna@email.com', '512-555-0103', 18, 162000, 9000, 'dine_in', '["Seco de Cordero", "Arroz con Mariscos"]', 4050.00, 'champion', 'low', 0.10, CURRENT_DATE + 5, CURRENT_DATE - 4),
  ('cc000002-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_champ_04', 2, 'Carmen', NULL, '512-555-0104', 16, 128000, 8000, 'dine_in', '["Pollo a la Brasa", "Pisco Sour"]', 3200.00, 'champion', 'low', 0.12, CURRENT_DATE + 6, CURRENT_DATE - 5),
  -- Loyal Regulars (6)
  ('cc000002-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_loyal_01', 2, 'Miguel', NULL, '512-555-0105', 10, 65000, 6500, 'dine_in', '["Aji de Gallina"]', 1625.00, 'regular', 'low', 0.15, CURRENT_DATE + 7, CURRENT_DATE - 6),
  ('cc000002-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_loyal_02', 2, 'Isabella', NULL, '512-555-0106', 9, 72000, 8000, 'takeout', '["Ceviche Clasico", "Causa Limena"]', 1800.00, 'regular', 'low', 0.18, CURRENT_DATE + 8, CURRENT_DATE - 5),
  ('cc000002-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_loyal_03', 2, 'Fernando', NULL, '512-555-0107', 8, 56000, 7000, 'dine_in', '["Lomo Saltado"]', 1400.00, 'regular', 'medium', 0.22, CURRENT_DATE + 10, CURRENT_DATE - 8),
  ('cc000002-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_loyal_04', 1, 'Lucia', NULL, NULL, 7, 49000, 7000, 'dine_in', '["Arroz con Mariscos"]', 1225.00, 'regular', 'low', 0.14, CURRENT_DATE + 9, CURRENT_DATE - 7),
  ('cc000002-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_loyal_05', 1, 'Eduardo', NULL, NULL, 6, 42000, 7000, 'delivery', '["Pollo a la Brasa"]', 1050.00, 'regular', 'medium', 0.25, CURRENT_DATE + 12, CURRENT_DATE - 10),
  ('cc000002-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_loyal_06', 2, 'Diana', NULL, '512-555-0110', 8, 64000, 8000, 'dine_in', '["Tiradito de Corvina", "Pisco Sour"]', 1600.00, 'regular', 'low', 0.12, CURRENT_DATE + 7, CURRENT_DATE - 4),
  -- At Risk (5)
  ('cc000002-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_risk_01', 2, 'Santiago', NULL, '512-555-0111', 8, 56000, 7000, 'dine_in', '["Lomo Saltado"]', 1400.00, 'at_risk', 'high', 0.72, CURRENT_DATE - 5, CURRENT_DATE - 25),
  ('cc000002-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_risk_02', 2, 'Valentina', NULL, '512-555-0112', 6, 54000, 9000, 'dine_in', '["Arroz con Mariscos", "Ceviche Clasico"]', 1350.00, 'at_risk', 'high', 0.68, CURRENT_DATE - 3, CURRENT_DATE - 22),
  ('cc000002-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_risk_03', 1, 'Andres', NULL, NULL, 5, 35000, 7000, 'takeout', '["Pollo a la Brasa"]', 875.00, 'at_risk', 'critical', 0.85, CURRENT_DATE - 10, CURRENT_DATE - 35),
  ('cc000002-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_risk_04', 1, 'Gabriela', NULL, NULL, 7, 49000, 7000, 'dine_in', '["Aji de Gallina"]', 1225.00, 'at_risk', 'high', 0.70, CURRENT_DATE - 7, CURRENT_DATE - 28),
  ('cc000002-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_risk_05', 2, 'Rafael', NULL, '512-555-0115', 9, 72000, 8000, 'dine_in', '["Seco de Cordero"]', 1800.00, 'at_risk', 'high', 0.65, CURRENT_DATE - 4, CURRENT_DATE - 21),
  -- New Discoverers (4)
  ('cc000002-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_new_01', 1, 'Emma', NULL, NULL, 2, 8500, 4250, 'dine_in', '["Ceviche Clasico"]', 212.50, 'new', 'medium', 0.45, CURRENT_DATE + 14, CURRENT_DATE - 5),
  ('cc000002-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_new_02', 1, 'Austin', NULL, NULL, 1, 4800, 4800, 'takeout', '["Lomo Saltado"]', 120.00, 'new', 'medium', 0.50, CURRENT_DATE + 21, CURRENT_DATE - 3),
  ('cc000002-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_new_03', 1, 'Sophia', NULL, NULL, 1, 5200, 5200, 'dine_in', '["Tiradito de Corvina"]', 130.00, 'new', 'medium', 0.48, CURRENT_DATE + 18, CURRENT_DATE - 7),
  ('cc000002-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_new_04', 1, 'Mateo', NULL, NULL, 2, 7600, 3800, 'delivery', '["Pollo a la Brasa"]', 190.00, 'new', 'medium', 0.42, CURRENT_DATE + 12, CURRENT_DATE - 4),
  -- Casual (6)
  ('cc000002-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_cas_01', 1, 'James', NULL, NULL, 4, 24000, 6000, 'dine_in', '[]', 600.00, 'casual', 'medium', 0.35, CURRENT_DATE + 20, CURRENT_DATE - 15),
  ('cc000002-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_cas_02', 1, 'Maria', NULL, NULL, 3, 21000, 7000, 'takeout', '["Ceviche Clasico"]', 525.00, 'casual', 'medium', 0.38, CURRENT_DATE + 25, CURRENT_DATE - 18),
  ('cc000002-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_cas_03', 1, 'David', NULL, NULL, 3, 18000, 6000, 'dine_in', '[]', 450.00, 'casual', 'low', 0.30, CURRENT_DATE + 22, CURRENT_DATE - 12),
  ('cc000002-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_cas_04', 1, 'Ana', NULL, NULL, 4, 28000, 7000, 'dine_in', '["Pisco Sour"]', 700.00, 'casual', 'medium', 0.32, CURRENT_DATE + 18, CURRENT_DATE - 14),
  ('cc000002-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_cas_05', 1, 'Carlos', NULL, NULL, 3, 15000, 5000, 'delivery', '[]', 375.00, 'casual', 'medium', 0.40, CURRENT_DATE + 30, CURRENT_DATE - 20),
  ('cc000002-0001-0001-0001-000000000025', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'hash_cas_06', 1, 'Nicole', NULL, NULL, 5, 35000, 7000, 'dine_in', '["Causa Limena"]', 875.00, 'casual', 'low', 0.28, CURRENT_DATE + 15, CURRENT_DATE - 11);

-- Guest visits (link to recent checks for champions/regulars)
DO $$
DECLARE
  guest RECORD;
  v INT;
  visit_date TIMESTAMPTZ;
  spend_amt BIGINT;
BEGIN
  FOR guest IN SELECT guest_id, total_visits, avg_check, segment, preferred_channel, last_visit_at FROM guest_profiles WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec' LOOP
    FOR v IN 1..guest.total_visits LOOP
      visit_date := guest.last_visit_at - ((v - 1) * (CASE
        WHEN guest.segment = 'champion' THEN 3
        WHEN guest.segment = 'regular' THEN 7
        WHEN guest.segment = 'at_risk' THEN 5
        WHEN guest.segment = 'new' THEN 14
        ELSE 10
      END))::int * interval '1 day';

      -- Only create visits within 90 days
      IF visit_date > CURRENT_DATE - 90 THEN
        spend_amt := guest.avg_check + (random() * 2000 - 1000)::bigint;
        INSERT INTO guest_visits (org_id, guest_id, location_id, channel, spend, item_count, party_size, visited_at)
        VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', guest.guest_id, 'a1111111-1111-1111-1111-111111111111',
          COALESCE(guest.preferred_channel, 'dine_in'), spend_amt,
          1 + floor(random() * 3)::int,
          CASE WHEN guest.preferred_channel = 'dine_in' THEN 1 + floor(random() * 3)::int ELSE 1 END,
          visit_date);
      END IF;
    END LOOP;
  END LOOP;
END $$;

-- ============================================================
-- J. KITCHEN DATA
-- ============================================================

-- 6 stations
INSERT INTO kitchen_stations (station_id, org_id, location_id, name, station_type, max_concurrent) VALUES
  ('ab000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ceviche Bar', 'ceviche_bar', 4),
  ('ab000001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Grill', 'grill', 6),
  ('ab000001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Saute', 'saute', 5),
  ('ab000001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Prep', 'prep', 8),
  ('ab000001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Expo', 'expo', 4),
  ('ab000001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Plancha', 'plancha', 4);

-- Resource profiles for menu items
INSERT INTO menu_item_resource_profiles (org_id, menu_item_id, station_type, task_sequence, duration_secs, elu_required, batch_size) VALUES
  -- Ceviche Clasico: ceviche_bar only
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000001', 'ceviche_bar', 1, 300, 1.0, 1),
  -- Tiradito: ceviche_bar
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000002', 'ceviche_bar', 1, 240, 1.2, 1),
  -- Causa: prep + ceviche_bar
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', 'prep', 1, 600, 1.0, 1),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000003', 'ceviche_bar', 2, 180, 0.5, 1),
  -- Anticuchos: grill
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000004', 'grill', 1, 480, 1.0, 1),
  -- Lomo Saltado: saute
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000005', 'saute', 1, 420, 1.5, 1),
  -- Aji de Gallina: saute + prep
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000006', 'prep', 1, 900, 1.0, 4),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000006', 'saute', 2, 300, 1.0, 1),
  -- Seco: saute (slow braise, batched)
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000007', 'saute', 1, 1800, 0.8, 6),
  -- Arroz con Mariscos: plancha + saute
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', 'plancha', 1, 360, 1.2, 1),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000008', 'saute', 2, 480, 1.0, 1),
  -- Pollo a la Brasa: grill (rotisserie)
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000009', 'grill', 1, 2400, 0.5, 8),
  -- Papas: prep
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000010', 'prep', 1, 300, 0.5, 1),
  -- Choclo: prep
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000011', 'prep', 1, 180, 0.3, 1),
  -- Pisco Sour: (no kitchen station, bar item)
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'dd000001-0001-0001-0001-000000000012', 'expo', 1, 120, 0.5, 1);

-- 5 active KDS tickets
INSERT INTO kds_tickets (ticket_id, org_id, location_id, order_number, channel, status, priority, estimated_ready_at, created_at) VALUES
  ('ac000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'DT-LIVE-001', 'dine_in', 'new', 0, now() + interval '15 minutes', now() - interval '2 minutes'),
  ('ac000001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'DT-LIVE-002', 'dine_in', 'in_progress', 0, now() + interval '8 minutes', now() - interval '7 minutes'),
  ('ac000001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'DT-LIVE-003', 'takeout', 'in_progress', 1, now() + interval '5 minutes', now() - interval '10 minutes'),
  ('ac000001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'DT-LIVE-004', 'delivery', 'new', 2, now() + interval '20 minutes', now() - interval '1 minute'),
  ('ac000001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'DT-LIVE-005', 'dine_in', 'ready', 0, now() - interval '2 minutes', now() - interval '15 minutes');

INSERT INTO kds_ticket_items (org_id, ticket_id, menu_item_id, item_name, quantity, station_type, status, fire_at) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000001', 'dd000001-0001-0001-0001-000000000001', 'Ceviche Clasico', 1, 'ceviche_bar', 'pending', now()),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000001', 'dd000001-0001-0001-0001-000000000005', 'Lomo Saltado', 1, 'saute', 'pending', now()),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000002', 'dd000001-0001-0001-0001-000000000005', 'Lomo Saltado', 2, 'saute', 'cooking', now() - interval '5 minutes'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000002', 'dd000001-0001-0001-0001-000000000012', 'Pisco Sour', 2, 'expo', 'ready', now() - interval '6 minutes'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000003', 'dd000001-0001-0001-0001-000000000009', 'Pollo a la Brasa', 1, 'grill', 'cooking', now() - interval '8 minutes'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000003', 'dd000001-0001-0001-0001-000000000010', 'Papas a la Huancaina', 1, 'prep', 'ready', now() - interval '9 minutes'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000004', 'dd000001-0001-0001-0001-000000000008', 'Arroz con Mariscos', 1, 'plancha', 'pending', now()),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000004', 'dd000001-0001-0001-0001-000000000002', 'Tiradito de Corvina', 1, 'ceviche_bar', 'pending', now()),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000005', 'dd000001-0001-0001-0001-000000000006', 'Aji de Gallina', 1, 'saute', 'ready', now() - interval '12 minutes'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac000001-0001-0001-0001-000000000005', 'dd000001-0001-0001-0001-000000000011', 'Choclo con Queso', 1, 'prep', 'ready', now() - interval '10 minutes');

-- ============================================================
-- L. VENDOR SCORES
-- ============================================================
INSERT INTO vendor_scores (org_id, location_id, vendor_name, overall_score, price_score, delivery_score, quality_score, accuracy_score, total_orders, otif_rate, on_time_rate, in_full_rate, avg_lead_days) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco', 72.50, 65.00, 78.00, 80.00, 70.00, 24, 68.00, 82.00, 75.00, 2.10),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Local Market', 85.00, 80.00, 90.00, 88.00, 85.00, 18, 88.00, 92.00, 90.00, 1.00),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 78.00, 70.00, 72.00, 90.00, 82.00, 12, 75.00, 78.00, 85.00, 3.50);

-- ============================================================
-- M. PRICE HISTORY (6 months, key proteins)
-- ============================================================
INSERT INTO ingredient_price_history (org_id, ingredient_id, vendor_name, unit_cost, quantity, source, recorded_at) VALUES
  -- Beef Tenderloin trending up
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000002', 'Sysco', 1650, 15.0, 'po_received', '2025-10-01'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000002', 'Sysco', 1680, 12.0, 'po_received', '2025-11-01'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000002', 'Sysco', 1720, 15.0, 'po_received', '2025-12-01'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000002', 'Sysco', 1750, 14.0, 'po_received', '2026-01-15'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000002', 'Sysco', 1780, 15.0, 'po_received', '2026-02-15'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000002', 'Sysco', 1850, 14.0, 'po_received', CURRENT_DATE - 10),
  -- Sea Bass seasonal
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000004', 'Sysco', 1950, 10.0, 'po_received', '2025-10-01'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000004', 'Sysco', 2100, 8.0, 'po_received', '2025-11-15'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000004', 'Sysco', 2350, 8.0, 'po_received', '2025-12-20'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000004', 'Sysco', 2250, 10.0, 'po_received', '2026-01-15'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000004', 'Sysco', 2200, 10.0, 'po_received', '2026-02-15'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000004', 'Sysco', 2200, 10.0, 'po_received', CURRENT_DATE - 10),
  -- Shrimp
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000005', 'Sysco', 1350, 12.0, 'po_received', '2025-10-15'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000005', 'Sysco', 1380, 10.0, 'po_received', '2025-12-01'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000005', 'Sysco', 1420, 12.0, 'po_received', '2026-01-20'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000005', 'Sysco', 1400, 11.0, 'po_received', CURRENT_DATE - 10),
  -- Chicken stable
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000003', 'Sysco', 340, 20.0, 'po_received', '2025-10-01'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000003', 'Sysco', 345, 18.0, 'po_received', '2025-12-15'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000003', 'Sysco', 350, 20.0, 'po_received', '2026-02-01'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '11aaaaaa-0001-0001-0001-000000000003', 'Sysco', 350, 20.0, 'po_received', CURRENT_DATE - 10);

-- ============================================================
-- N. MARKETING CAMPAIGNS
-- ============================================================
INSERT INTO campaigns (org_id, location_id, name, campaign_type, status, target_segment, channel, discount_type, discount_value, start_at, end_at, recurring, recurrence_rule, redemptions, revenue_attributed, cost_of_promotion, created_by) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Pisco Hour', 'happy_hour', 'active', 'all', 'in_app', 'percentage', 25.00, CURRENT_DATE - 14, CURRENT_DATE + 30, true, 'FREQ=WEEKLY;BYDAY=TU,WE,TH;BYHOUR=16,17,18', 145, 1450000, 362500, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Weekend Ceviche Festival', 'bundle', 'completed', 'regular', 'email', 'bundle', 15.00, CURRENT_DATE - 30, CURRENT_DATE - 16, false, NULL, 89, 890000, 133500, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', NULL, 'Loyalty Perks', 'loyalty_reward', 'draft', 'champion', 'all', 'dollar_off', 10.00, NULL, NULL, false, NULL, 0, 0, 0, '0d55e810-1e4a-417a-8a70-08b98f4595c2');

-- Loyalty members (8 from champions/regulars)
INSERT INTO loyalty_members (member_id, org_id, guest_id, points_balance, lifetime_points, tier, joined_at) VALUES
  ('ad000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000001', 480, 1920, 'platinum', CURRENT_DATE - 180),
  ('ad000001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000002', 350, 1800, 'gold', CURRENT_DATE - 150),
  ('ad000001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000003', 280, 1620, 'gold', CURRENT_DATE - 120),
  ('ad000001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000004', 200, 1280, 'silver', CURRENT_DATE - 90),
  ('ad000001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000005', 120, 650, 'silver', CURRENT_DATE - 75),
  ('ad000001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000006', 150, 720, 'silver', CURRENT_DATE - 60),
  ('ad000001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000010', 100, 640, 'bronze', CURRENT_DATE - 45),
  ('ad000001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'cc000002-0001-0001-0001-000000000015', 90, 720, 'silver', CURRENT_DATE - 80);

-- Loyalty transactions
INSERT INTO loyalty_transactions (org_id, member_id, type, points, description, created_at) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000001', 'earn', 80, 'Dinner visit - 80 pts', CURRENT_DATE - 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000001', 'redeem', -100, 'Free Pisco Sour redemption', CURRENT_DATE - 12),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000002', 'earn', 90, 'Dinner visit - 90 pts', CURRENT_DATE - 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000003', 'earn', 70, 'Lunch visit - 70 pts', CURRENT_DATE - 7),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000004', 'earn', 65, 'Dinner visit - 65 pts', CURRENT_DATE - 8),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000005', 'earn', 55, 'Visit - 55 pts', CURRENT_DATE - 10),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000006', 'redeem', -50, 'Appetizer discount', CURRENT_DATE - 6),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000006', 'earn', 80, 'Dinner visit - 80 pts', CURRENT_DATE - 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000007', 'earn', 60, 'Visit - 60 pts', CURRENT_DATE - 4),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ad000001-0001-0001-0001-000000000008', 'earn', 75, 'Visit - 75 pts', CURRENT_DATE - 22);

-- ============================================================
-- O. PORTFOLIO + BENCHMARKS
-- ============================================================
INSERT INTO portfolio_nodes (node_id, org_id, parent_node_id, name, node_type, location_id, sort_order) VALUES
  ('ae000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', NULL, 'Chicha Group', 'org', NULL, 0),
  ('ae000001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae000001-0001-0001-0001-000000000001', 'Austin', 'region', NULL, 1),
  ('ae000001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae000001-0001-0001-0001-000000000002', 'Chicha Downtown', 'location', 'a1111111-1111-1111-1111-111111111111', 1),
  ('ae000001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae000001-0001-0001-0001-000000000002', 'Chicha Domain', 'location', 'b2222222-2222-2222-2222-222222222222', 2);

INSERT INTO location_benchmarks (org_id, location_id, period_start, period_end, revenue, food_cost_pct, labor_cost_pct, avg_check_cents, check_count, revenue_percentile, food_cost_percentile, labor_cost_percentile, avg_check_percentile) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '2026-03-01', '2026-03-20', 5800000, 31.500, 25.800, 4850, 1195, 75.0, 60.0, 55.0, 70.0),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '2026-03-01', '2026-03-20', 3200000, 29.200, 23.500, 5200, 615, 45.0, 75.0, 70.0, 80.0);

-- Best practices
INSERT INTO best_practices (org_id, title, description, metric, source_location_id, impact_pct, status) VALUES
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Optimized Prep Schedule', 'Domain location reduced food waste 15% by shifting prep times to align with demand patterns', 'food_cost_pct', 'b2222222-2222-2222-2222-222222222222', 2.300, 'suggested'),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cross-trained Staff Model', 'Downtown achieved lower labor costs through ELU-based cross-training program', 'labor_cost_pct', 'a1111111-1111-1111-1111-111111111111', 1.800, 'adopted');

-- ============================================================
-- P. SCHEDULING
-- ============================================================

-- Current week schedule
INSERT INTO schedules (schedule_id, org_id, location_id, week_start, status, created_by, published_at) VALUES
  ('af000001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', date_trunc('week', CURRENT_DATE)::date, 'published', '0d55e810-1e4a-417a-8a70-08b98f4595c2', now() - interval '3 days');

-- Scheduled shifts for current week (all 5 DT employees)
DO $$
DECLARE
  d DATE;
  dow INT;
  emp_ids UUID[] := ARRAY[
    'ee111111-1111-1111-1111-111111111111',
    'ee111111-2222-2222-2222-222222222222',
    'ee111111-3333-3333-3333-333333333333',
    'ee111111-4444-4444-4444-444444444444',
    'ee111111-5555-5555-5555-555555555555'
  ];
  emp UUID;
  idx INT;
  stations TEXT[] := ARRAY['expo', 'grill', 'saute', 'prep', 'ceviche_bar'];
BEGIN
  FOR d IN SELECT generate_series(date_trunc('week', CURRENT_DATE)::date, date_trunc('week', CURRENT_DATE)::date + 6, '1 day'::interval)::date LOOP
    dow := EXTRACT(DOW FROM d);
    FOREACH emp IN ARRAY emp_ids LOOP
      idx := array_position(emp_ids, emp);
      IF (dow + idx) % 7 < 5 THEN
        INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
        VALUES ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af000001-0001-0001-0001-000000000001', emp, d,
          CASE WHEN idx <= 3 THEN '08:00' ELSE '14:00' END::time,
          CASE WHEN idx <= 3 THEN '16:00' ELSE '22:00' END::time,
          stations[idx], 'confirmed');
      END IF;
    END LOOP;
  END LOOP;
END $$;

-- Labor demand forecast for today and tomorrow
INSERT INTO labor_demand_forecast (org_id, location_id, forecast_date, time_block, forecasted_covers, required_elu, required_headcount) VALUES
  -- Today
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '08:00', 5, 2.0, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '09:00', 8, 3.0, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '10:00', 12, 4.5, 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '11:00', 25, 8.0, 4),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '12:00', 40, 12.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '13:00', 35, 10.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '14:00', 15, 5.0, 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '15:00', 10, 3.5, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '16:00', 8, 3.0, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '17:00', 15, 5.0, 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '18:00', 45, 14.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '19:00', 50, 15.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '20:00', 40, 12.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE, '21:00', 20, 6.0, 3),
  -- Tomorrow
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '08:00', 5, 2.0, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '09:00', 8, 3.0, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '10:00', 12, 4.5, 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '11:00', 28, 9.0, 4),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '12:00', 45, 14.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '13:00', 38, 11.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '14:00', 15, 5.0, 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '15:00', 10, 3.5, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '16:00', 8, 3.0, 2),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '17:00', 18, 6.0, 3),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '18:00', 50, 15.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '19:00', 55, 16.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '20:00', 45, 13.0, 5),
  ('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', CURRENT_DATE + 1, '21:00', 25, 8.0, 4);

COMMIT;
