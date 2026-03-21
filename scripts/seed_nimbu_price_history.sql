-- =============================================================================
-- STEP 9: Ingredient Price History (6 months, 20 entries for key proteins)
-- =============================================================================

SET app.current_org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';

INSERT INTO ingredient_price_history (org_id, ingredient_id, vendor_name, unit_cost, quantity, source, recorded_at)
SELECT
  '3f7ef589-f499-43e3-a1c5-aaacd9d543ec'::uuid,
  i.ingredient_id,
  v.vendor,
  v.cost,
  v.qty,
  'po_received',
  v.rec_at
FROM ingredients i
JOIN (VALUES
  -- Ribeye Steak: trending up from 160K to 180K over 6 months
  ('Ribeye Steak','Premium Meats Co',155000,50.0,now()-interval '180 days'),
  ('Ribeye Steak','Premium Meats Co',158000,45.0,now()-interval '165 days'),
  ('Ribeye Steak','Premium Meats Co',160000,50.0,now()-interval '150 days'),
  ('Ribeye Steak','Premium Meats Co',162000,48.0,now()-interval '135 days'),
  ('Ribeye Steak','Premium Meats Co',163000,50.0,now()-interval '120 days'),
  ('Ribeye Steak','Premium Meats Co',165000,45.0,now()-interval '105 days'),
  ('Ribeye Steak','Premium Meats Co',167000,50.0,now()-interval '90 days'),
  ('Ribeye Steak','Premium Meats Co',170000,48.0,now()-interval '75 days'),
  ('Ribeye Steak','Premium Meats Co',172000,50.0,now()-interval '60 days'),
  ('Ribeye Steak','Premium Meats Co',175000,45.0,now()-interval '45 days'),
  ('Ribeye Steak','Premium Meats Co',176000,50.0,now()-interval '30 days'),
  ('Ribeye Steak','Premium Meats Co',178000,48.0,now()-interval '15 days'),
  ('Ribeye Steak','Premium Meats Co',180000,50.0,now()-interval '3 days'),

  -- Wagyu Beef: stable high with slight increases
  ('Wagyu Beef','Premium Meats Co',320000,20.0,now()-interval '180 days'),
  ('Wagyu Beef','Premium Meats Co',325000,18.0,now()-interval '150 days'),
  ('Wagyu Beef','Premium Meats Co',330000,20.0,now()-interval '120 days'),
  ('Wagyu Beef','Premium Meats Co',335000,18.0,now()-interval '90 days'),
  ('Wagyu Beef','Premium Meats Co',340000,20.0,now()-interval '60 days'),
  ('Wagyu Beef','Premium Meats Co',345000,18.0,now()-interval '30 days'),
  ('Wagyu Beef','Premium Meats Co',350000,20.0,now()-interval '5 days'),

  -- Lobster Tail: seasonal fluctuation
  ('Lobster Tail','Ocean Fresh Egypt',85000,30.0,now()-interval '180 days'),
  ('Lobster Tail','Ocean Fresh Egypt',88000,25.0,now()-interval '150 days'),
  ('Lobster Tail','Ocean Fresh Egypt',92000,30.0,now()-interval '120 days'),
  ('Lobster Tail','Ocean Fresh Egypt',95000,28.0,now()-interval '90 days'),
  ('Lobster Tail','Ocean Fresh Egypt',98000,25.0,now()-interval '60 days'),
  ('Lobster Tail','Ocean Fresh Egypt',95000,30.0,now()-interval '30 days'),
  ('Lobster Tail','Ocean Fresh Egypt',95000,28.0,now()-interval '7 days'),

  -- Salmon Fillet: moderate increase
  ('Salmon Fillet','Ocean Fresh Egypt',125000,40.0,now()-interval '180 days'),
  ('Salmon Fillet','Ocean Fresh Egypt',128000,35.0,now()-interval '150 days'),
  ('Salmon Fillet','Ocean Fresh Egypt',130000,40.0,now()-interval '120 days'),
  ('Salmon Fillet','Ocean Fresh Egypt',132000,38.0,now()-interval '90 days'),
  ('Salmon Fillet','Ocean Fresh Egypt',135000,40.0,now()-interval '60 days'),
  ('Salmon Fillet','Ocean Fresh Egypt',138000,35.0,now()-interval '30 days'),
  ('Salmon Fillet','Ocean Fresh Egypt',140000,40.0,now()-interval '4 days'),

  -- Tuna Sashimi Grade: rising significantly
  ('Tuna Sashimi Grade','Ocean Fresh Egypt',130000,25.0,now()-interval '180 days'),
  ('Tuna Sashimi Grade','Ocean Fresh Egypt',135000,22.0,now()-interval '150 days'),
  ('Tuna Sashimi Grade','Ocean Fresh Egypt',140000,25.0,now()-interval '120 days'),
  ('Tuna Sashimi Grade','Ocean Fresh Egypt',145000,23.0,now()-interval '90 days'),
  ('Tuna Sashimi Grade','Ocean Fresh Egypt',150000,25.0,now()-interval '60 days'),
  ('Tuna Sashimi Grade','Ocean Fresh Egypt',155000,22.0,now()-interval '30 days'),
  ('Tuna Sashimi Grade','Ocean Fresh Egypt',160000,25.0,now()-interval '6 days'),

  -- Lamb Rack: stable
  ('Lamb Rack','Premium Meats Co',190000,30.0,now()-interval '180 days'),
  ('Lamb Rack','Premium Meats Co',192000,28.0,now()-interval '150 days'),
  ('Lamb Rack','Premium Meats Co',194000,30.0,now()-interval '120 days'),
  ('Lamb Rack','Premium Meats Co',196000,28.0,now()-interval '90 days'),
  ('Lamb Rack','Premium Meats Co',197000,30.0,now()-interval '60 days'),
  ('Lamb Rack','Premium Meats Co',199000,28.0,now()-interval '30 days'),
  ('Lamb Rack','Premium Meats Co',200000,30.0,now()-interval '2 days'),

  -- Sea Bass Fillet: slight increase
  ('Sea Bass Fillet','Ocean Fresh Egypt',110000,35.0,now()-interval '180 days'),
  ('Sea Bass Fillet','Ocean Fresh Egypt',112000,32.0,now()-interval '150 days'),
  ('Sea Bass Fillet','Ocean Fresh Egypt',114000,35.0,now()-interval '120 days'),
  ('Sea Bass Fillet','Ocean Fresh Egypt',116000,33.0,now()-interval '90 days'),
  ('Sea Bass Fillet','Ocean Fresh Egypt',118000,35.0,now()-interval '60 days'),
  ('Sea Bass Fillet','Ocean Fresh Egypt',119000,32.0,now()-interval '30 days'),
  ('Sea Bass Fillet','Ocean Fresh Egypt',120000,35.0,now()-interval '5 days'),

  -- Tiger Prawns: moderate fluctuation
  ('Tiger Prawns','Ocean Fresh Egypt',100000,40.0,now()-interval '180 days'),
  ('Tiger Prawns','Ocean Fresh Egypt',102000,38.0,now()-interval '150 days'),
  ('Tiger Prawns','Ocean Fresh Egypt',105000,40.0,now()-interval '120 days'),
  ('Tiger Prawns','Ocean Fresh Egypt',108000,35.0,now()-interval '90 days'),
  ('Tiger Prawns','Ocean Fresh Egypt',110000,40.0,now()-interval '60 days'),
  ('Tiger Prawns','Ocean Fresh Egypt',108000,38.0,now()-interval '30 days'),
  ('Tiger Prawns','Ocean Fresh Egypt',110000,40.0,now()-interval '3 days'),

  -- Duck Leg: stable
  ('Duck Leg','Premium Meats Co',42000,50.0,now()-interval '180 days'),
  ('Duck Leg','Premium Meats Co',43000,48.0,now()-interval '120 days'),
  ('Duck Leg','Premium Meats Co',44000,50.0,now()-interval '60 days'),
  ('Duck Leg','Premium Meats Co',45000,48.0,now()-interval '10 days'),

  -- Chicken Breast: stable low
  ('Chicken Breast','Metro Market',7500,100.0,now()-interval '180 days'),
  ('Chicken Breast','Metro Market',7700,95.0,now()-interval '120 days'),
  ('Chicken Breast','Metro Market',7800,100.0,now()-interval '60 days'),
  ('Chicken Breast','Metro Market',8000,95.0,now()-interval '8 days')
) AS v(ing_name, vendor, cost, qty, rec_at)
ON i.name = v.ing_name AND i.org_id = '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
