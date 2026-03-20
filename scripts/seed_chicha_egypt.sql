-- ============================================================================
-- CHICHA EGYPT: Modern Peruvian Restaurant Chain - 4 Branches
-- Complete reseed for org 3f7ef589-f499-43e3-a1c5-aaacd9d543ec
-- Currency: EGP stored as piasters (1 EGP = 100 piasters)
-- Run: docker exec -i fireline-postgres-1 psql -U fireline -d fireline < scripts/seed_chicha_egypt.sql
-- ============================================================================

SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

BEGIN;

-- ============================================================================
-- STEP 1: CLEAN EVERYTHING (FK order)
-- ============================================================================
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
DELETE FROM item_id_mappings WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
-- Delete employees (will cascade from shifts already deleted)
DELETE FROM employees WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
-- Delete extra locations and user_location_access
DELETE FROM user_location_access WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- ============================================================================
-- STEP 2: UPDATE ORG + CREATE 4 LOCATIONS
-- ============================================================================
UPDATE organizations SET name = 'Chicha Egypt', slug = 'chicha-egypt', updated_at = now()
  WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- Delete any extra locations first, then update existing and insert new
DELETE FROM locations WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
  AND location_id NOT IN ('a1111111-1111-1111-1111-111111111111', 'b2222222-2222-2222-2222-222222222222');

UPDATE locations SET name = 'Chicha El Gouna', address = 'Abu Tig Marina, El Gouna, Red Sea, Egypt', timezone = 'Africa/Cairo', updated_at = now()
  WHERE location_id = 'a1111111-1111-1111-1111-111111111111';
UPDATE locations SET name = 'Chicha New Cairo', address = 'Downtown Katameya, New Cairo, Egypt', timezone = 'Africa/Cairo', updated_at = now()
  WHERE location_id = 'b2222222-2222-2222-2222-222222222222';

INSERT INTO locations (location_id, org_id, name, address, timezone, status) VALUES
  ('c3333333-3333-3333-3333-333333333333', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Chicha Sheikh Zayed', 'Arkan Plaza, Sheikh Zayed City, Giza, Egypt', 'Africa/Cairo', 'active'),
  ('d4444444-4444-4444-4444-444444444444', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Chicha North Coast', 'Marassi, Sidi Abdel Rahman, North Coast, Egypt', 'Africa/Cairo', 'active')
ON CONFLICT DO NOTHING;

-- Grant owner access to all 4
INSERT INTO user_location_access (user_id, location_id, org_id)
SELECT u.user_id, l.location_id, l.org_id
FROM users u, locations l
WHERE u.org_id = l.org_id AND u.email = 'owner@bistrocloud.com'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- STEP 3: CREATE 120 EMPLOYEES (30 per branch)
-- ============================================================================
-- El Gouna (loc A) - 30 employees
INSERT INTO employees (employee_id, org_id, location_id, display_name, role, status, elu_ratings, staff_points, certifications, availability) VALUES
-- GM
('e1a00001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ahmed Hassan', 'gm', 'active', '{"grill":4.5,"saute":4.2,"prep":4.0,"expo":4.8,"ceviche_bar":3.8}', 142, ARRAY['servsafe_manager','food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
-- Shift Managers
('e1a00001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Fatma El-Sayed', 'shift_manager', 'active', '{"grill":4.0,"saute":3.8,"prep":3.5,"expo":4.2,"ceviche_bar":4.0}', 118, ARRAY['food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1a00001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Omar Farouk', 'shift_manager', 'active', '{"grill":3.8,"saute":4.0,"prep":3.5,"expo":4.0,"ceviche_bar":3.5}', 105, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
-- Servers
('e1a00001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Nour Ibrahim', 'staff', 'active', '{"expo":4.5,"prep":2.0}', 95, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sara Mahmoud', 'staff', 'active', '{"expo":4.2,"prep":2.5}', 88, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Yasmine Adel', 'staff', 'active', '{"expo":3.8,"prep":2.0}', 72, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
-- Line Cooks
('e1a00001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Karim Mostafa', 'staff', 'active', '{"grill":4.5,"saute":4.0,"prep":3.5,"ceviche_bar":4.2}', 130, ARRAY['food_handler','servsafe'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1a00001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Hassan Ali', 'staff', 'active', '{"grill":4.0,"saute":4.5,"prep":3.8,"ceviche_bar":3.5}', 115, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Tarek Nabil', 'staff', 'active', '{"grill":3.5,"saute":3.8,"prep":4.0,"ceviche_bar":4.5}', 98, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Amr Youssef', 'staff', 'active', '{"grill":3.8,"saute":3.5,"prep":4.2,"ceviche_bar":3.0}', 82, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
-- Prep Cooks
('e1a00001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Mona Khalil', 'staff', 'active', '{"prep":4.8,"saute":2.5,"ceviche_bar":3.0}', 110, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
('e1a00001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Heba Samir', 'staff', 'active', '{"prep":4.5,"saute":2.0,"ceviche_bar":2.5}', 76, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Dina Ashraf', 'staff', 'active', '{"prep":4.2,"saute":2.8,"ceviche_bar":3.5}', 65, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
-- Bartenders
('e1a00001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Layla Osman', 'staff', 'active', '{"expo":3.5,"prep":2.0}', 90, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Youssef Ramadan', 'staff', 'active', '{"expo":3.0,"prep":1.5}', 68, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
-- Hosts
('e1a00001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Rana Sherif', 'staff', 'active', '{"expo":3.0}', 55, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Mariam Gamal', 'staff', 'active', '{"expo":2.8}', 48, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
-- Dishwashers
('e1a00001-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Mahmoud Saber', 'staff', 'active', '{"dish":4.5,"prep":2.0}', 35, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ali Fathi', 'staff', 'active', '{"dish":4.0,"prep":2.5}', 28, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1a00001-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Khaled Zaki', 'staff', 'active', '{"dish":3.8,"prep":1.5}', 22, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
-- Runners
('e1a00001-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Abdallah Medhat', 'staff', 'active', '{"expo":3.5,"dish":2.0}', 40, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Hesham Wael', 'staff', 'active', '{"expo":3.2,"dish":2.5}', 33, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
-- Additional staff
('e1a00001-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Noha Tamer', 'staff', 'active', '{"prep":3.0,"expo":2.5}', 45, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1a00001-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sherif Magdy', 'staff', 'active', '{"grill":3.0,"prep":3.5}', 52, ARRAY['food_handler'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000025', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Salma Ehab', 'staff', 'active', '{"expo":3.8,"prep":2.0}', 60, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
('e1a00001-0001-0001-0001-000000000026', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ayman Hosny', 'staff', 'active', '{"saute":3.5,"prep":3.8}', 70, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000027', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Reem Magdi', 'staff', 'active', '{"expo":3.5,"ceviche_bar":3.0}', 38, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000028', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Mostafa Reda', 'staff', 'active', '{"grill":2.5,"prep":3.0,"dish":3.5}', 30, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1a00001-0001-0001-0001-000000000029', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Hana Ashour', 'staff', 'active', '{"prep":3.5,"expo":3.0}', 42, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1a00001-0001-0001-0001-000000000030', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Waleed Sameh', 'staff', 'active', '{"dish":3.5,"prep":2.5}', 25, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}');

-- New Cairo (loc B) - 30 employees
INSERT INTO employees (employee_id, org_id, location_id, display_name, role, status, elu_ratings, staff_points, certifications, availability) VALUES
('e1b00001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Mohamed Ezzat', 'gm', 'active', '{"grill":4.2,"saute":4.5,"prep":4.0,"expo":4.8,"ceviche_bar":4.0}', 150, ARRAY['servsafe_manager','food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
('e1b00001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Aisha Lotfy', 'shift_manager', 'active', '{"grill":3.8,"saute":4.0,"prep":3.5,"expo":4.2,"ceviche_bar":3.5}', 125, ARRAY['food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1b00001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Tamer Adly', 'shift_manager', 'active', '{"grill":4.0,"saute":3.5,"prep":3.8,"expo":3.8,"ceviche_bar":3.0}', 108, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Farida Hossam', 'staff', 'active', '{"expo":4.5,"prep":2.0}', 92, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Adham Khaled', 'staff', 'active', '{"expo":4.0,"prep":2.5}', 85, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Malak Hany', 'staff', 'active', '{"expo":3.8,"prep":2.2}', 78, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Wael Sobhy', 'staff', 'active', '{"grill":4.5,"saute":4.2,"prep":3.5,"ceviche_bar":4.0}', 135, ARRAY['food_handler','servsafe'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1b00001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Rania Fouad', 'staff', 'active', '{"grill":3.8,"saute":4.5,"prep":4.0,"ceviche_bar":3.8}', 120, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Hazem Atef', 'staff', 'active', '{"grill":4.0,"saute":3.8,"prep":3.5,"ceviche_bar":4.5}', 100, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Samar Helmy', 'staff', 'active', '{"grill":3.5,"saute":3.2,"prep":4.0,"ceviche_bar":3.0}', 75, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
('e1b00001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Bassem Naguib', 'staff', 'active', '{"prep":4.8,"saute":2.5,"ceviche_bar":3.0}', 105, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
('e1b00001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Eman Ragab', 'staff', 'active', '{"prep":4.5,"saute":2.0}', 80, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Shady Emad', 'staff', 'active', '{"prep":4.0,"saute":3.0,"ceviche_bar":3.2}', 62, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Habiba Saeed', 'staff', 'active', '{"expo":3.5,"prep":2.0}', 88, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Kareem Hafez', 'staff', 'active', '{"expo":3.0,"prep":1.5}', 65, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Nadine Wahba', 'staff', 'active', '{"expo":3.0}', 50, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Yehia Fawzy', 'staff', 'active', '{"expo":2.8}', 42, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Gamal Abdel-Nasser', 'staff', 'active', '{"dish":4.5,"prep":2.0}', 32, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Samia Refaat', 'staff', 'active', '{"dish":4.0,"prep":2.5}', 26, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1b00001-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Osama Hamdy', 'staff', 'active', '{"dish":3.8,"prep":1.5}', 20, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Nada Kamel', 'staff', 'active', '{"expo":3.5,"dish":2.0}', 38, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Ibrahim Taha', 'staff', 'active', '{"expo":3.2,"dish":2.5}', 30, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Abeer Nagah', 'staff', 'active', '{"prep":3.0,"expo":2.5}', 48, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1b00001-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Ziad Mansour', 'staff', 'active', '{"grill":3.0,"prep":3.5}', 55, ARRAY['food_handler'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000025', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Hoda Essam', 'staff', 'active', '{"expo":3.8,"prep":2.0}', 58, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
('e1b00001-0001-0001-0001-000000000026', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Mahmoud Farid', 'staff', 'active', '{"saute":3.5,"prep":3.8}', 72, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000027', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Menna Allah', 'staff', 'active', '{"expo":3.5,"ceviche_bar":3.0}', 35, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000028', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Seif Barakat', 'staff', 'active', '{"grill":2.5,"prep":3.0,"dish":3.5}', 28, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1b00001-0001-0001-0001-000000000029', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Lina Fathi', 'staff', 'active', '{"prep":3.5,"expo":3.0}', 44, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1b00001-0001-0001-0001-000000000030', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Hamza Shawky', 'staff', 'active', '{"dish":3.5,"prep":2.5}', 22, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}');

-- Sheikh Zayed (loc C) - 30 employees
INSERT INTO employees (employee_id, org_id, location_id, display_name, role, status, elu_ratings, staff_points, certifications, availability) VALUES
('e1c00001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Hossam Galal', 'gm', 'active', '{"grill":4.0,"saute":4.5,"prep":4.2,"expo":4.5,"ceviche_bar":4.0}', 138, ARRAY['servsafe_manager','food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
('e1c00001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Ghada Nabil', 'shift_manager', 'active', '{"grill":3.8,"saute":4.0,"prep":3.5,"expo":4.0,"ceviche_bar":3.5}', 115, ARRAY['food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1c00001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Ramy Said', 'shift_manager', 'active', '{"grill":4.0,"saute":3.5,"prep":3.8,"expo":3.8,"ceviche_bar":3.2}', 98, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Amira Sami', 'staff', 'active', '{"expo":4.5,"prep":2.0}', 90, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Bassel Amin', 'staff', 'active', '{"expo":4.0,"prep":2.5}', 82, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Lamia Ashraf', 'staff', 'active', '{"expo":3.8,"prep":2.0}', 68, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Ehab Saeed', 'staff', 'active', '{"grill":4.5,"saute":4.0,"prep":3.5,"ceviche_bar":4.0}', 128, ARRAY['food_handler','servsafe'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1c00001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Niveen Adel', 'staff', 'active', '{"grill":3.8,"saute":4.5,"prep":3.8,"ceviche_bar":3.5}', 112, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Fady Girgis', 'staff', 'active', '{"grill":3.5,"saute":3.8,"prep":4.0,"ceviche_bar":4.2}', 95, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Dalia Hosni', 'staff', 'active', '{"grill":3.5,"saute":3.2,"prep":4.0,"ceviche_bar":3.0}', 70, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
('e1c00001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Akram Talaat', 'staff', 'active', '{"prep":4.8,"saute":2.5}', 100, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
('e1c00001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Sawsan Emad', 'staff', 'active', '{"prep":4.5,"saute":2.0}', 75, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Khairy Abbas', 'staff', 'active', '{"prep":4.2,"saute":2.8}', 58, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Yara Maher', 'staff', 'active', '{"expo":3.5,"prep":2.0}', 85, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Ashraf Helal', 'staff', 'active', '{"expo":3.0,"prep":1.5}', 62, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Soha Rashed', 'staff', 'active', '{"expo":3.0}', 45, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Tayseer Moussa', 'staff', 'active', '{"expo":2.8}', 38, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Sayed Hussein', 'staff', 'active', '{"dish":4.5,"prep":2.0}', 30, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Nagwa Hashem', 'staff', 'active', '{"dish":4.0,"prep":2.5}', 24, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1c00001-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Essam Tolba', 'staff', 'active', '{"dish":3.8,"prep":1.5}', 20, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Maha Yasser', 'staff', 'active', '{"expo":3.5,"dish":2.0}', 35, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Alaa Gouda', 'staff', 'active', '{"expo":3.2,"dish":2.5}', 28, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Aya Farouk', 'staff', 'active', '{"prep":3.0,"expo":2.5}', 40, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1c00001-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Ismail Shaker', 'staff', 'active', '{"grill":3.0,"prep":3.5}', 50, ARRAY['food_handler'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000025', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Rehab Sobhy', 'staff', 'active', '{"expo":3.8,"prep":2.0}', 55, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
('e1c00001-0001-0001-0001-000000000026', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Medhat Ramzy', 'staff', 'active', '{"saute":3.5,"prep":3.8}', 65, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000027', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Perihan Lotfy', 'staff', 'active', '{"expo":3.5,"ceviche_bar":3.0}', 32, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000028', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Haitham Eid', 'staff', 'active', '{"grill":2.5,"prep":3.0,"dish":3.5}', 25, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1c00001-0001-0001-0001-000000000029', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Shahira Ismail', 'staff', 'active', '{"prep":3.5,"expo":3.0}', 40, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1c00001-0001-0001-0001-000000000030', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Sameh Tawfik', 'staff', 'active', '{"dish":3.5,"prep":2.5}', 22, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}');

-- North Coast (loc D) - 30 employees
INSERT INTO employees (employee_id, org_id, location_id, display_name, role, status, elu_ratings, staff_points, certifications, availability) VALUES
('e1d00001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Sherif Abdel-Fattah', 'gm', 'active', '{"grill":4.2,"saute":4.0,"prep":4.2,"expo":4.5,"ceviche_bar":4.2}', 145, ARRAY['servsafe_manager','food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
('e1d00001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Mai Talaat', 'shift_manager', 'active', '{"grill":3.5,"saute":4.0,"prep":3.5,"expo":4.2,"ceviche_bar":3.8}', 120, ARRAY['food_handler','first_aid'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1d00001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Emad Labib', 'shift_manager', 'active', '{"grill":4.0,"saute":3.5,"prep":3.8,"expo":3.5,"ceviche_bar":3.0}', 102, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Ingy Mohsen', 'staff', 'active', '{"expo":4.5,"prep":2.0}', 88, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Murad Helmy', 'staff', 'active', '{"expo":4.0,"prep":2.5}', 80, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Omnia Adel', 'staff', 'active', '{"expo":3.8,"prep":2.0}', 65, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Nasser Gaber', 'staff', 'active', '{"grill":4.5,"saute":4.2,"prep":3.5,"ceviche_bar":4.0}', 132, ARRAY['food_handler','servsafe'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1d00001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Ola Fouad', 'staff', 'active', '{"grill":3.8,"saute":4.5,"prep":3.8,"ceviche_bar":3.5}', 118, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Sami Refaat', 'staff', 'active', '{"grill":4.0,"saute":3.8,"prep":3.5,"ceviche_bar":4.5}', 95, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Dalia Rizk', 'staff', 'active', '{"grill":3.5,"saute":3.2,"prep":4.0,"ceviche_bar":3.0}', 72, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
('e1d00001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Wagdy Shokry', 'staff', 'active', '{"prep":4.8,"saute":2.5}', 108, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":false}'),
('e1d00001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Amal Naguib', 'staff', 'active', '{"prep":4.5,"saute":2.0}', 78, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Shawky Bakr', 'staff', 'active', '{"prep":4.0,"saute":3.0}', 55, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Nesma Fawzy', 'staff', 'active', '{"expo":3.5,"prep":2.0}', 82, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Hisham Anwar', 'staff', 'active', '{"expo":3.0,"prep":1.5}', 60, ARRAY['food_handler','tips_certified','alcohol_cert'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Doha Magdy', 'staff', 'active', '{"expo":3.0}', 42, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Rafik Abdallah', 'staff', 'active', '{"expo":2.8}', 35, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Ashraf Badawy', 'staff', 'active', '{"dish":4.5,"prep":2.0}', 28, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Hanan Mohsen', 'staff', 'active', '{"dish":4.0,"prep":2.5}', 22, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1d00001-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Reda Mostafa', 'staff', 'active', '{"dish":3.8,"prep":1.5}', 20, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Nashwa Atef', 'staff', 'active', '{"expo":3.5,"dish":2.0}', 38, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":false,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Hatem Galal', 'staff', 'active', '{"expo":3.2,"dish":2.5}', 30, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":false,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Mervat Osman', 'staff', 'active', '{"prep":3.0,"expo":2.5}', 42, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1d00001-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Tawfik Shehata', 'staff', 'active', '{"grill":3.0,"prep":3.5}', 48, ARRAY['food_handler'], '{"sun":true,"mon":false,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000025', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Riham Saad', 'staff', 'active', '{"expo":3.8,"prep":2.0}', 52, ARRAY['food_handler','tips_certified'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":false}'),
('e1d00001-0001-0001-0001-000000000026', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Magdy Helal', 'staff', 'active', '{"saute":3.5,"prep":3.8}', 68, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000027', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Azza Ramadan', 'staff', 'active', '{"expo":3.5,"ceviche_bar":3.0}', 30, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":false,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000028', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Fathy Mourad', 'staff', 'active', '{"grill":2.5,"prep":3.0,"dish":3.5}', 24, ARRAY['food_handler'], '{"sun":false,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}'),
('e1d00001-0001-0001-0001-000000000029', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Rawya Shams', 'staff', 'active', '{"prep":3.5,"expo":3.0}', 38, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":false,"sat":true}'),
('e1d00001-0001-0001-0001-000000000030', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Taher Yousry', 'staff', 'active', '{"dish":3.5,"prep":2.5}', 20, ARRAY['food_handler'], '{"sun":true,"mon":true,"tue":true,"wed":true,"thu":true,"fri":true,"sat":true}');

-- ============================================================================
-- STEP 4: MENU ITEMS (15 per location, same menu across chain)
-- ============================================================================
-- Menu item IDs: mi{loc_num}{item_num} pattern
-- Loc A items: aa100001-..., Loc B: bb100001-..., Loc C: cc100001-..., Loc D: dd100001-...

INSERT INTO menu_items (menu_item_id, org_id, location_id, name, category, price, available, description, source) VALUES
-- El Gouna (A)
('aa100001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ceviche Clasico', 'appetizers', 28500, true, 'Fresh sea bass cured in tiger''s milk with red onion, cilantro, and cancha corn', 'manual'),
('aa100001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Tiradito Nikkei', 'appetizers', 32000, true, 'Thinly sliced sea bass with Japanese-Peruvian aji amarillo sauce and sesame', 'manual'),
('aa100001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Causa Limena', 'appetizers', 22000, true, 'Layered purple potato terrine with avocado, chicken, and aji amarillo', 'manual'),
('aa100001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Anticuchos de Corazon', 'appetizers', 24500, true, 'Grilled beef heart skewers with rocoto pepper sauce and boiled potatoes', 'manual'),
('aa100001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Empanadas (3 pcs)', 'appetizers', 18000, true, 'Crispy empanadas filled with seasoned beef and olives', 'manual'),
('aa100001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Lomo Saltado', 'mains', 42000, true, 'Stir-fried beef tenderloin with tomatoes, onions, and fries in soy-vinegar sauce', 'manual'),
('aa100001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Aji de Gallina', 'mains', 35000, true, 'Shredded chicken in creamy aji amarillo and walnut sauce with rice', 'manual'),
('aa100001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Arroz con Mariscos', 'mains', 48000, true, 'Peruvian seafood rice with shrimp, calamari, and mussels in aji panca', 'manual'),
('aa100001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Pollo a la Brasa', 'mains', 38000, true, 'Rotisserie chicken marinated in cumin, paprika, and aji panca with green sauce', 'manual'),
('aa100001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Seco de Res', 'mains', 39500, true, 'Slow-braised beef in cilantro and beer sauce with canario beans', 'manual'),
('aa100001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Churrasco Chimichurri', 'mains', 52000, true, 'Grilled beef tenderloin with house chimichurri, yuca fries, and salad', 'manual'),
('aa100001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Papas a la Huancaina', 'sides', 14500, true, 'Boiled potatoes with creamy aji amarillo cheese sauce', 'manual'),
('aa100001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Choclo con Queso', 'sides', 12000, true, 'Giant Peruvian corn with fresh cheese', 'manual'),
('aa100001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Pisco Sour', 'beverages', 18500, true, 'Classic Peruvian cocktail with pisco, lime, egg white, and bitters', 'manual'),
('aa100001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Chicha Morada', 'beverages', 8500, true, 'Traditional purple corn drink with cinnamon, cloves, and lime', 'manual'),
-- New Cairo (B) - same menu, same prices
('bb100001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Ceviche Clasico', 'appetizers', 28500, true, 'Fresh sea bass cured in tiger''s milk with red onion, cilantro, and cancha corn', 'manual'),
('bb100001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Tiradito Nikkei', 'appetizers', 32000, true, 'Thinly sliced sea bass with Japanese-Peruvian aji amarillo sauce and sesame', 'manual'),
('bb100001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Causa Limena', 'appetizers', 22000, true, 'Layered purple potato terrine with avocado, chicken, and aji amarillo', 'manual'),
('bb100001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Anticuchos de Corazon', 'appetizers', 24500, true, 'Grilled beef heart skewers with rocoto pepper sauce and boiled potatoes', 'manual'),
('bb100001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Empanadas (3 pcs)', 'appetizers', 18000, true, 'Crispy empanadas filled with seasoned beef and olives', 'manual'),
('bb100001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Lomo Saltado', 'mains', 42000, true, 'Stir-fried beef tenderloin with tomatoes, onions, and fries in soy-vinegar sauce', 'manual'),
('bb100001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Aji de Gallina', 'mains', 35000, true, 'Shredded chicken in creamy aji amarillo and walnut sauce with rice', 'manual'),
('bb100001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Arroz con Mariscos', 'mains', 48000, true, 'Peruvian seafood rice with shrimp, calamari, and mussels in aji panca', 'manual'),
('bb100001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Pollo a la Brasa', 'mains', 38000, true, 'Rotisserie chicken marinated in cumin, paprika, and aji panca with green sauce', 'manual'),
('bb100001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Seco de Res', 'mains', 39500, true, 'Slow-braised beef in cilantro and beer sauce with canario beans', 'manual'),
('bb100001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Churrasco Chimichurri', 'mains', 52000, true, 'Grilled beef tenderloin with house chimichurri, yuca fries, and salad', 'manual'),
('bb100001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Papas a la Huancaina', 'sides', 14500, true, 'Boiled potatoes with creamy aji amarillo cheese sauce', 'manual'),
('bb100001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Choclo con Queso', 'sides', 12000, true, 'Giant Peruvian corn with fresh cheese', 'manual'),
('bb100001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Pisco Sour', 'beverages', 18500, true, 'Classic Peruvian cocktail with pisco, lime, egg white, and bitters', 'manual'),
('bb100001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Chicha Morada', 'beverages', 8500, true, 'Traditional purple corn drink with cinnamon, cloves, and lime', 'manual'),
-- Sheikh Zayed (C)
('cc100001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Ceviche Clasico', 'appetizers', 28500, true, 'Fresh sea bass cured in tiger''s milk with red onion, cilantro, and cancha corn', 'manual'),
('cc100001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Tiradito Nikkei', 'appetizers', 32000, true, 'Thinly sliced sea bass with Japanese-Peruvian aji amarillo sauce and sesame', 'manual'),
('cc100001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Causa Limena', 'appetizers', 22000, true, 'Layered purple potato terrine with avocado, chicken, and aji amarillo', 'manual'),
('cc100001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Anticuchos de Corazon', 'appetizers', 24500, true, 'Grilled beef heart skewers with rocoto pepper sauce and boiled potatoes', 'manual'),
('cc100001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Empanadas (3 pcs)', 'appetizers', 18000, true, 'Crispy empanadas filled with seasoned beef and olives', 'manual'),
('cc100001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Lomo Saltado', 'mains', 42000, true, 'Stir-fried beef tenderloin with tomatoes, onions, and fries in soy-vinegar sauce', 'manual'),
('cc100001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Aji de Gallina', 'mains', 35000, true, 'Shredded chicken in creamy aji amarillo and walnut sauce with rice', 'manual'),
('cc100001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Arroz con Mariscos', 'mains', 48000, true, 'Peruvian seafood rice with shrimp, calamari, and mussels in aji panca', 'manual'),
('cc100001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Pollo a la Brasa', 'mains', 38000, true, 'Rotisserie chicken marinated in cumin, paprika, and aji panca with green sauce', 'manual'),
('cc100001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Seco de Res', 'mains', 39500, true, 'Slow-braised beef in cilantro and beer sauce with canario beans', 'manual'),
('cc100001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Churrasco Chimichurri', 'mains', 52000, true, 'Grilled beef tenderloin with house chimichurri, yuca fries, and salad', 'manual'),
('cc100001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Papas a la Huancaina', 'sides', 14500, true, 'Boiled potatoes with creamy aji amarillo cheese sauce', 'manual'),
('cc100001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Choclo con Queso', 'sides', 12000, true, 'Giant Peruvian corn with fresh cheese', 'manual'),
('cc100001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Pisco Sour', 'beverages', 18500, true, 'Classic Peruvian cocktail with pisco, lime, egg white, and bitters', 'manual'),
('cc100001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Chicha Morada', 'beverages', 8500, true, 'Traditional purple corn drink with cinnamon, cloves, and lime', 'manual'),
-- North Coast (D)
('dd100001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Ceviche Clasico', 'appetizers', 28500, true, 'Fresh sea bass cured in tiger''s milk with red onion, cilantro, and cancha corn', 'manual'),
('dd100001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Tiradito Nikkei', 'appetizers', 32000, true, 'Thinly sliced sea bass with Japanese-Peruvian aji amarillo sauce and sesame', 'manual'),
('dd100001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Causa Limena', 'appetizers', 22000, true, 'Layered purple potato terrine with avocado, chicken, and aji amarillo', 'manual'),
('dd100001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Anticuchos de Corazon', 'appetizers', 24500, true, 'Grilled beef heart skewers with rocoto pepper sauce and boiled potatoes', 'manual'),
('dd100001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Empanadas (3 pcs)', 'appetizers', 18000, true, 'Crispy empanadas filled with seasoned beef and olives', 'manual'),
('dd100001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Lomo Saltado', 'mains', 42000, true, 'Stir-fried beef tenderloin with tomatoes, onions, and fries in soy-vinegar sauce', 'manual'),
('dd100001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Aji de Gallina', 'mains', 35000, true, 'Shredded chicken in creamy aji amarillo and walnut sauce with rice', 'manual'),
('dd100001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Arroz con Mariscos', 'mains', 48000, true, 'Peruvian seafood rice with shrimp, calamari, and mussels in aji panca', 'manual'),
('dd100001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Pollo a la Brasa', 'mains', 38000, true, 'Rotisserie chicken marinated in cumin, paprika, and aji panca with green sauce', 'manual'),
('dd100001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Seco de Res', 'mains', 39500, true, 'Slow-braised beef in cilantro and beer sauce with canario beans', 'manual'),
('dd100001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Churrasco Chimichurri', 'mains', 52000, true, 'Grilled beef tenderloin with house chimichurri, yuca fries, and salad', 'manual'),
('dd100001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Papas a la Huancaina', 'sides', 14500, true, 'Boiled potatoes with creamy aji amarillo cheese sauce', 'manual'),
('dd100001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Choclo con Queso', 'sides', 12000, true, 'Giant Peruvian corn with fresh cheese', 'manual'),
('dd100001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Pisco Sour', 'beverages', 18500, true, 'Classic Peruvian cocktail with pisco, lime, egg white, and bitters', 'manual'),
('dd100001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Chicha Morada', 'beverages', 8500, true, 'Traditional purple corn drink with cinnamon, cloves, and lime', 'manual');

-- ============================================================================
-- STEP 5: INGREDIENTS (25) + RECIPES + RECIPE_EXPLOSION
-- ============================================================================
-- Ingredient IDs: 1a100001-... through in100025-...
-- Cost in piasters per unit

INSERT INTO ingredients (ingredient_id, org_id, name, category, unit, cost_per_unit, prep_yield_factor, allergens, status) VALUES
('1a100001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Sea Bass Fillet', 'protein', 'lb', 4500, 0.8500, ARRAY['fish'], 'active'),
('1a100001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Beef Tenderloin', 'protein', 'lb', 6500, 0.9000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Chicken Thighs', 'protein', 'lb', 1200, 0.8500, '{}', 'active'),
('1a100001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Jumbo Shrimp', 'protein', 'lb', 5200, 0.8000, ARRAY['shellfish'], 'active'),
('1a100001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Beef Heart', 'protein', 'lb', 1800, 0.9000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Purple Potatoes', 'produce', 'lb', 450, 0.9000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Sweet Potato', 'produce', 'lb', 300, 0.8500, '{}', 'active'),
('1a100001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Red Onion', 'produce', 'lb', 200, 0.9000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Lime', 'produce', 'ea', 50, 0.9500, '{}', 'active'),
('1a100001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cilantro', 'produce', 'bunch', 150, 0.7000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Avocado', 'produce', 'ea', 350, 0.7500, '{}', 'active'),
('1a100001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Corn (Choclo)', 'produce', 'ea', 200, 0.8500, '{}', 'active'),
('1a100001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Aji Amarillo Paste', 'sauce', 'oz', 80, 1.0000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Rocoto Pepper', 'produce', 'lb', 1200, 0.8500, '{}', 'active'),
('1a100001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Pisco', 'beverage', 'oz', 200, 1.0000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cancha Corn', 'grain', 'lb', 650, 1.0000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Leche de Tigre Base', 'sauce', 'oz', 90, 1.0000, ARRAY['fish'], 'active'),
('1a100001-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cotija Cheese', 'dairy', 'lb', 1400, 0.9500, ARRAY['dairy'], 'active'),
('1a100001-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Rice', 'grain', 'lb', 150, 1.0000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Calamari', 'protein', 'lb', 3800, 0.8000, ARRAY['shellfish'], 'active'),
('1a100001-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Egg', 'dairy', 'ea', 80, 0.9000, ARRAY['eggs'], 'active'),
('1a100001-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Purple Corn', 'produce', 'lb', 500, 0.9000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Empanada Dough', 'bakery', 'ea', 120, 1.0000, ARRAY['gluten'], 'active'),
('1a100001-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Yuca', 'produce', 'lb', 350, 0.8000, '{}', 'active'),
('1a100001-0001-0001-0001-000000000025', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Beer (Cooking)', 'beverage', 'oz', 30, 1.0000, ARRAY['gluten'], 'active');

-- Recipes (one per menu item, using Loc A items — explosion materialized for all locations)
INSERT INTO recipes (recipe_id, org_id, menu_item_id, name, yield_quantity, yield_unit, prep_time_minutes, status) VALUES
('ae100001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000001', 'Ceviche Clasico', 1, 'ea', 15, 'active'),
('ae100001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000002', 'Tiradito Nikkei', 1, 'ea', 12, 'active'),
('ae100001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000003', 'Causa Limena', 1, 'ea', 20, 'active'),
('ae100001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000004', 'Anticuchos de Corazon', 1, 'ea', 25, 'active'),
('ae100001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000005', 'Empanadas (3 pcs)', 1, 'ea', 30, 'active'),
('ae100001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000006', 'Lomo Saltado', 1, 'ea', 18, 'active'),
('ae100001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000007', 'Aji de Gallina', 1, 'ea', 35, 'active'),
('ae100001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000008', 'Arroz con Mariscos', 1, 'ea', 25, 'active'),
('ae100001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000009', 'Pollo a la Brasa', 1, 'ea', 60, 'active'),
('ae100001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000010', 'Seco de Res', 1, 'ea', 90, 'active'),
('ae100001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000011', 'Churrasco Chimichurri', 1, 'ea', 20, 'active'),
('ae100001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000012', 'Papas a la Huancaina', 1, 'ea', 15, 'active'),
('ae100001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000013', 'Choclo con Queso', 1, 'ea', 10, 'active'),
('ae100001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000014', 'Pisco Sour', 1, 'ea', 5, 'active'),
('ae100001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000015', 'Chicha Morada', 1, 'ea', 5, 'active');

-- Recipe Ingredients (3-5 per recipe)
INSERT INTO recipe_ingredients (org_id, recipe_id, ingredient_id, quantity, unit) VALUES
-- Ceviche Clasico: sea bass, lime, red onion, cilantro, cancha, leche de tigre
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000001', 0.375, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000009', 2.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000008', 0.125, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000010', 0.25, 'bunch'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000016', 0.0625, 'lb'),
-- Tiradito Nikkei: sea bass, aji amarillo, lime, sesame
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000002', '1a100001-0001-0001-0001-000000000001', 0.375, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000002', '1a100001-0001-0001-0001-000000000013', 1.0, 'oz'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000002', '1a100001-0001-0001-0001-000000000009', 1.5, 'ea'),
-- Causa Limena: purple potatoes, avocado, chicken, aji amarillo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000003', '1a100001-0001-0001-0001-000000000006', 0.5, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000003', '1a100001-0001-0001-0001-000000000011', 1.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000003', '1a100001-0001-0001-0001-000000000003', 0.25, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000003', '1a100001-0001-0001-0001-000000000013', 0.5, 'oz'),
-- Anticuchos: beef heart, rocoto pepper, sweet potato
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000004', '1a100001-0001-0001-0001-000000000005', 0.5, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000004', '1a100001-0001-0001-0001-000000000014', 0.0625, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000004', '1a100001-0001-0001-0001-000000000007', 0.375, 'lb'),
-- Empanadas: beef tenderloin, empanada dough, red onion
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000005', '1a100001-0001-0001-0001-000000000002', 0.25, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000005', '1a100001-0001-0001-0001-000000000023', 3.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000005', '1a100001-0001-0001-0001-000000000008', 0.125, 'lb'),
-- Lomo Saltado: beef tenderloin, red onion, lime, rice
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000006', '1a100001-0001-0001-0001-000000000002', 0.5, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000006', '1a100001-0001-0001-0001-000000000008', 0.25, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000006', '1a100001-0001-0001-0001-000000000009', 1.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000006', '1a100001-0001-0001-0001-000000000019', 0.375, 'lb'),
-- Aji de Gallina: chicken, aji amarillo, rice, egg
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000007', '1a100001-0001-0001-0001-000000000003', 0.5, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000007', '1a100001-0001-0001-0001-000000000013', 1.5, 'oz'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000007', '1a100001-0001-0001-0001-000000000019', 0.375, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000007', '1a100001-0001-0001-0001-000000000021', 1.0, 'ea'),
-- Arroz con Mariscos: shrimp, calamari, rice, aji amarillo, lime
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000008', '1a100001-0001-0001-0001-000000000004', 0.375, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000008', '1a100001-0001-0001-0001-000000000020', 0.25, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000008', '1a100001-0001-0001-0001-000000000019', 0.5, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000008', '1a100001-0001-0001-0001-000000000013', 1.0, 'oz'),
-- Pollo a la Brasa: chicken, lime, aji amarillo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000009', '1a100001-0001-0001-0001-000000000003', 0.75, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000009', '1a100001-0001-0001-0001-000000000009', 1.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000009', '1a100001-0001-0001-0001-000000000013', 0.5, 'oz'),
-- Seco de Res: beef tenderloin, cilantro, beer, rice
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000010', '1a100001-0001-0001-0001-000000000002', 0.5, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000010', '1a100001-0001-0001-0001-000000000010', 0.5, 'bunch'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000010', '1a100001-0001-0001-0001-000000000025', 4.0, 'oz'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000010', '1a100001-0001-0001-0001-000000000019', 0.375, 'lb'),
-- Churrasco Chimichurri: beef tenderloin, cilantro, lime, yuca
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000011', '1a100001-0001-0001-0001-000000000002', 0.625, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000011', '1a100001-0001-0001-0001-000000000010', 0.5, 'bunch'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000011', '1a100001-0001-0001-0001-000000000009', 1.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000011', '1a100001-0001-0001-0001-000000000024', 0.375, 'lb'),
-- Papas a la Huancaina: purple potatoes, aji amarillo, cotija cheese
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000012', '1a100001-0001-0001-0001-000000000006', 0.5, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000012', '1a100001-0001-0001-0001-000000000013', 1.0, 'oz'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000012', '1a100001-0001-0001-0001-000000000018', 0.125, 'lb'),
-- Choclo con Queso: corn, cotija cheese
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000013', '1a100001-0001-0001-0001-000000000012', 2.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000013', '1a100001-0001-0001-0001-000000000018', 0.125, 'lb'),
-- Pisco Sour: pisco, lime, egg
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000014', '1a100001-0001-0001-0001-000000000015', 3.0, 'oz'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000014', '1a100001-0001-0001-0001-000000000009', 1.0, 'ea'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000014', '1a100001-0001-0001-0001-000000000021', 1.0, 'ea'),
-- Chicha Morada: purple corn, lime
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000015', '1a100001-0001-0001-0001-000000000022', 0.125, 'lb'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae100001-0001-0001-0001-000000000015', '1a100001-0001-0001-0001-000000000009', 0.5, 'ea');

-- Materialize recipe_explosion for ALL 60 menu items (15 per location x 4)
-- We use a DO block to copy from loc A recipes to all locations
DO $$
DECLARE
  v_org uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
  v_loc_prefixes text[] := ARRAY['aa','bb','cc','dd'];
  v_recipe_items text[] := ARRAY['01','02','03','04','05','06','07','08','09','10','11','12','13','14','15'];
  v_prefix text;
  v_item_num text;
  v_menu_item_id uuid;
  v_ri RECORD;
  i int; j int;
BEGIN
  FOR i IN 1..4 LOOP
    v_prefix := v_loc_prefixes[i];
    FOR j IN 1..15 LOOP
      v_item_num := lpad(j::text, 2, '0');
      v_menu_item_id := (v_prefix || '100001-0001-0001-0001-0000000000' || v_item_num)::uuid;
      -- Get recipe ingredients from the loc A recipe (recipe index j)
      FOR v_ri IN
        SELECT ingredient_id, quantity, unit
        FROM recipe_ingredients
        WHERE recipe_id = ('ae100001-0001-0001-0001-0000000000' || v_item_num)::uuid
      LOOP
        INSERT INTO recipe_explosion (org_id, menu_item_id, ingredient_id, quantity_per_unit, unit)
        VALUES (v_org, v_menu_item_id, v_ri.ingredient_id, v_ri.quantity, v_ri.unit)
        ON CONFLICT (menu_item_id, ingredient_id) DO NOTHING;
      END LOOP;
    END LOOP;
  END LOOP;
  RAISE NOTICE 'Recipe explosion materialized for all 60 menu items';
END $$;

COMMIT;
-- ============================================================================
-- CHICHA EGYPT Part 2: Orders, Inventory, Waste, POs, Budgets, Shifts
-- ============================================================================
SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

BEGIN;

-- ============================================================================
-- STEP 6: GENERATE 30 DAYS OF ORDERS
-- Target: ~250 orders/branch/day = ~100K EGP/day avg check ~400 EGP (40000 piasters)
-- 4 branches x 30 days = ~30,000 orders
-- Egypt weekend = Fri/Sat (DOW 5,6)
-- ============================================================================
DO $$
DECLARE
    v_org_id      uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_locations   uuid[] := ARRAY[
        'a1111111-1111-1111-1111-111111111111'::uuid,
        'b2222222-2222-2222-2222-222222222222'::uuid,
        'c3333333-3333-3333-3333-333333333333'::uuid,
        'd4444444-4444-4444-4444-444444444444'::uuid
    ];
    v_loc_prefixes text[] := ARRAY['aa','bb','cc','dd'];
    v_day         date;
    v_check_id    uuid;
    v_num_orders  int;
    v_channel     text;
    v_channels    text[] := ARRAY['dine_in','dine_in','dine_in','dine_in','dine_in','dine_in',
                                   'takeout','takeout','takeout','takeout','takeout',
                                   'delivery','delivery','delivery','delivery'];
    v_item_ids    uuid[];
    v_item_names  text[] := ARRAY['Ceviche Clasico','Tiradito Nikkei','Causa Limena','Anticuchos de Corazon',
                                   'Empanadas (3 pcs)','Lomo Saltado','Aji de Gallina','Arroz con Mariscos',
                                   'Pollo a la Brasa','Seco de Res','Churrasco Chimichurri',
                                   'Papas a la Huancaina','Choclo con Queso','Pisco Sour','Chicha Morada'];
    v_item_prices int[] := ARRAY[28500,32000,22000,24500,18000,42000,35000,48000,38000,39500,52000,14500,12000,18500,8500];
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
    v_methods     text[] := ARRAY['card','card','card','card','cash','card','other'];
    v_loc_id      uuid;
    v_prefix      text;
    v_order_num   int := 10000;
    v_dow         int;
    v_hour        int;
    v_rand        float;
    i int; j int; k int;
BEGIN
    FOR v_day IN SELECT generate_series(CURRENT_DATE - 30, CURRENT_DATE - 1, '1 day')::date LOOP
        v_dow := EXTRACT(DOW FROM v_day)::int;
        FOR k IN 1..4 LOOP
            v_loc_id := v_locations[k];
            v_prefix := v_loc_prefixes[k];

            -- Build item_ids for this location
            v_item_ids := ARRAY[]::uuid[];
            FOR j IN 1..15 LOOP
                v_item_ids := v_item_ids || (v_prefix || '100001-0001-0001-0001-0000000000' || lpad(j::text, 2, '0'))::uuid;
            END LOOP;

            -- Weekend (Fri=5, Sat=6) higher volume
            IF v_dow IN (5,6) THEN
                v_num_orders := 280 + floor(random()*40)::int;
            ELSE
                v_num_orders := 200 + floor(random()*40)::int;
            END IF;

            FOR i IN 1..v_num_orders LOOP
                v_check_id := gen_random_uuid();
                v_channel := v_channels[1 + floor(random()*15)::int];
                v_order_num := v_order_num + 1;

                -- Time distribution: lunch 40%, dinner 50%, other 10%
                v_rand := random();
                IF v_rand < 0.10 THEN
                    v_hour := 10 + floor(random()*1)::int; -- 10-10:59
                ELSIF v_rand < 0.50 THEN
                    v_hour := 11 + floor(random()*4)::int; -- 11-14:59 (lunch)
                ELSE
                    v_hour := 18 + floor(random()*5)::int; -- 18-22:59 (dinner)
                END IF;
                v_opened := v_day + (v_hour * interval '1 hour') + (random() * interval '59 minutes');

                v_subtotal := 0;
                v_num_items := 1 + floor(random()*3)::int;

                INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel,
                    subtotal, tax, total, tip, discount, opened_at, closed_at, source, created_at)
                VALUES (v_check_id, v_org_id, v_loc_id, 'ORD-' || v_order_num, 'closed', v_channel,
                    0, 0, 0, 0, 0, v_opened, v_opened + interval '25 minutes' + (random() * interval '20 minutes'),
                    'manual', v_opened);

                FOR j IN 1..v_num_items LOOP
                    v_item_idx := 1 + floor(random() * 15)::int;
                    v_qty := 1 + floor(random()*2)::int;
                    v_price := v_item_prices[v_item_idx];
                    v_subtotal := v_subtotal + (v_price * v_qty);

                    INSERT INTO check_items (org_id, check_id, menu_item_id, name, quantity, unit_price, created_at)
                    VALUES (v_org_id, v_check_id, v_item_ids[v_item_idx], v_item_names[v_item_idx], v_qty, v_price, v_opened);
                END LOOP;

                v_tax := (v_subtotal * 0.14)::int; -- Egypt VAT 14%
                v_total := v_subtotal + v_tax;
                v_tip := CASE WHEN v_channel = 'dine_in' THEN (v_subtotal * (0.10 + random()*0.10))::int ELSE 0 END;
                v_method := v_methods[1 + floor(random()*7)::int];

                UPDATE checks SET subtotal = v_subtotal, tax = v_tax, total = v_total, tip = v_tip
                WHERE check_id = v_check_id;

                INSERT INTO payments (org_id, check_id, amount, tip, method, status, created_at)
                VALUES (v_org_id, v_check_id, v_total, v_tip, v_method, 'completed', v_opened + interval '30 minutes');
            END LOOP;
        END LOOP;
    END LOOP;
    RAISE NOTICE 'Orders seeded: % orders generated', v_order_num - 10000;
END $$;

COMMIT;
-- ============================================================================
-- CHICHA EGYPT Part 3: Inventory, Waste, POs, Budgets, Shifts, Staff Points
-- ============================================================================
SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

BEGIN;

-- ============================================================================
-- STEP 7: INVENTORY COUNTS (4 per branch = 16 total, weekly)
-- ============================================================================
INSERT INTO inventory_counts (count_id, org_id, location_id, counted_by, count_type, status, started_at, submitted_at, approved_by, approved_at, created_at, updated_at) VALUES
-- El Gouna
('fc110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'e1a00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 28, CURRENT_DATE - 28, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 28, CURRENT_DATE - 28, CURRENT_DATE - 28),
('fc110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'e1a00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 21, CURRENT_DATE - 21, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 21, CURRENT_DATE - 21, CURRENT_DATE - 21),
('fc110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'e1a00001-0001-0001-0001-000000000012', 'full', 'approved', CURRENT_DATE - 14, CURRENT_DATE - 14, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14, CURRENT_DATE - 14, CURRENT_DATE - 14),
('fc110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'e1a00001-0001-0001-0001-000000000013', 'full', 'approved', CURRENT_DATE - 7, CURRENT_DATE - 7, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7, CURRENT_DATE - 7, CURRENT_DATE - 7),
-- New Cairo
('fc110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'e1b00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 28, CURRENT_DATE - 28, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 28, CURRENT_DATE - 28, CURRENT_DATE - 28),
('fc110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'e1b00001-0001-0001-0001-000000000012', 'full', 'approved', CURRENT_DATE - 21, CURRENT_DATE - 21, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 21, CURRENT_DATE - 21, CURRENT_DATE - 21),
('fc110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'e1b00001-0001-0001-0001-000000000013', 'full', 'approved', CURRENT_DATE - 14, CURRENT_DATE - 14, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14, CURRENT_DATE - 14, CURRENT_DATE - 14),
('fc110001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'e1b00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 7, CURRENT_DATE - 7, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7, CURRENT_DATE - 7, CURRENT_DATE - 7),
-- Sheikh Zayed
('fc110001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'e1c00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 28, CURRENT_DATE - 28, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 28, CURRENT_DATE - 28, CURRENT_DATE - 28),
('fc110001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'e1c00001-0001-0001-0001-000000000012', 'full', 'approved', CURRENT_DATE - 21, CURRENT_DATE - 21, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 21, CURRENT_DATE - 21, CURRENT_DATE - 21),
('fc110001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'e1c00001-0001-0001-0001-000000000013', 'full', 'approved', CURRENT_DATE - 14, CURRENT_DATE - 14, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14, CURRENT_DATE - 14, CURRENT_DATE - 14),
('fc110001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'e1c00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 7, CURRENT_DATE - 7, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7, CURRENT_DATE - 7, CURRENT_DATE - 7),
-- North Coast
('fc110001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'e1d00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 28, CURRENT_DATE - 28, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 28, CURRENT_DATE - 28, CURRENT_DATE - 28),
('fc110001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'e1d00001-0001-0001-0001-000000000012', 'full', 'approved', CURRENT_DATE - 21, CURRENT_DATE - 21, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 21, CURRENT_DATE - 21, CURRENT_DATE - 21),
('fc110001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'e1d00001-0001-0001-0001-000000000013', 'full', 'approved', CURRENT_DATE - 14, CURRENT_DATE - 14, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14, CURRENT_DATE - 14, CURRENT_DATE - 14),
('fc110001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'e1d00001-0001-0001-0001-000000000011', 'full', 'approved', CURRENT_DATE - 7, CURRENT_DATE - 7, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 7, CURRENT_DATE - 7, CURRENT_DATE - 7);

-- Count lines (5 key ingredients per count, 16 counts = 80 lines)
INSERT INTO inventory_count_lines (org_id, count_id, location_id, ingredient_id, expected_qty, counted_qty, unit, note)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', c.count_id, c.location_id,
  i.ingredient_id,
  CASE WHEN i.name = 'Sea Bass Fillet' THEN 40.0
       WHEN i.name = 'Beef Tenderloin' THEN 35.0
       WHEN i.name = 'Chicken Thighs' THEN 30.0
       WHEN i.name = 'Jumbo Shrimp' THEN 25.0
       WHEN i.name = 'Avocado' THEN 60.0
  END * (0.95 + random()*0.10),
  CASE WHEN i.name = 'Sea Bass Fillet' THEN 40.0
       WHEN i.name = 'Beef Tenderloin' THEN 35.0
       WHEN i.name = 'Chicken Thighs' THEN 30.0
       WHEN i.name = 'Jumbo Shrimp' THEN 25.0
       WHEN i.name = 'Avocado' THEN 60.0
  END * (0.88 + random()*0.12),
  i.unit,
  CASE WHEN random() > 0.7 THEN 'Some variance noted' ELSE NULL END
FROM inventory_counts c
CROSS JOIN ingredients i
WHERE c.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
  AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
  AND i.name IN ('Sea Bass Fillet','Beef Tenderloin','Chicken Thighs','Jumbo Shrimp','Avocado');

-- ============================================================================
-- STEP 8: WASTE LOGS (80+ across all branches, 20 per branch)
-- ============================================================================
INSERT INTO waste_logs (org_id, location_id, ingredient_id, quantity, unit, reason, logged_by, logged_at, note) VALUES
-- El Gouna (20 logs)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000001', 2.5, 'lb', 'expired', 'e1a00001-0001-0001-0001-000000000007', CURRENT_DATE - 28 + time '09:00', 'Sea bass past use-by'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000011', 4.0, 'ea', 'expired', 'e1a00001-0001-0001-0001-000000000011', CURRENT_DATE - 27 + time '08:30', 'Avocados overripe'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000003', 1.5, 'lb', 'overcooked', 'e1a00001-0001-0001-0001-000000000008', CURRENT_DATE - 26 + time '13:15', 'Chicken dried out on brasa'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000002', 1.0, 'lb', 'dropped', 'e1a00001-0001-0001-0001-000000000009', CURRENT_DATE - 25 + time '12:00', 'Tenderloin dropped during rush'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000010', 2.0, 'bunch', 'expired', 'e1a00001-0001-0001-0001-000000000011', CURRENT_DATE - 24 + time '09:30', 'Cilantro wilted'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000004', 1.5, 'lb', 'contaminated', 'e1a00001-0001-0001-0001-000000000007', CURRENT_DATE - 23 + time '10:00', 'Shrimp temp exceeded safe zone'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000001', 1.0, 'lb', 'overproduction', 'e1a00001-0001-0001-0001-000000000008', CURRENT_DATE - 22 + time '21:00', 'End of night ceviche prep'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000006', 3.0, 'lb', 'expired', 'e1a00001-0001-0001-0001-000000000012', CURRENT_DATE - 20 + time '09:00', 'Potatoes sprouting'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000008', 1.0, 'lb', 'dropped', 'e1a00001-0001-0001-0001-000000000009', CURRENT_DATE - 18 + time '12:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000002', 0.5, 'lb', 'overcooked', 'e1a00001-0001-0001-0001-000000000007', CURRENT_DATE - 16 + time '19:45', 'Churrasco overcooked'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000015', 6.0, 'oz', 'dropped', 'e1a00001-0001-0001-0001-000000000014', CURRENT_DATE - 15 + time '20:00', 'Pisco bottle broken'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000012', 3.0, 'ea', 'expired', 'e1a00001-0001-0001-0001-000000000011', CURRENT_DATE - 14 + time '08:45', 'Corn not fresh'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000001', 1.5, 'lb', 'expired', 'e1a00001-0001-0001-0001-000000000007', CURRENT_DATE - 12 + time '09:15', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000009', 5.0, 'ea', 'expired', 'e1a00001-0001-0001-0001-000000000012', CURRENT_DATE - 10 + time '09:00', 'Limes dried out'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000005', 0.5, 'lb', 'overcooked', 'e1a00001-0001-0001-0001-000000000008', CURRENT_DATE - 8 + time '19:30', 'Anticucho charred'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000013', 2.0, 'oz', 'expired', 'e1a00001-0001-0001-0001-000000000011', CURRENT_DATE - 6 + time '10:00', 'Aji paste oxidized'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000011', 3.0, 'ea', 'expired', 'e1a00001-0001-0001-0001-000000000012', CURRENT_DATE - 4 + time '09:30', 'Avocados brown'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000003', 1.0, 'lb', 'overproduction', 'e1a00001-0001-0001-0001-000000000007', CURRENT_DATE - 3 + time '21:30', 'Slow Tuesday night'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000019', 2.0, 'lb', 'overproduction', 'e1a00001-0001-0001-0001-000000000009', CURRENT_DATE - 2 + time '21:00', 'Too much rice prepped'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '1a100001-0001-0001-0001-000000000004', 0.75, 'lb', 'expired', 'e1a00001-0001-0001-0001-000000000008', CURRENT_DATE - 1 + time '09:00', 'Shrimp past date'),
-- New Cairo (20 logs)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000001', 3.0, 'lb', 'expired', 'e1b00001-0001-0001-0001-000000000007', CURRENT_DATE - 27 + time '09:00', 'Sea bass not fresh'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000002', 1.5, 'lb', 'overcooked', 'e1b00001-0001-0001-0001-000000000008', CURRENT_DATE - 26 + time '13:00', 'Tenderloin over-seared'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000011', 5.0, 'ea', 'expired', 'e1b00001-0001-0001-0001-000000000011', CURRENT_DATE - 25 + time '08:30', 'Avocados overripe batch'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000010', 3.0, 'bunch', 'expired', 'e1b00001-0001-0001-0001-000000000012', CURRENT_DATE - 23 + time '09:00', 'Cilantro wilted overnight'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000004', 2.0, 'lb', 'contaminated', 'e1b00001-0001-0001-0001-000000000009', CURRENT_DATE - 22 + time '10:00', 'Shrimp cross-contamination'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000003', 2.0, 'lb', 'overproduction', 'e1b00001-0001-0001-0001-000000000007', CURRENT_DATE - 20 + time '21:00', 'Over-prepped for quiet night'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000006', 2.0, 'lb', 'expired', 'e1b00001-0001-0001-0001-000000000011', CURRENT_DATE - 18 + time '09:30', 'Potatoes soft'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000001', 2.0, 'lb', 'dropped', 'e1b00001-0001-0001-0001-000000000008', CURRENT_DATE - 16 + time '12:30', 'Plate dropped'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000002', 0.75, 'lb', 'overcooked', 'e1b00001-0001-0001-0001-000000000009', CURRENT_DATE - 14 + time '19:45', 'Lomo over-seared'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000012', 4.0, 'ea', 'expired', 'e1b00001-0001-0001-0001-000000000012', CURRENT_DATE - 12 + time '09:00', 'Corn past prime'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000009', 8.0, 'ea', 'expired', 'e1b00001-0001-0001-0001-000000000011', CURRENT_DATE - 10 + time '09:00', 'Limes dried'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000005', 1.0, 'lb', 'overcooked', 'e1b00001-0001-0001-0001-000000000007', CURRENT_DATE - 8 + time '19:30', 'Anticucho burned'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000015', 4.0, 'oz', 'dropped', 'e1b00001-0001-0001-0001-000000000014', CURRENT_DATE - 7 + time '20:00', 'Glass broke'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000008', 1.5, 'lb', 'dropped', 'e1b00001-0001-0001-0001-000000000008', CURRENT_DATE - 6 + time '12:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000013', 3.0, 'oz', 'expired', 'e1b00001-0001-0001-0001-000000000011', CURRENT_DATE - 5 + time '10:00', 'Aji paste discolored'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000019', 3.0, 'lb', 'overproduction', 'e1b00001-0001-0001-0001-000000000009', CURRENT_DATE - 4 + time '21:30', 'Excess rice'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000011', 3.0, 'ea', 'expired', 'e1b00001-0001-0001-0001-000000000012', CURRENT_DATE - 3 + time '09:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000001', 1.5, 'lb', 'overproduction', 'e1b00001-0001-0001-0001-000000000007', CURRENT_DATE - 2 + time '21:00', 'Ceviche prep excess'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000003', 1.0, 'lb', 'dropped', 'e1b00001-0001-0001-0001-000000000008', CURRENT_DATE - 1 + time '12:45', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '1a100001-0001-0001-0001-000000000014', 0.25, 'lb', 'dropped', 'e1b00001-0001-0001-0001-000000000009', CURRENT_DATE - 1 + time '19:00', 'Rocoto fell'),
-- Sheikh Zayed (20 logs)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000001', 2.0, 'lb', 'expired', 'e1c00001-0001-0001-0001-000000000007', CURRENT_DATE - 28 + time '09:00', 'Sea bass past date'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000011', 3.0, 'ea', 'expired', 'e1c00001-0001-0001-0001-000000000011', CURRENT_DATE - 26 + time '08:30', 'Avocados brown'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000002', 1.0, 'lb', 'overcooked', 'e1c00001-0001-0001-0001-000000000008', CURRENT_DATE - 24 + time '13:15', 'Churrasco overdone'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000004', 1.0, 'lb', 'contaminated', 'e1c00001-0001-0001-0001-000000000009', CURRENT_DATE - 22 + time '10:00', 'Shrimp temp issue'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000010', 1.5, 'bunch', 'expired', 'e1c00001-0001-0001-0001-000000000011', CURRENT_DATE - 20 + time '09:00', 'Cilantro wilted'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000003', 1.5, 'lb', 'overproduction', 'e1c00001-0001-0001-0001-000000000007', CURRENT_DATE - 18 + time '21:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000001', 1.5, 'lb', 'dropped', 'e1c00001-0001-0001-0001-000000000008', CURRENT_DATE - 16 + time '12:30', 'Plate dropped'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000006', 2.5, 'lb', 'expired', 'e1c00001-0001-0001-0001-000000000012', CURRENT_DATE - 14 + time '09:00', 'Potatoes gone bad'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000005', 0.75, 'lb', 'overcooked', 'e1c00001-0001-0001-0001-000000000009', CURRENT_DATE - 12 + time '19:45', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000009', 6.0, 'ea', 'expired', 'e1c00001-0001-0001-0001-000000000011', CURRENT_DATE - 10 + time '09:00', 'Limes dried'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000002', 0.5, 'lb', 'dropped', 'e1c00001-0001-0001-0001-000000000007', CURRENT_DATE - 8 + time '19:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000015', 3.0, 'oz', 'dropped', 'e1c00001-0001-0001-0001-000000000014', CURRENT_DATE - 7 + time '20:00', 'Glass shattered'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000012', 2.0, 'ea', 'expired', 'e1c00001-0001-0001-0001-000000000012', CURRENT_DATE - 5 + time '08:45', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000019', 2.5, 'lb', 'overproduction', 'e1c00001-0001-0001-0001-000000000009', CURRENT_DATE - 4 + time '21:30', 'Excess rice'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000008', 0.5, 'lb', 'dropped', 'e1c00001-0001-0001-0001-000000000008', CURRENT_DATE - 3 + time '12:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000011', 2.0, 'ea', 'expired', 'e1c00001-0001-0001-0001-000000000011', CURRENT_DATE - 2 + time '09:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000003', 0.75, 'lb', 'dropped', 'e1c00001-0001-0001-0001-000000000007', CURRENT_DATE - 1 + time '12:45', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000004', 0.5, 'lb', 'expired', 'e1c00001-0001-0001-0001-000000000009', CURRENT_DATE - 1 + time '09:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000013', 1.5, 'oz', 'expired', 'e1c00001-0001-0001-0001-000000000011', CURRENT_DATE - 1 + time '10:00', 'Paste oxidized'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '1a100001-0001-0001-0001-000000000007', 1.0, 'lb', 'expired', 'e1c00001-0001-0001-0001-000000000012', CURRENT_DATE - 1 + time '09:30', 'Sweet potatoes soft'),
-- North Coast (20 logs)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000001', 3.5, 'lb', 'expired', 'e1d00001-0001-0001-0001-000000000007', CURRENT_DATE - 28 + time '09:00', 'Sea bass spoiled in transport'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000002', 2.0, 'lb', 'overcooked', 'e1d00001-0001-0001-0001-000000000008', CURRENT_DATE - 26 + time '13:15', 'New cook training error'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000011', 6.0, 'ea', 'expired', 'e1d00001-0001-0001-0001-000000000011', CURRENT_DATE - 25 + time '08:30', 'Avocados overripe'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000004', 2.5, 'lb', 'contaminated', 'e1d00001-0001-0001-0001-000000000009', CURRENT_DATE - 23 + time '10:00', 'Cold chain broke during delivery'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000010', 2.0, 'bunch', 'expired', 'e1d00001-0001-0001-0001-000000000012', CURRENT_DATE - 22 + time '09:00', 'Cilantro not stored properly'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000003', 2.0, 'lb', 'overproduction', 'e1d00001-0001-0001-0001-000000000007', CURRENT_DATE - 20 + time '21:00', 'Rain killed turnout'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000001', 2.0, 'lb', 'dropped', 'e1d00001-0001-0001-0001-000000000008', CURRENT_DATE - 18 + time '12:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000006', 3.0, 'lb', 'expired', 'e1d00001-0001-0001-0001-000000000011', CURRENT_DATE - 16 + time '09:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000002', 1.0, 'lb', 'dropped', 'e1d00001-0001-0001-0001-000000000009', CURRENT_DATE - 14 + time '19:00', 'Dropped during plating'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000012', 3.0, 'ea', 'expired', 'e1d00001-0001-0001-0001-000000000012', CURRENT_DATE - 12 + time '08:45', 'Corn not fresh'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000015', 8.0, 'oz', 'dropped', 'e1d00001-0001-0001-0001-000000000014', CURRENT_DATE - 10 + time '20:00', 'Pisco bottle dropped'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000005', 1.0, 'lb', 'overcooked', 'e1d00001-0001-0001-0001-000000000007', CURRENT_DATE - 8 + time '19:30', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000009', 10.0, 'ea', 'expired', 'e1d00001-0001-0001-0001-000000000011', CURRENT_DATE - 6 + time '09:00', 'Large batch of limes'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000008', 1.0, 'lb', 'dropped', 'e1d00001-0001-0001-0001-000000000008', CURRENT_DATE - 5 + time '12:15', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000013', 2.0, 'oz', 'expired', 'e1d00001-0001-0001-0001-000000000012', CURRENT_DATE - 4 + time '10:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000019', 3.0, 'lb', 'overproduction', 'e1d00001-0001-0001-0001-000000000009', CURRENT_DATE - 3 + time '21:00', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000011', 4.0, 'ea', 'expired', 'e1d00001-0001-0001-0001-000000000011', CURRENT_DATE - 2 + time '09:30', 'Avocados brown'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000003', 1.5, 'lb', 'dropped', 'e1d00001-0001-0001-0001-000000000007', CURRENT_DATE - 1 + time '12:45', NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000004', 1.0, 'lb', 'expired', 'e1d00001-0001-0001-0001-000000000008', CURRENT_DATE - 1 + time '09:00', 'Shrimp past date'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '1a100001-0001-0001-0001-000000000014', 0.5, 'lb', 'dropped', 'e1d00001-0001-0001-0001-000000000009', CURRENT_DATE - 1 + time '18:30', NULL);

-- ============================================================================
-- STEP 9: PURCHASE ORDERS (12 total, 3 per branch)
-- ============================================================================
-- Received POs
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, received_by, received_at, total_estimated, total_actual, notes, created_at, updated_at) VALUES
('da110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco Egypt', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 18, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 16, 1250000, 1285000, 'Weekly protein order', CURRENT_DATE - 20, CURRENT_DATE - 16),
('da110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Metro Market', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 14, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 12, 980000, 1020000, 'Bi-weekly restock', CURRENT_DATE - 16, CURRENT_DATE - 12),
('da110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Sysco Egypt', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 12, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 10, 1180000, 1210000, 'Weekly order', CURRENT_DATE - 14, CURRENT_DATE - 10),
('da110001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Metro Market', 'received', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 10, '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 8, 1350000, 1395000, 'High season stock-up', CURRENT_DATE - 12, CURRENT_DATE - 8);
-- Approved POs
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, approved_by, approved_at, total_estimated, total_actual, notes, created_at, updated_at) VALUES
('da110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Seoudi Fresh', 'approved', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 2, 450000, 0, 'Produce order', CURRENT_DATE - 3, CURRENT_DATE - 2),
('da110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Sysco Egypt', 'approved', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1, 1100000, 0, 'Weekly protein', CURRENT_DATE - 2, CURRENT_DATE - 1),
('da110001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Metro Market', 'approved', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1, 520000, 0, 'Produce and dairy', CURRENT_DATE - 2, CURRENT_DATE - 1),
('da110001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Sysco Egypt', 'approved', 'manual', '0d55e810-1e4a-417a-8a70-08b98f4595c2', CURRENT_DATE - 1, 950000, 0, 'Weekend prep', CURRENT_DATE - 2, CURRENT_DATE - 1);
-- Draft POs
INSERT INTO purchase_orders (purchase_order_id, org_id, location_id, vendor_name, status, source, total_estimated, total_actual, notes, created_at, updated_at) VALUES
('da110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 'draft', 'system_recommended', 680000, 0, 'Auto-generated: Pisco and aji paste reorder', CURRENT_DATE, CURRENT_DATE),
('da110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Seoudi Fresh', 'draft', 'system_recommended', 380000, 0, 'Auto-generated: produce reorder', CURRENT_DATE, CURRENT_DATE),
('da110001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Specialty Imports', 'draft', 'system_recommended', 720000, 0, 'Auto: specialty ingredients low', CURRENT_DATE, CURRENT_DATE),
('da110001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Seoudi Fresh', 'draft', 'system_recommended', 420000, 0, 'Auto: produce running low', CURRENT_DATE, CURRENT_DATE);

-- PO Lines for received orders (simplified)
INSERT INTO purchase_order_lines (org_id, purchase_order_id, ingredient_id, ordered_qty, ordered_unit, estimated_unit_cost, received_qty, received_unit_cost, variance_qty, variance_flag, received_at) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000001', 50.0, 'lb', 4500, 48.0, 4600, -2.0, 'short', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000002', 40.0, 'lb', 6500, 40.0, 6700, 0.0, 'exact', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000001', '1a100001-0001-0001-0001-000000000004', 30.0, 'lb', 5200, 28.0, 5300, -2.0, 'short', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000004', '1a100001-0001-0001-0001-000000000001', 45.0, 'lb', 4500, 45.0, 4500, 0.0, 'exact', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000004', '1a100001-0001-0001-0001-000000000003', 60.0, 'lb', 1200, 58.0, 1250, -2.0, 'short', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000007', '1a100001-0001-0001-0001-000000000002', 35.0, 'lb', 6500, 35.0, 6500, 0.0, 'exact', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000007', '1a100001-0001-0001-0001-000000000004', 25.0, 'lb', 5200, 24.0, 5400, -1.0, 'short', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000010', '1a100001-0001-0001-0001-000000000001', 55.0, 'lb', 4500, 52.0, 4800, -3.0, 'short', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'da110001-0001-0001-0001-000000000010', '1a100001-0001-0001-0001-000000000002', 45.0, 'lb', 6500, 45.0, 6500, 0.0, 'exact', CURRENT_DATE - 8);

-- ============================================================================
-- STEP 10: BUDGETS (monthly + weekly for all 4 branches)
-- ============================================================================
INSERT INTO budgets (org_id, location_id, period_type, period_start, period_end, revenue_target, food_cost_pct_target, labor_cost_pct_target, cogs_target, created_by) VALUES
-- Monthly (revenue in piasters: 300M piasters = 3M EGP)
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'monthly', '2026-03-01', '2026-03-31', 300000000, 32.00, 28.00, 96000000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'monthly', '2026-03-01', '2026-03-31', 300000000, 32.00, 28.00, 96000000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'monthly', '2026-03-01', '2026-03-31', 300000000, 32.00, 28.00, 96000000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'monthly', '2026-03-01', '2026-03-31', 300000000, 32.00, 28.00, 96000000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
-- Weekly
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'weekly', date_trunc('week', CURRENT_DATE)::date, (date_trunc('week', CURRENT_DATE) + interval '6 days')::date, 70000000, 32.00, 28.00, 22400000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'weekly', date_trunc('week', CURRENT_DATE)::date, (date_trunc('week', CURRENT_DATE) + interval '6 days')::date, 70000000, 32.00, 28.00, 22400000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'weekly', date_trunc('week', CURRENT_DATE)::date, (date_trunc('week', CURRENT_DATE) + interval '6 days')::date, 70000000, 32.00, 28.00, 22400000, '0d55e810-1e4a-417a-8a70-08b98f4595c2'),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'weekly', date_trunc('week', CURRENT_DATE)::date, (date_trunc('week', CURRENT_DATE) + interval '6 days')::date, 70000000, 32.00, 28.00, 22400000, '0d55e810-1e4a-417a-8a70-08b98f4595c2');

-- ============================================================================
-- STEP 11: SHIFTS (using PL/pgSQL for all 4 branches)
-- 30 employees x ~5 shifts/week x 4 weeks per branch
-- ============================================================================
DO $$
DECLARE
    v_org_id    uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_locations uuid[] := ARRAY[
        'a1111111-1111-1111-1111-111111111111'::uuid,
        'b2222222-2222-2222-2222-222222222222'::uuid,
        'c3333333-3333-3333-3333-333333333333'::uuid,
        'd4444444-4444-4444-4444-444444444444'::uuid
    ];
    v_emp_prefixes text[] := ARRAY['e1a','e1b','e1c','e1d'];
    v_loc_id    uuid;
    v_emp_prefix text;
    v_day       date;
    v_dow       int;
    v_emp_id    uuid;
    v_role      text;
    v_rate      int;
    v_clock_in  timestamptz;
    v_clock_out timestamptz;
    i int; j int; k int;
BEGIN
    FOR k IN 1..4 LOOP
        v_loc_id := v_locations[k];
        v_emp_prefix := v_emp_prefixes[k];
        FOR v_day IN SELECT generate_series(CURRENT_DATE - 28, CURRENT_DATE - 1, '1 day')::date LOOP
            v_dow := EXTRACT(DOW FROM v_day)::int;
            -- Schedule ~20 of 30 employees per day
            FOR j IN 1..30 LOOP
                v_emp_id := (v_emp_prefix || '00001-0001-0001-0001-0000000000' || lpad(j::text, 2, '0'))::uuid;

                -- Determine role and rate based on position
                IF j = 1 THEN v_role := 'gm'; v_rate := 15000;
                ELSIF j IN (2,3) THEN v_role := 'shift_manager'; v_rate := 10000;
                ELSIF j IN (7,8,9,10) THEN v_role := 'staff'; v_rate := 6000; -- cooks
                ELSIF j IN (4,5,6) THEN v_role := 'staff'; v_rate := 5000; -- servers
                ELSIF j IN (14,15) THEN v_role := 'staff'; v_rate := 5500; -- bartenders
                ELSIF j IN (18,19,20) THEN v_role := 'staff'; v_rate := 3500; -- dishwashers
                ELSE v_role := 'staff'; v_rate := 4500; -- other
                END IF;

                -- GM works Sun-Thu (Egypt work week)
                IF j = 1 AND v_dow NOT IN (5,6) THEN
                    v_clock_in := v_day + interval '9 hours' + (random() * interval '15 minutes');
                    v_clock_out := v_clock_in + interval '9 hours' + (random() * interval '1 hour');
                    INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status, created_at, updated_at)
                    VALUES (v_org_id, v_loc_id, v_emp_id, v_role, v_clock_in, v_clock_out, v_rate, 'completed', v_clock_in, v_clock_out);
                -- Shift managers split shifts
                ELSIF j = 2 AND v_dow NOT IN (5) THEN
                    v_clock_in := v_day + interval '10 hours' + (random() * interval '15 minutes');
                    v_clock_out := v_clock_in + interval '8 hours' + (random() * interval '1 hour');
                    INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status, created_at, updated_at)
                    VALUES (v_org_id, v_loc_id, v_emp_id, v_role, v_clock_in, v_clock_out, v_rate, 'completed', v_clock_in, v_clock_out);
                ELSIF j = 3 AND v_dow NOT IN (0) THEN
                    v_clock_in := v_day + interval '15 hours' + (random() * interval '15 minutes');
                    v_clock_out := v_clock_in + interval '8 hours' + (random() * interval '1 hour');
                    INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status, created_at, updated_at)
                    VALUES (v_org_id, v_loc_id, v_emp_id, v_role, v_clock_in, v_clock_out, v_rate, 'completed', v_clock_in, v_clock_out);
                -- Staff: ~5 days per week, rotating days off
                ELSIF j >= 4 AND j <= 30 THEN
                    -- Skip their day off (based on employee number)
                    IF v_dow = (j % 7) THEN
                        CONTINUE;
                    END IF;
                    -- Skip ~15% randomly for variation
                    IF random() < 0.15 THEN
                        CONTINUE;
                    END IF;

                    IF j <= 13 THEN -- Kitchen/server staff: split between AM and PM
                        IF j % 2 = 0 THEN
                            v_clock_in := v_day + interval '10 hours' + (random() * interval '30 minutes');
                            v_clock_out := v_clock_in + interval '7 hours' + (random() * interval '1 hour');
                        ELSE
                            v_clock_in := v_day + interval '16 hours' + (random() * interval '30 minutes');
                            v_clock_out := v_clock_in + interval '6 hours' + (random() * interval '2 hours');
                        END IF;
                    ELSE -- Support staff
                        v_clock_in := v_day + interval '11 hours' + (random() * interval '2 hours');
                        v_clock_out := v_clock_in + interval '6 hours' + (random() * interval '2 hours');
                    END IF;

                    INSERT INTO shifts (org_id, location_id, employee_id, role, clock_in, clock_out, hourly_rate, status, created_at, updated_at)
                    VALUES (v_org_id, v_loc_id, v_emp_id, v_role, v_clock_in, v_clock_out, v_rate, 'completed', v_clock_in, v_clock_out);
                END IF;
            END LOOP;
        END LOOP;
    END LOOP;
    RAISE NOTICE 'Shifts seeded for all 4 branches';
END $$;

COMMIT;
-- ============================================================================
-- CHICHA EGYPT Part 4: Staff Points, Guests, Kitchen, Vendors, Marketing,
--                      Portfolio, Schedules, KDS, Forecasts
-- ============================================================================
SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

BEGIN;

-- ============================================================================
-- STEP 12: STAFF POINT EVENTS (100+ across all branches)
-- ============================================================================
INSERT INTO staff_point_events (org_id, employee_id, points, reason, description, created_at) VALUES
-- El Gouna staff
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000001', 15.00, 'task_completion', 'Completed monthly inventory audit', CURRENT_DATE - 28),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000001', 20.00, 'attendance', 'Perfect attendance March week 1', CURRENT_DATE - 21),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000001', 10.00, 'peer_nominated', 'Team leadership recognition', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000002', 15.00, 'speed_bonus', 'Fastest shift turnover', CURRENT_DATE - 25),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000002', -5.00, 'late', '10 minutes late', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000002', 20.00, 'task_completion', 'Trained 3 new staff', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000003', 10.00, 'accuracy_bonus', 'Zero cash variance', CURRENT_DATE - 22),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000003', 15.00, 'task_completion', 'Organized storage room', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000007', 25.00, 'speed_bonus', 'Fastest ceviche prep record', CURRENT_DATE - 27),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000007', 15.00, 'peer_nominated', 'Kitchen MVP', CURRENT_DATE - 15),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000007', 20.00, 'attendance', 'Perfect attendance', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000008', 10.00, 'task_completion', 'Deep clean grill station', CURRENT_DATE - 20),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000008', -10.00, 'incomplete_task', 'Did not complete closing checklist', CURRENT_DATE - 13),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000008', 15.00, 'speed_bonus', 'Handled Friday rush solo', CURRENT_DATE - 6),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000009', 20.00, 'task_completion', 'Cross-trained on ceviche bar', CURRENT_DATE - 24),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000009', 10.00, 'accuracy_bonus', 'Perfect portion control', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000011', 15.00, 'task_completion', 'Reorganized walk-in cooler', CURRENT_DATE - 19),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000011', 20.00, 'accuracy_bonus', 'Perfect inventory count', CURRENT_DATE - 7),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000004', 10.00, 'speed_bonus', 'Fastest table turn', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000004', -5.00, 'late', '8 minutes late', CURRENT_DATE - 9),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000004', 15.00, 'peer_nominated', 'Best customer service', CURRENT_DATE - 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000014', 10.00, 'task_completion', 'New cocktail recipe developed', CURRENT_DATE - 17),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000014', 15.00, 'speed_bonus', 'Managed bar rush solo', CURRENT_DATE - 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000018', 10.00, 'attendance', 'On time all month', CURRENT_DATE - 7),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1a00001-0001-0001-0001-000000000005', 15.00, 'task_completion', 'Upselling champion', CURRENT_DATE - 11),
-- New Cairo staff
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000001', 20.00, 'task_completion', 'Monthly report completed early', CURRENT_DATE - 26),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000001', 15.00, 'attendance', 'Perfect attendance', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000001', 10.00, 'peer_nominated', 'Leadership award', CURRENT_DATE - 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000002', 15.00, 'speed_bonus', 'Record ticket times', CURRENT_DATE - 20),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000002', 20.00, 'task_completion', 'Staff scheduling optimization', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000007', 25.00, 'speed_bonus', 'Fastest prep time this month', CURRENT_DATE - 25),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000007', 20.00, 'peer_nominated', 'Kitchen team MVP', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000007', 15.00, 'accuracy_bonus', 'Zero waste on lomo saltado', CURRENT_DATE - 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000008', 10.00, 'task_completion', 'New sauce recipe perfected', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000008', 15.00, 'speed_bonus', 'Handled catering order', CURRENT_DATE - 6),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000009', -5.00, 'late', '12 minutes late', CURRENT_DATE - 22),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000009', 20.00, 'task_completion', 'Cross-trained 2 new hires', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000011', 15.00, 'accuracy_bonus', 'Perfect prep count', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000011', 10.00, 'task_completion', 'Organized dry storage', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000004', 15.00, 'peer_nominated', 'Guest favorite server', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000004', 10.00, 'speed_bonus', 'Record upselling week', CURRENT_DATE - 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000014', 10.00, 'task_completion', 'Pisco tasting menu created', CURRENT_DATE - 11),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1b00001-0001-0001-0001-000000000018', 10.00, 'attendance', 'On time all month', CURRENT_DATE - 7),
-- Sheikh Zayed staff
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000001', 20.00, 'task_completion', 'Cost optimization initiative', CURRENT_DATE - 24),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000001', 15.00, 'attendance', 'Perfect attendance', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000002', 10.00, 'speed_bonus', 'Quick shift transition', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000002', 15.00, 'task_completion', 'Menu training completed', CURRENT_DATE - 6),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000007', 25.00, 'peer_nominated', 'Kitchen excellence award', CURRENT_DATE - 22),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000007', 15.00, 'speed_bonus', 'Record dinner service', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000008', 10.00, 'task_completion', 'New plating style implemented', CURRENT_DATE - 15),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000008', -5.00, 'late', '15 minutes late', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000009', 20.00, 'task_completion', 'Cross-trained on all stations', CURRENT_DATE - 20),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000011', 15.00, 'accuracy_bonus', 'Perfect count 3 weeks straight', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000004', 10.00, 'speed_bonus', 'Table turn record', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000004', 15.00, 'peer_nominated', 'Team spirit award', CURRENT_DATE - 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000014', 10.00, 'task_completion', 'Cocktail menu refresh', CURRENT_DATE - 9),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1c00001-0001-0001-0001-000000000018', 10.00, 'attendance', 'Perfect attendance', CURRENT_DATE - 7),
-- North Coast staff
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000001', 20.00, 'task_completion', 'Seasonal menu launch', CURRENT_DATE - 26),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000001', 15.00, 'peer_nominated', 'Leadership recognition', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000002', 15.00, 'speed_bonus', 'Fastest service record', CURRENT_DATE - 20),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000002', 10.00, 'task_completion', 'Staff training session', CURRENT_DATE - 4),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000007', 25.00, 'speed_bonus', 'Ceviche bar sprint record', CURRENT_DATE - 24),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000007', 15.00, 'accuracy_bonus', 'Zero waste week', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000008', -5.00, 'late', '20 minutes late', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000008', 20.00, 'task_completion', 'Menu item innovation', CURRENT_DATE - 6),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000009', 10.00, 'task_completion', 'Cross-trained on grill', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000009', 15.00, 'peer_nominated', 'Most improved', CURRENT_DATE - 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000011', 15.00, 'accuracy_bonus', 'Perfect prep portions', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000004', 10.00, 'speed_bonus', 'Best server this week', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000004', -10.00, 'incomplete_task', 'Table not properly reset', CURRENT_DATE - 3),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000014', 10.00, 'task_completion', 'Beach cocktail menu designed', CURRENT_DATE - 11),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000018', 10.00, 'attendance', 'Full month attendance', CURRENT_DATE - 7),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'e1d00001-0001-0001-0001-000000000005', 15.00, 'peer_nominated', 'Best team player', CURRENT_DATE - 5);

-- Update staff_points totals
UPDATE employees SET staff_points = COALESCE((
    SELECT SUM(points) FROM staff_point_events WHERE staff_point_events.employee_id = employees.employee_id
), staff_points)
WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

-- ============================================================================
-- STEP 13: GUEST PROFILES (50 guests with Arabic names)
-- ============================================================================
INSERT INTO guest_profiles (guest_id, org_id, payment_token_hash, privacy_tier, first_name, email, phone, total_visits, total_spend, avg_check, preferred_channel, favorite_items, clv_score, segment, churn_risk, churn_probability, next_visit_predicted, last_visit_at) VALUES
-- Champions (8)
('aa110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_001', 3, 'Khaled El-Masry', 'khaled.m@email.com', '+20-100-555-0101', 32, 2880000, 90000, 'dine_in', '["Ceviche Clasico","Lomo Saltado"]', 4200.00, 'champion', 'low', 0.0500, CURRENT_DATE + 3, CURRENT_DATE - 2),
('aa110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_002', 3, 'Nadia Farouk', 'nadia.f@email.com', '+20-101-555-0102', 28, 2520000, 90000, 'dine_in', '["Tiradito Nikkei","Churrasco Chimichurri"]', 3800.00, 'champion', 'low', 0.0800, CURRENT_DATE + 5, CURRENT_DATE - 4),
('aa110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_003', 3, 'Amr Abdel-Rahman', 'amr.a@email.com', '+20-102-555-0103', 35, 3150000, 90000, 'delivery', '["Arroz con Mariscos","Pisco Sour"]', 4600.00, 'champion', 'low', 0.0300, CURRENT_DATE + 2, CURRENT_DATE - 1),
('aa110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_004', 3, 'Layla Samir', 'layla.s@email.com', '+20-100-555-0104', 24, 2400000, 100000, 'dine_in', '["Churrasco Chimichurri","Causa Limena"]', 3500.00, 'champion', 'low', 0.0600, CURRENT_DATE + 4, CURRENT_DATE - 3),
('aa110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_005', 3, 'Tarek Helmy', 'tarek.h@email.com', '+20-101-555-0105', 30, 2850000, 95000, 'dine_in', '["Lomo Saltado","Pisco Sour"]', 4100.00, 'champion', 'low', 0.0400, CURRENT_DATE + 3, CURRENT_DATE - 2),
('aa110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_006', 3, 'Dina Ashraf', 'dina.a@email.com', '+20-102-555-0106', 22, 2200000, 100000, 'dine_in', '["Ceviche Clasico","Anticuchos"]', 3200.00, 'champion', 'low', 0.0700, CURRENT_DATE + 6, CURRENT_DATE - 5),
('aa110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_007', 3, 'Hossam Galal', 'hossam.g@email.com', '+20-100-555-0107', 26, 2470000, 95000, 'dine_in', '["Arroz con Mariscos","Tiradito"]', 3600.00, 'champion', 'low', 0.0500, CURRENT_DATE + 4, CURRENT_DATE - 3),
('aa110001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ch_008', 3, 'Mona Youssef', 'mona.y@email.com', '+20-101-555-0108', 20, 1900000, 95000, 'takeout', '["Aji de Gallina","Chicha Morada"]', 2800.00, 'champion', 'low', 0.0900, CURRENT_DATE + 7, CURRENT_DATE - 6),
-- Loyal Regulars (12)
('aa110001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_001', 2, 'Mohamed Ezz', 'mo.ezz@email.com', '+20-100-555-0201', 15, 1200000, 80000, 'dine_in', '["Lomo Saltado"]', 1800.00, 'loyal_regular', 'low', 0.1200, CURRENT_DATE + 7, CURRENT_DATE - 5),
('aa110001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_002', 2, 'Sara Nabil', NULL, '+20-101-555-0202', 12, 960000, 80000, 'takeout', '["Pollo a la Brasa","Papas Huancaina"]', 1500.00, 'loyal_regular', 'low', 0.1500, CURRENT_DATE + 6, CURRENT_DATE - 6),
('aa110001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_003', 2, 'Omar Hassan', 'omar.h@email.com', '+20-102-555-0203', 14, 1050000, 75000, 'dine_in', '["Ceviche Clasico","Chicha Morada"]', 1600.00, 'loyal_regular', 'medium', 0.2000, CURRENT_DATE + 10, CURRENT_DATE - 9),
('aa110001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_004', 2, 'Yasmine Adel', 'yasmine.a@email.com', NULL, 10, 850000, 85000, 'delivery', '["Empanadas","Causa Limena"]', 1300.00, 'loyal_regular', 'low', 0.1000, CURRENT_DATE + 4, CURRENT_DATE - 3),
('aa110001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_005', 2, 'Hassan Mahmoud', 'hassan.m@email.com', '+20-100-555-0205', 11, 935000, 85000, 'dine_in', '["Seco de Res","Pisco Sour"]', 1400.00, 'loyal_regular', 'low', 0.1100, CURRENT_DATE + 8, CURRENT_DATE - 7),
('aa110001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_006', 2, 'Fatma Ibrahim', 'fatma.i@email.com', '+20-101-555-0206', 13, 1040000, 80000, 'dine_in', '["Tiradito Nikkei"]', 1550.00, 'loyal_regular', 'low', 0.1300, CURRENT_DATE + 5, CURRENT_DATE - 4),
('aa110001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_007', 2, 'Karim Sobhy', NULL, '+20-102-555-0207', 9, 720000, 80000, 'takeout', '["Pollo a la Brasa"]', 1100.00, 'loyal_regular', 'medium', 0.1800, CURRENT_DATE + 9, CURRENT_DATE - 8),
('aa110001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_008', 2, 'Heba Saleh', 'heba.s@email.com', '+20-100-555-0208', 16, 1280000, 80000, 'dine_in', '["Anticuchos","Choclo con Queso"]', 1900.00, 'loyal_regular', 'low', 0.0800, CURRENT_DATE + 3, CURRENT_DATE - 2),
('aa110001-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_009', 2, 'Nour El-Din', 'nour.e@email.com', '+20-101-555-0209', 11, 880000, 80000, 'delivery', '["Arroz con Mariscos"]', 1350.00, 'loyal_regular', 'low', 0.1400, CURRENT_DATE + 6, CURRENT_DATE - 5),
('aa110001-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_010', 2, 'Reem Atef', 'reem.a@email.com', '+20-102-555-0210', 8, 640000, 80000, 'dine_in', '["Causa Limena","Pisco Sour"]', 980.00, 'loyal_regular', 'medium', 0.2200, CURRENT_DATE + 12, CURRENT_DATE - 11),
('aa110001-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_011', 2, 'Sherif Badawy', 'sherif.b@email.com', '+20-100-555-0211', 10, 850000, 85000, 'dine_in', '["Churrasco Chimichurri"]', 1300.00, 'loyal_regular', 'low', 0.1000, CURRENT_DATE + 5, CURRENT_DATE - 4),
('aa110001-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_lr_012', 2, 'Aya Mostafa', NULL, '+20-101-555-0212', 7, 560000, 80000, 'takeout', '["Empanadas","Chicha Morada"]', 860.00, 'loyal_regular', 'medium', 0.2500, CURRENT_DATE + 14, CURRENT_DATE - 12),
-- At Risk (10) - high CLV but lapsed
('aa110001-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_001', 2, 'Waleed El-Sharkawy', 'waleed.s@email.com', '+20-100-555-0301', 20, 1800000, 90000, 'dine_in', '["Ceviche Clasico","Lomo Saltado"]', 2700.00, 'at_risk', 'high', 0.6500, CURRENT_DATE - 5, CURRENT_DATE - 25),
('aa110001-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_002', 2, 'Noha Galal', 'noha.g@email.com', '+20-101-555-0302', 16, 1440000, 90000, 'dine_in', '["Arroz con Mariscos","Pisco Sour"]', 2200.00, 'at_risk', 'high', 0.7000, CURRENT_DATE - 10, CURRENT_DATE - 32),
('aa110001-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_003', 1, 'Essam Fathy', NULL, '+20-102-555-0303', 18, 1620000, 90000, 'takeout', '["Churrasco Chimichurri","Tiradito"]', 2500.00, 'at_risk', 'critical', 0.8200, CURRENT_DATE - 15, CURRENT_DATE - 45),
('aa110001-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_004', 2, 'Ghada Ramzy', 'ghada.r@email.com', '+20-100-555-0304', 14, 1260000, 90000, 'delivery', '["Pollo a la Brasa"]', 1900.00, 'at_risk', 'medium', 0.4500, CURRENT_DATE + 2, CURRENT_DATE - 22),
('aa110001-0001-0001-0001-000000000025', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_005', 2, 'Tamer Shawky', 'tamer.s@email.com', '+20-101-555-0305', 12, 1080000, 90000, 'dine_in', '["Lomo Saltado"]', 1650.00, 'at_risk', 'high', 0.5800, CURRENT_DATE - 3, CURRENT_DATE - 28),
('aa110001-0001-0001-0001-000000000026', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_006', 2, 'Amal Kamel', 'amal.k@email.com', '+20-102-555-0306', 15, 1350000, 90000, 'dine_in', '["Ceviche Clasico"]', 2050.00, 'at_risk', 'high', 0.6200, CURRENT_DATE - 8, CURRENT_DATE - 30),
('aa110001-0001-0001-0001-000000000027', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_007', 1, 'Basem Fouad', NULL, '+20-100-555-0307', 10, 900000, 90000, 'delivery', '["Aji de Gallina"]', 1380.00, 'at_risk', 'medium', 0.5000, CURRENT_DATE + 1, CURRENT_DATE - 20),
('aa110001-0001-0001-0001-000000000028', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_008', 2, 'Sawsan Ismail', 'sawsan.i@email.com', '+20-101-555-0308', 8, 720000, 90000, 'dine_in', '["Empanadas","Pisco Sour"]', 1100.00, 'at_risk', 'medium', 0.4800, CURRENT_DATE + 3, CURRENT_DATE - 24),
('aa110001-0001-0001-0001-000000000029', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_009', 2, 'Medhat Anwar', 'medhat.a@email.com', '+20-102-555-0309', 11, 990000, 90000, 'dine_in', '["Seco de Res"]', 1520.00, 'at_risk', 'high', 0.7500, CURRENT_DATE - 12, CURRENT_DATE - 40),
('aa110001-0001-0001-0001-000000000030', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ar_010', 2, 'Ingy Lotfy', 'ingy.l@email.com', '+20-100-555-0310', 9, 810000, 90000, 'takeout', '["Causa Limena","Chicha Morada"]', 1240.00, 'at_risk', 'high', 0.6800, CURRENT_DATE - 7, CURRENT_DATE - 35),
-- New Discoverers (7)
('aa110001-0001-0001-0001-000000000031', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_nd_001', 1, 'Ziad', NULL, NULL, 2, 160000, 80000, 'dine_in', '["Ceviche Clasico"]', 240.00, 'new_discoverer', 'medium', 0.3500, CURRENT_DATE + 12, CURRENT_DATE - 8),
('aa110001-0001-0001-0001-000000000032', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_nd_002', 1, 'Perihan', NULL, NULL, 1, 95000, 95000, 'takeout', '["Lomo Saltado"]', 145.00, 'new_discoverer', 'medium', 0.4000, CURRENT_DATE + 20, CURRENT_DATE - 12),
('aa110001-0001-0001-0001-000000000033', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_nd_003', 1, 'Ismail', NULL, NULL, 3, 225000, 75000, 'delivery', '["Empanadas","Chicha Morada"]', 340.00, 'new_discoverer', 'low', 0.2500, CURRENT_DATE + 6, CURRENT_DATE - 5),
('aa110001-0001-0001-0001-000000000034', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_nd_004', 1, 'Rehab', NULL, NULL, 2, 180000, 90000, 'dine_in', '["Tiradito Nikkei"]', 275.00, 'new_discoverer', 'medium', 0.3200, CURRENT_DATE + 15, CURRENT_DATE - 10),
('aa110001-0001-0001-0001-000000000035', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_nd_005', 1, 'Hazem', NULL, NULL, 1, 110000, 110000, 'dine_in', '["Churrasco Chimichurri"]', 170.00, 'new_discoverer', 'high', 0.4500, CURRENT_DATE + 25, CURRENT_DATE - 15),
('aa110001-0001-0001-0001-000000000036', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_nd_006', 1, 'Malak', NULL, NULL, 2, 145000, 72500, 'takeout', '["Pollo a la Brasa"]', 220.00, 'new_discoverer', 'medium', 0.3800, CURRENT_DATE + 18, CURRENT_DATE - 11),
('aa110001-0001-0001-0001-000000000037', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_nd_007', 1, 'Adham', NULL, NULL, 3, 210000, 70000, 'dine_in', '["Ceviche Clasico","Papas Huancaina"]', 320.00, 'new_discoverer', 'low', 0.2000, CURRENT_DATE + 8, CURRENT_DATE - 4),
-- Casual (13)
('aa110001-0001-0001-0001-000000000038', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_001', 1, NULL, NULL, NULL, 5, 375000, 75000, 'dine_in', '["Ceviche Clasico"]', 570.00, 'casual', 'medium', 0.3000, CURRENT_DATE + 15, CURRENT_DATE - 18),
('aa110001-0001-0001-0001-000000000039', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_002', 1, NULL, NULL, NULL, 4, 300000, 75000, 'takeout', '["Empanadas"]', 460.00, 'casual', 'medium', 0.3500, CURRENT_DATE + 20, CURRENT_DATE - 22),
('aa110001-0001-0001-0001-000000000040', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_003', 1, NULL, NULL, NULL, 6, 480000, 80000, 'dine_in', '["Lomo Saltado"]', 730.00, 'casual', 'low', 0.2000, CURRENT_DATE + 10, CURRENT_DATE - 12),
('aa110001-0001-0001-0001-000000000041', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_004', 1, NULL, NULL, NULL, 3, 225000, 75000, 'delivery', '["Pollo a la Brasa"]', 345.00, 'casual', 'high', 0.5500, CURRENT_DATE + 25, CURRENT_DATE - 30),
('aa110001-0001-0001-0001-000000000042', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_005', 1, NULL, NULL, NULL, 4, 340000, 85000, 'dine_in', '["Pisco Sour"]', 520.00, 'casual', 'medium', 0.3200, CURRENT_DATE + 18, CURRENT_DATE - 15),
('aa110001-0001-0001-0001-000000000043', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_006', 1, NULL, NULL, NULL, 5, 400000, 80000, 'dine_in', '["Tiradito Nikkei"]', 610.00, 'casual', 'low', 0.2500, CURRENT_DATE + 12, CURRENT_DATE - 10),
('aa110001-0001-0001-0001-000000000044', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_007', 1, NULL, NULL, NULL, 3, 270000, 90000, 'takeout', '["Seco de Res"]', 415.00, 'casual', 'medium', 0.4000, CURRENT_DATE + 22, CURRENT_DATE - 20),
('aa110001-0001-0001-0001-000000000045', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_008', 1, NULL, NULL, NULL, 6, 510000, 85000, 'dine_in', '["Anticuchos"]', 780.00, 'casual', 'low', 0.1800, CURRENT_DATE + 8, CURRENT_DATE - 6),
('aa110001-0001-0001-0001-000000000046', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_009', 1, NULL, NULL, NULL, 2, 170000, 85000, 'delivery', '["Aji de Gallina"]', 260.00, 'casual', 'high', 0.5000, CURRENT_DATE + 28, CURRENT_DATE - 25),
('aa110001-0001-0001-0001-000000000047', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_010', 1, NULL, NULL, NULL, 4, 320000, 80000, 'dine_in', '["Choclo con Queso"]', 490.00, 'casual', 'medium', 0.3500, CURRENT_DATE + 16, CURRENT_DATE - 14),
('aa110001-0001-0001-0001-000000000048', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_011', 1, NULL, NULL, NULL, 3, 255000, 85000, 'takeout', '["Causa Limena"]', 390.00, 'casual', 'medium', 0.3800, CURRENT_DATE + 20, CURRENT_DATE - 18),
('aa110001-0001-0001-0001-000000000049', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_012', 1, NULL, NULL, NULL, 5, 425000, 85000, 'dine_in', '["Papas Huancaina"]', 650.00, 'casual', 'low', 0.2200, CURRENT_DATE + 10, CURRENT_DATE - 8),
('aa110001-0001-0001-0001-000000000050', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'tok_ca_013', 1, NULL, NULL, NULL, 4, 360000, 90000, 'dine_in', '["Arroz con Mariscos"]', 550.00, 'casual', 'medium', 0.2800, CURRENT_DATE + 14, CURRENT_DATE - 12);

-- ============================================================================
-- STEP 14: KITCHEN STATIONS + KDS (6 per branch = 24)
-- ============================================================================
INSERT INTO kitchen_stations (station_id, org_id, location_id, name, station_type, max_concurrent, status) VALUES
-- El Gouna
('ad110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Grill Station', 'grill', 4, 'active'),
('ad110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ceviche Bar', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Saute Station', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Prep Station', 'prep', 4, 'active'),
('ad110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Expo Station', 'expo', 2, 'active'),
('ad110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Dish Pit', 'dish', 2, 'active'),
-- New Cairo
('ad110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Grill Station', 'grill', 4, 'active'),
('ad110001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Ceviche Bar', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Saute Station', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Prep Station', 'prep', 4, 'active'),
('ad110001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Expo Station', 'expo', 2, 'active'),
('ad110001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Dish Pit', 'dish', 2, 'active'),
-- Sheikh Zayed
('ad110001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Grill Station', 'grill', 4, 'active'),
('ad110001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Ceviche Bar', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Saute Station', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000016', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Prep Station', 'prep', 4, 'active'),
('ad110001-0001-0001-0001-000000000017', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Expo Station', 'expo', 2, 'active'),
('ad110001-0001-0001-0001-000000000018', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Dish Pit', 'dish', 2, 'active'),
-- North Coast
('ad110001-0001-0001-0001-000000000019', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Grill Station', 'grill', 4, 'active'),
('ad110001-0001-0001-0001-000000000020', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Ceviche Bar', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000021', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Saute Station', 'saute', 3, 'active'),
('ad110001-0001-0001-0001-000000000022', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Prep Station', 'prep', 4, 'active'),
('ad110001-0001-0001-0001-000000000023', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Expo Station', 'expo', 2, 'active'),
('ad110001-0001-0001-0001-000000000024', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Dish Pit', 'dish', 2, 'active');

-- Resource profiles for El Gouna menu items (representative)
INSERT INTO menu_item_resource_profiles (org_id, menu_item_id, station_type, task_sequence, duration_secs, elu_required, batch_size) VALUES
-- Ceviche: prep -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000001', 'prep', 1, 120, 1.00, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000001', 'expo', 2, 30, 0.25, 1),
-- Tiradito: prep -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000002', 'prep', 1, 90, 0.75, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000002', 'expo', 2, 30, 0.25, 1),
-- Lomo Saltado: prep -> saute -> grill -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000006', 'prep', 1, 60, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000006', 'saute', 2, 180, 1.00, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000006', 'grill', 3, 120, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000006', 'expo', 4, 45, 0.50, 1),
-- Churrasco: grill -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000011', 'grill', 1, 480, 1.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000011', 'expo', 2, 60, 0.50, 1),
-- Pollo a la Brasa: grill -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000009', 'grill', 1, 600, 1.00, 2),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000009', 'expo', 2, 45, 0.50, 1),
-- Pisco Sour: prep -> expo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000014', 'prep', 1, 60, 0.50, 1),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa100001-0001-0001-0001-000000000014', 'expo', 2, 15, 0.25, 1);

-- KDS Tickets (8 active across branches)
INSERT INTO checks (check_id, org_id, location_id, order_number, status, channel, subtotal, tax, total, tip, discount, opened_at, source) VALUES
('ab110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'KDS-101', 'open', 'dine_in', 60500, 8470, 68970, 0, 0, NOW() - interval '8 minutes', 'manual'),
('ab110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'KDS-102', 'open', 'dine_in', 94000, 13160, 107160, 0, 0, NOW() - interval '5 minutes', 'manual'),
('ab110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'KDS-201', 'open', 'takeout', 46500, 6510, 53010, 0, 0, NOW() - interval '3 minutes', 'manual'),
('ab110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'KDS-202', 'open', 'delivery', 80000, 11200, 91200, 0, 0, NOW() - interval '12 minutes', 'manual'),
('ab110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'KDS-301', 'open', 'dine_in', 52000, 7280, 59280, 0, 0, NOW() - interval '6 minutes', 'manual'),
('ab110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'KDS-302', 'open', 'dine_in', 28500, 3990, 32490, 0, 0, NOW() - interval '2 minutes', 'manual'),
('ab110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'KDS-401', 'open', 'dine_in', 70000, 9800, 79800, 0, 0, NOW() - interval '10 minutes', 'manual'),
('ab110001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'KDS-402', 'open', 'takeout', 38000, 5320, 43320, 0, 0, NOW() - interval '4 minutes', 'manual');

INSERT INTO kds_tickets (ticket_id, org_id, location_id, check_id, order_number, channel, status, priority, estimated_ready_at, created_at, updated_at) VALUES
('ac110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ab110001-0001-0001-0001-000000000001', 'KDS-101', 'dine_in', 'in_progress', 0, NOW() + interval '8 minutes', NOW() - interval '8 minutes', NOW()),
('ac110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'ab110001-0001-0001-0001-000000000002', 'KDS-102', 'dine_in', 'in_progress', 0, NOW() + interval '12 minutes', NOW() - interval '5 minutes', NOW()),
('ac110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'ab110001-0001-0001-0001-000000000003', 'KDS-201', 'takeout', 'new', 0, NOW() + interval '15 minutes', NOW() - interval '3 minutes', NOW()),
('ac110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'ab110001-0001-0001-0001-000000000004', 'KDS-202', 'delivery', 'in_progress', 1, NOW() + interval '5 minutes', NOW() - interval '12 minutes', NOW()),
('ac110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'ab110001-0001-0001-0001-000000000005', 'KDS-301', 'dine_in', 'in_progress', 0, NOW() + interval '10 minutes', NOW() - interval '6 minutes', NOW()),
('ac110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'ab110001-0001-0001-0001-000000000006', 'KDS-302', 'dine_in', 'new', 0, NOW() + interval '18 minutes', NOW() - interval '2 minutes', NOW()),
('ac110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'ab110001-0001-0001-0001-000000000007', 'KDS-401', 'dine_in', 'in_progress', 0, NOW() + interval '6 minutes', NOW() - interval '10 minutes', NOW()),
('ac110001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'ab110001-0001-0001-0001-000000000008', 'KDS-402', 'takeout', 'new', 0, NOW() + interval '16 minutes', NOW() - interval '4 minutes', NOW());

INSERT INTO kds_ticket_items (org_id, ticket_id, menu_item_id, item_name, quantity, station_type, status, fire_at, started_at, completed_at, duration_secs) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000001', 'aa100001-0001-0001-0001-000000000001', 'Ceviche Clasico', 1, 'prep', 'cooking', NOW() - interval '8 minutes', NOW() - interval '5 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000001', 'aa100001-0001-0001-0001-000000000006', 'Lomo Saltado', 1, 'grill', 'cooking', NOW() - interval '8 minutes', NOW() - interval '6 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000002', 'aa100001-0001-0001-0001-000000000011', 'Churrasco Chimichurri', 2, 'grill', 'cooking', NOW() - interval '5 minutes', NOW() - interval '3 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000002', 'aa100001-0001-0001-0001-000000000014', 'Pisco Sour', 2, 'prep', 'ready', NOW() - interval '5 minutes', NOW() - interval '4 minutes', NOW() - interval '2 minutes', 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000003', 'bb100001-0001-0001-0001-000000000009', 'Pollo a la Brasa', 1, 'grill', 'pending', NULL, NULL, NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000004', 'bb100001-0001-0001-0001-000000000008', 'Arroz con Mariscos', 2, 'saute', 'cooking', NOW() - interval '12 minutes', NOW() - interval '10 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000005', 'cc100001-0001-0001-0001-000000000011', 'Churrasco Chimichurri', 1, 'grill', 'cooking', NOW() - interval '6 minutes', NOW() - interval '4 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000006', 'cc100001-0001-0001-0001-000000000001', 'Ceviche Clasico', 1, 'prep', 'pending', NULL, NULL, NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000007', 'dd100001-0001-0001-0001-000000000006', 'Lomo Saltado', 1, 'saute', 'cooking', NOW() - interval '10 minutes', NOW() - interval '8 minutes', NULL, NULL),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000007', 'dd100001-0001-0001-0001-000000000002', 'Tiradito Nikkei', 1, 'prep', 'ready', NOW() - interval '10 minutes', NOW() - interval '9 minutes', NOW() - interval '6 minutes', 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ac110001-0001-0001-0001-000000000008', 'dd100001-0001-0001-0001-000000000009', 'Pollo a la Brasa', 1, 'grill', 'pending', NULL, NULL, NULL, NULL);

-- ============================================================================
-- STEP 15: VENDOR SCORES + PRICE HISTORY
-- ============================================================================
INSERT INTO vendor_scores (org_id, location_id, vendor_name, overall_score, price_score, delivery_score, quality_score, accuracy_score, total_orders, otif_rate, on_time_rate, in_full_rate, avg_lead_days, calculated_at) VALUES
-- Scores for all 4 locations x 4 vendors = 16
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Sysco Egypt', 82.00, 75.00, 88.00, 84.00, 81.00, 28, 85.00, 90.00, 88.00, 2.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Metro Market', 78.00, 80.00, 72.00, 82.00, 78.00, 15, 72.00, 78.00, 72.00, 1.50, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Seoudi Fresh', 85.00, 72.00, 90.00, 92.00, 86.00, 20, 88.00, 92.00, 90.00, 1.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Specialty Imports', 75.00, 60.00, 70.00, 90.00, 80.00, 8, 75.00, 72.00, 78.00, 5.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Sysco Egypt', 80.00, 74.00, 85.00, 82.00, 79.00, 25, 83.00, 88.00, 85.00, 2.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Metro Market', 84.00, 82.00, 80.00, 88.00, 86.00, 18, 82.00, 85.00, 82.00, 1.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Seoudi Fresh', 86.00, 74.00, 92.00, 90.00, 88.00, 22, 90.00, 94.00, 92.00, 1.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Specialty Imports', 73.00, 58.00, 68.00, 88.00, 78.00, 6, 70.00, 68.00, 72.00, 5.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Sysco Egypt', 81.00, 73.00, 86.00, 83.00, 82.00, 22, 84.00, 89.00, 86.00, 2.50, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Metro Market', 83.00, 81.00, 78.00, 88.00, 85.00, 16, 80.00, 82.00, 80.00, 1.50, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Seoudi Fresh', 84.00, 70.00, 88.00, 94.00, 84.00, 18, 86.00, 90.00, 88.00, 1.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Specialty Imports', 76.00, 62.00, 72.00, 90.00, 80.00, 7, 74.00, 74.00, 76.00, 5.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Sysco Egypt', 79.00, 72.00, 82.00, 80.00, 82.00, 20, 80.00, 85.00, 82.00, 3.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Metro Market', 72.00, 78.00, 68.00, 76.00, 66.00, 14, 72.00, 70.00, 68.00, 2.00, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Seoudi Fresh', 80.00, 68.00, 84.00, 90.00, 78.00, 16, 82.00, 86.00, 84.00, 1.50, NOW()),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'Specialty Imports', 70.00, 55.00, 65.00, 88.00, 72.00, 5, 68.00, 65.00, 70.00, 6.00, NOW());

-- Price history (30+ entries)
INSERT INTO ingredient_price_history (org_id, ingredient_id, vendor_name, unit_cost, quantity, source, recorded_at) VALUES
-- Sea Bass trending up
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000001', 'Sysco Egypt', 3800, 50.0, 'po_received', CURRENT_DATE - 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000001', 'Sysco Egypt', 4000, 48.0, 'po_received', CURRENT_DATE - 150),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000001', 'Sysco Egypt', 4200, 50.0, 'po_received', CURRENT_DATE - 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000001', 'Sysco Egypt', 4350, 45.0, 'po_received', CURRENT_DATE - 90),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000001', 'Sysco Egypt', 4500, 50.0, 'po_received', CURRENT_DATE - 60),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000001', 'Sysco Egypt', 4600, 48.0, 'po_received', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000001', 'Sysco Egypt', 4800, 52.0, 'po_received', CURRENT_DATE - 8),
-- Beef stable
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000002', 'Sysco Egypt', 6200, 40.0, 'po_received', CURRENT_DATE - 150),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000002', 'Sysco Egypt', 6300, 40.0, 'po_received', CURRENT_DATE - 90),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000002', 'Sysco Egypt', 6500, 40.0, 'po_received', CURRENT_DATE - 30),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000002', 'Sysco Egypt', 6700, 45.0, 'po_received', CURRENT_DATE - 8),
-- Shrimp seasonal
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000004', 'Metro Market', 4500, 30.0, 'po_received', CURRENT_DATE - 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000004', 'Metro Market', 4800, 28.0, 'po_received', CURRENT_DATE - 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000004', 'Metro Market', 5200, 25.0, 'po_received', CURRENT_DATE - 60),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000004', 'Metro Market', 5400, 24.0, 'po_received', CURRENT_DATE - 10),
-- Pisco trending up
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000015', 'Specialty Imports', 160, 200.0, 'po_received', CURRENT_DATE - 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000015', 'Specialty Imports', 175, 180.0, 'po_received', CURRENT_DATE - 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000015', 'Specialty Imports', 190, 200.0, 'po_received', CURRENT_DATE - 60),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000015', 'Specialty Imports', 200, 150.0, 'po_received', CURRENT_DATE - 16),
-- Aji amarillo
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000013', 'Specialty Imports', 65, 100.0, 'po_received', CURRENT_DATE - 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000013', 'Specialty Imports', 72, 90.0, 'po_received', CURRENT_DATE - 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000013', 'Specialty Imports', 80, 100.0, 'po_received', CURRENT_DATE - 60),
-- Avocado seasonal
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000011', 'Seoudi Fresh', 250, 80.0, 'po_received', CURRENT_DATE - 180),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000011', 'Seoudi Fresh', 280, 70.0, 'po_received', CURRENT_DATE - 120),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000011', 'Seoudi Fresh', 320, 60.0, 'po_received', CURRENT_DATE - 60),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000011', 'Seoudi Fresh', 350, 65.0, 'po_received', CURRENT_DATE - 16),
-- Chicken stable
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000003', 'Seoudi Fresh', 1100, 60.0, 'po_received', CURRENT_DATE - 150),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000003', 'Seoudi Fresh', 1150, 60.0, 'po_received', CURRENT_DATE - 90),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000003', 'Seoudi Fresh', 1200, 58.0, 'po_received', CURRENT_DATE - 30),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000003', 'Seoudi Fresh', 1250, 60.0, 'po_received', CURRENT_DATE - 12),
-- Rice
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000019', 'Metro Market', 130, 100.0, 'po_received', CURRENT_DATE - 150),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000019', 'Metro Market', 140, 100.0, 'po_received', CURRENT_DATE - 90),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', '1a100001-0001-0001-0001-000000000019', 'Metro Market', 150, 100.0, 'po_received', CURRENT_DATE - 30);

-- ============================================================================
-- STEP 16: MARKETING CAMPAIGNS + LOYALTY
-- ============================================================================
INSERT INTO campaigns (campaign_id, org_id, location_id, name, campaign_type, status, target_segment, channel, discount_type, discount_value, min_purchase, start_at, end_at, recurring, recurrence_rule, redemptions, revenue_attributed, cost_of_promotion) VALUES
('ca110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Pisco Hour', 'happy_hour', 'active', 'casual', 'all', 'percentage', 20.00, 10000, CURRENT_DATE - 14, CURRENT_DATE + 14, true, 'FREQ=WEEKLY;BYDAY=SU,MO,TU,WE,TH', 85, 1575000, 315000),
('ca110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', 'Ceviche Festival', 'bogo', 'completed', 'loyal_regular', 'email', 'bogo', NULL, 15000, CURRENT_DATE - 45, CURRENT_DATE - 15, false, NULL, 220, 3960000, 792000),
('ca110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', 'Loyalty Launch', 'loyalty_reward', 'active', 'champion', 'push', 'dollar_off', 50.00, 20000, CURRENT_DATE - 30, CURRENT_DATE + 60, false, NULL, 45, 900000, 225000),
('ca110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', 'North Coast Summer', 'happy_hour', 'draft', 'casual', 'all', 'percentage', 15.00, 15000, '2026-06-01', '2026-09-30', true, 'FREQ=DAILY', 0, 0, 0),
('ca110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', 'Ramadan Special', 'loyalty_reward', 'draft', 'loyal_regular', 'push', 'dollar_off', 30.00, 25000, NULL, NULL, false, NULL, 0, 0, 0);

-- Loyalty Members (15)
INSERT INTO loyalty_members (member_id, org_id, guest_id, points_balance, lifetime_points, tier, joined_at) VALUES
('ae110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000001', 2400.00, 8500.00, 'platinum', CURRENT_DATE - 180),
('ae110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000002', 1800.00, 7200.00, 'platinum', CURRENT_DATE - 150),
('ae110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000003', 3200.00, 9800.00, 'platinum', CURRENT_DATE - 200),
('ae110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000004', 1500.00, 6000.00, 'gold', CURRENT_DATE - 160),
('ae110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000005', 2100.00, 7800.00, 'platinum', CURRENT_DATE - 170),
('ae110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000009', 800.00, 3200.00, 'gold', CURRENT_DATE - 120),
('ae110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000010', 550.00, 2400.00, 'silver', CURRENT_DATE - 100),
('ae110001-0001-0001-0001-000000000008', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000011', 900.00, 2800.00, 'gold', CURRENT_DATE - 90),
('ae110001-0001-0001-0001-000000000009', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000013', 650.00, 2200.00, 'silver', CURRENT_DATE - 80),
('ae110001-0001-0001-0001-000000000010', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000016', 1200.00, 3800.00, 'gold', CURRENT_DATE - 130),
('ae110001-0001-0001-0001-000000000011', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000021', 400.00, 5200.00, 'gold', CURRENT_DATE - 160),
('ae110001-0001-0001-0001-000000000012', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000022', 300.00, 4400.00, 'gold', CURRENT_DATE - 140),
('ae110001-0001-0001-0001-000000000013', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000033', 150.00, 150.00, 'bronze', CURRENT_DATE - 30),
('ae110001-0001-0001-0001-000000000014', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000037', 180.00, 180.00, 'bronze', CURRENT_DATE - 20),
('ae110001-0001-0001-0001-000000000015', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'aa110001-0001-0001-0001-000000000014', 720.00, 2600.00, 'silver', CURRENT_DATE - 110);

-- Loyalty Transactions (20)
INSERT INTO loyalty_transactions (org_id, member_id, type, points, description, created_at) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000001', 'earn', 450.00, 'Purchase - dinner for four', CURRENT_DATE - 28),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000001', 'earn', 380.00, 'Purchase - weekend brunch', CURRENT_DATE - 21),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000001', 'redeem', -1000.00, 'Redeemed 100 EGP reward', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000001', 'earn', 320.00, 'Purchase - ceviche dinner', CURRENT_DATE - 7),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000002', 'earn', 400.00, 'Purchase - anniversary dinner', CURRENT_DATE - 25),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000002', 'earn', 350.00, 'Purchase - takeout order', CURRENT_DATE - 14),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000002', 'redeem', -500.00, 'Redeemed 50 EGP reward', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000003', 'earn', 500.00, 'Purchase - family delivery', CURRENT_DATE - 22),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000003', 'earn', 420.00, 'Purchase - delivery', CURRENT_DATE - 15),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000003', 'redeem', -2000.00, 'Redeemed 200 EGP reward', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000003', 'earn', 480.00, 'Purchase - large delivery', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000004', 'earn', 350.00, 'Purchase - dinner', CURRENT_DATE - 20),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000006', 'earn', 280.00, 'Purchase', CURRENT_DATE - 15),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000006', 'earn', 250.00, 'Purchase', CURRENT_DATE - 8),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000007', 'earn', 300.00, 'Purchase', CURRENT_DATE - 18),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000007', 'redeem', -200.00, 'Redeemed 20 EGP reward', CURRENT_DATE - 12),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000010', 'earn', 320.00, 'Purchase', CURRENT_DATE - 16),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000010', 'earn', 380.00, 'Purchase', CURRENT_DATE - 6),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000013', 'earn', 150.00, 'Welcome bonus + first purchase', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'ae110001-0001-0001-0001-000000000014', 'earn', 180.00, 'Welcome bonus + first purchase', CURRENT_DATE - 4);

-- ============================================================================
-- STEP 17: PORTFOLIO HIERARCHY
-- ============================================================================
INSERT INTO portfolio_nodes (node_id, org_id, parent_node_id, name, node_type, location_id, sort_order) VALUES
('af110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', NULL, 'Chicha Egypt', 'org', NULL, 0),
('af110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af110001-0001-0001-0001-000000000001', 'Cairo Region', 'region', NULL, 1),
('af110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af110001-0001-0001-0001-000000000001', 'Coast Region', 'region', NULL, 2),
('af110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af110001-0001-0001-0001-000000000002', 'New Cairo', 'location', 'b2222222-2222-2222-2222-222222222222', 1),
('af110001-0001-0001-0001-000000000005', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af110001-0001-0001-0001-000000000002', 'Sheikh Zayed', 'location', 'c3333333-3333-3333-3333-333333333333', 2),
('af110001-0001-0001-0001-000000000006', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af110001-0001-0001-0001-000000000003', 'El Gouna', 'location', 'a1111111-1111-1111-1111-111111111111', 1),
('af110001-0001-0001-0001-000000000007', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'af110001-0001-0001-0001-000000000003', 'North Coast', 'location', 'd4444444-4444-4444-4444-444444444444', 2);

-- Benchmarks
INSERT INTO location_benchmarks (org_id, location_id, period_start, period_end, revenue, food_cost_pct, labor_cost_pct, avg_check_cents, check_count, revenue_percentile, food_cost_percentile, labor_cost_percentile, avg_check_percentile) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', '2026-03-01', '2026-03-31', 624000000, 31.200, 27.500, 86100, 7251, 72.000, 68.000, 65.000, 78.000),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', '2026-03-01', '2026-03-31', 621000000, 33.500, 28.800, 85900, 7223, 68.000, 55.000, 58.000, 76.000),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', '2026-03-01', '2026-03-31', 637000000, 30.800, 26.500, 87200, 7307, 78.000, 72.000, 70.000, 82.000),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', '2026-03-01', '2026-03-31', 627000000, 32.800, 29.200, 86900, 7215, 65.000, 60.000, 52.000, 80.000);

-- Best Practices (3)
INSERT INTO best_practices (org_id, title, description, metric, source_location_id, impact_pct, status, detected_at) VALUES
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Ceviche Bar Prep Scheduling Reduces Fish Waste', 'Sheikh Zayed reduced fish waste by 22% after implementing demand-based ceviche prep scheduling. Morning prep quantities adjusted based on day-of-week booking patterns and seasonal trends.', 'food_waste_pct', 'c3333333-3333-3333-3333-333333333333', 22.000, 'suggested', CURRENT_DATE - 5),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Cross-Training Program Improves Kitchen Flexibility', 'El Gouna cross-trained staff on ceviche bar + grill positions, reducing overtime by 15% while maintaining ticket times. Staff certified on 3+ stations can flex during demand spikes.', 'labor_cost_pct', 'a1111111-1111-1111-1111-111111111111', 15.000, 'adopted', CURRENT_DATE - 10),
('3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'Pisco Sour Batch Pre-Mix Reduces Cocktail Prep Time', 'New Cairo pre-mixes Pisco Sour base in batches, cutting cocktail prep from 90 to 30 seconds. Bar revenue per labor hour increased 18%.', 'labor_cost_pct', 'b2222222-2222-2222-2222-222222222222', 18.000, 'suggested', CURRENT_DATE - 3);

-- ============================================================================
-- STEP 18: SCHEDULES + DEMAND FORECASTS
-- ============================================================================
-- Published schedules for current week at all branches
INSERT INTO schedules (schedule_id, org_id, location_id, week_start, status, created_by, published_at) VALUES
('5a110001-0001-0001-0001-000000000001', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'a1111111-1111-1111-1111-111111111111', date_trunc('week', CURRENT_DATE)::date, 'published', '0d55e810-1e4a-417a-8a70-08b98f4595c2', NOW() - interval '2 days'),
('5a110001-0001-0001-0001-000000000002', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'b2222222-2222-2222-2222-222222222222', date_trunc('week', CURRENT_DATE)::date, 'published', '0d55e810-1e4a-417a-8a70-08b98f4595c2', NOW() - interval '2 days'),
('5a110001-0001-0001-0001-000000000003', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'c3333333-3333-3333-3333-333333333333', date_trunc('week', CURRENT_DATE)::date, 'published', '0d55e810-1e4a-417a-8a70-08b98f4595c2', NOW() - interval '2 days'),
('5a110001-0001-0001-0001-000000000004', '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', 'd4444444-4444-4444-4444-444444444444', date_trunc('week', CURRENT_DATE)::date, 'published', '0d55e810-1e4a-417a-8a70-08b98f4595c2', NOW() - interval '2 days');

-- Scheduled shifts for GM and shift managers at each branch this week
DO $$
DECLARE
    v_org_id uuid := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_schedules uuid[] := ARRAY[
        '5a110001-0001-0001-0001-000000000001'::uuid,
        '5a110001-0001-0001-0001-000000000002'::uuid,
        '5a110001-0001-0001-0001-000000000003'::uuid,
        '5a110001-0001-0001-0001-000000000004'::uuid
    ];
    v_emp_prefixes text[] := ARRAY['e1a','e1b','e1c','e1d'];
    v_week_start date := date_trunc('week', CURRENT_DATE)::date;
    v_day date;
    v_dow int;
    v_stations text[] := ARRAY['expo','grill','prep','saute','grill'];
    i int; j int;
BEGIN
    FOR i IN 1..4 LOOP
        FOR j IN 0..6 LOOP
            v_day := v_week_start + j;
            v_dow := EXTRACT(DOW FROM v_day)::int;
            -- GM: Sun-Thu
            IF v_dow NOT IN (5,6) THEN
                INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
                VALUES (v_org_id, v_schedules[i], (v_emp_prefixes[i] || '00001-0001-0001-0001-000000000001')::uuid, v_day, '09:00', '18:00', 'expo', 'confirmed');
            END IF;
            -- Shift Manager 1: Sun-Thu
            IF v_dow NOT IN (5) THEN
                INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
                VALUES (v_org_id, v_schedules[i], (v_emp_prefixes[i] || '00001-0001-0001-0001-000000000002')::uuid, v_day, '10:00', '19:00', 'grill', 'confirmed');
            END IF;
            -- Shift Manager 2: Mon-Sat
            IF v_dow NOT IN (0) THEN
                INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
                VALUES (v_org_id, v_schedules[i], (v_emp_prefixes[i] || '00001-0001-0001-0001-000000000003')::uuid, v_day, '15:00', '23:00', 'saute', 'confirmed');
            END IF;
            -- 3 cooks: various schedules
            IF v_dow NOT IN (5) THEN
                INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
                VALUES (v_org_id, v_schedules[i], (v_emp_prefixes[i] || '00001-0001-0001-0001-000000000007')::uuid, v_day, '10:00', '18:00', 'grill', 'scheduled');
            END IF;
            IF v_dow NOT IN (0) THEN
                INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
                VALUES (v_org_id, v_schedules[i], (v_emp_prefixes[i] || '00001-0001-0001-0001-000000000008')::uuid, v_day, '16:00', '23:00', 'saute', 'scheduled');
            END IF;
            IF v_dow BETWEEN 1 AND 6 THEN
                INSERT INTO scheduled_shifts (org_id, schedule_id, employee_id, shift_date, start_time, end_time, station, status)
                VALUES (v_org_id, v_schedules[i], (v_emp_prefixes[i] || '00001-0001-0001-0001-000000000009')::uuid, v_day, '11:00', '19:00', 'prep', 'scheduled');
            END IF;
        END LOOP;
    END LOOP;
    RAISE NOTICE 'Scheduled shifts created';
END $$;

-- Demand forecasts for all 4 locations (today + tomorrow)
INSERT INTO labor_demand_forecast (org_id, location_id, forecast_date, time_block, forecasted_covers, required_elu, required_headcount)
SELECT '3f7ef589-f499-43e3-a1c5-aaacd9d543ec', l.location_id, d.forecast_date, t.time_block,
  CASE
    WHEN t.time_block IN ('11:00'::time,'12:00'::time,'13:00'::time) THEN 30 + floor(random()*15)::int
    WHEN t.time_block IN ('19:00'::time,'20:00'::time,'21:00'::time) THEN 40 + floor(random()*20)::int
    WHEN t.time_block IN ('14:00'::time,'15:00'::time) THEN 12 + floor(random()*8)::int
    ELSE 6 + floor(random()*6)::int
  END,
  CASE
    WHEN t.time_block IN ('11:00'::time,'12:00'::time,'13:00'::time) THEN 10.0 + random()*5
    WHEN t.time_block IN ('19:00'::time,'20:00'::time,'21:00'::time) THEN 14.0 + random()*6
    ELSE 3.0 + random()*3
  END,
  CASE
    WHEN t.time_block IN ('11:00'::time,'12:00'::time,'13:00'::time) THEN 4 + floor(random()*2)::int
    WHEN t.time_block IN ('19:00'::time,'20:00'::time,'21:00'::time) THEN 5 + floor(random()*2)::int
    ELSE 2 + floor(random()*2)::int
  END
FROM (VALUES ('a1111111-1111-1111-1111-111111111111'::uuid),('b2222222-2222-2222-2222-222222222222'::uuid),('c3333333-3333-3333-3333-333333333333'::uuid),('d4444444-4444-4444-4444-444444444444'::uuid)) l(location_id)
CROSS JOIN (VALUES (CURRENT_DATE),(CURRENT_DATE + 1)) d(forecast_date)
CROSS JOIN (VALUES ('09:00'::time),('10:00'::time),('11:00'::time),('12:00'::time),('13:00'::time),('14:00'::time),('15:00'::time),('16:00'::time),('17:00'::time),('18:00'::time),('19:00'::time),('20:00'::time),('21:00'::time),('22:00'::time)) t(time_block);

COMMIT;

-- Final summary
SELECT 'checks' AS table_name, COUNT(*) AS row_count FROM checks WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'check_items', COUNT(*) FROM check_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'payments', COUNT(*) FROM payments WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'employees', COUNT(*) FROM employees WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'menu_items', COUNT(*) FROM menu_items WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'ingredients', COUNT(*) FROM ingredients WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'recipes', COUNT(*) FROM recipes WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'recipe_explosion', COUNT(*) FROM recipe_explosion WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'inventory_counts', COUNT(*) FROM inventory_counts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'waste_logs', COUNT(*) FROM waste_logs WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'purchase_orders', COUNT(*) FROM purchase_orders WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'budgets', COUNT(*) FROM budgets WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'shifts', COUNT(*) FROM shifts WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'staff_point_events', COUNT(*) FROM staff_point_events WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'guest_profiles', COUNT(*) FROM guest_profiles WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
UNION ALL SELECT 'kitchen_stations', COUNT(*) FROM kitchen_stations WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
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
UNION ALL SELECT 'locations', COUNT(*) FROM locations WHERE org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'
ORDER BY table_name;
