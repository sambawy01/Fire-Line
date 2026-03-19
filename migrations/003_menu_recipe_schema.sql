-- Menu, recipe, ingredient, check, and raw data log tables

-- ============================================================
-- INGREDIENTS
-- ============================================================
CREATE TABLE ingredients (
    ingredient_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'uncategorized',
    unit TEXT NOT NULL DEFAULT 'oz', -- oz, lb, ea, ml, g, kg
    cost_per_unit INT NOT NULL DEFAULT 0, -- cents per unit
    prep_yield_factor NUMERIC(5,4) NOT NULL DEFAULT 1.0000, -- e.g., 0.83 for onions
    allergens TEXT[] NOT NULL DEFAULT '{}', -- e.g., {'dairy','gluten'}
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ingredients_org ON ingredients(org_id);

ALTER TABLE ingredients ENABLE ROW LEVEL SECURITY;
ALTER TABLE ingredients FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON ingredients
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON ingredients TO fireline_app;

-- Per-location ingredient configuration (different vendors/costs per location)
CREATE TABLE ingredient_location_configs (
    config_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    ingredient_id UUID NOT NULL REFERENCES ingredients(ingredient_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    vendor_name TEXT,
    vendor_item_code TEXT,
    local_cost_per_unit INT, -- cents, overrides ingredient.cost_per_unit
    par_level NUMERIC(10,2), -- target stock level
    reorder_point NUMERIC(10,2), -- trigger reorder when below this
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(ingredient_id, location_id)
);
CREATE INDEX idx_ilc_org ON ingredient_location_configs(org_id);
CREATE INDEX idx_ilc_location ON ingredient_location_configs(org_id, location_id);

ALTER TABLE ingredient_location_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE ingredient_location_configs FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON ingredient_location_configs
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON ingredient_location_configs TO fireline_app;

-- ============================================================
-- MENU ITEMS
-- ============================================================
CREATE TABLE menu_items (
    menu_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    external_id TEXT, -- POS system's ID for this item
    name TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'uncategorized',
    price INT NOT NULL DEFAULT 0, -- cents
    available BOOLEAN NOT NULL DEFAULT true,
    description TEXT,
    source TEXT NOT NULL DEFAULT 'manual', -- 'toast', 'square', 'manual'
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_menu_items_org ON menu_items(org_id);
CREATE INDEX idx_menu_items_location ON menu_items(org_id, location_id);
CREATE INDEX idx_menu_items_external ON menu_items(location_id, external_id) WHERE external_id IS NOT NULL;

ALTER TABLE menu_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE menu_items FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON menu_items
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON menu_items TO fireline_app;

-- ============================================================
-- RECIPES (supports DAG via parent_recipe_id)
-- ============================================================
CREATE TABLE recipes (
    recipe_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    menu_item_id UUID REFERENCES menu_items(menu_item_id), -- null for sub-recipes
    parent_recipe_id UUID REFERENCES recipes(recipe_id), -- null for top-level
    name TEXT NOT NULL,
    yield_quantity NUMERIC(10,2) NOT NULL DEFAULT 1.00,
    yield_unit TEXT NOT NULL DEFAULT 'ea',
    prep_time_minutes INT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_recipes_org ON recipes(org_id);
CREATE INDEX idx_recipes_menu_item ON recipes(menu_item_id) WHERE menu_item_id IS NOT NULL;
CREATE INDEX idx_recipes_parent ON recipes(parent_recipe_id) WHERE parent_recipe_id IS NOT NULL;

ALTER TABLE recipes ENABLE ROW LEVEL SECURITY;
ALTER TABLE recipes FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON recipes
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON recipes TO fireline_app;

-- ============================================================
-- RECIPE INGREDIENTS (junction)
-- ============================================================
CREATE TABLE recipe_ingredients (
    recipe_ingredient_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    recipe_id UUID NOT NULL REFERENCES recipes(recipe_id),
    ingredient_id UUID NOT NULL REFERENCES ingredients(ingredient_id),
    quantity NUMERIC(10,4) NOT NULL, -- amount of ingredient per recipe yield
    unit TEXT NOT NULL DEFAULT 'oz',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ri_org ON recipe_ingredients(org_id);
CREATE INDEX idx_ri_recipe ON recipe_ingredients(recipe_id);

ALTER TABLE recipe_ingredients ENABLE ROW LEVEL SECURITY;
ALTER TABLE recipe_ingredients FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON recipe_ingredients
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON recipe_ingredients TO fireline_app;

-- ============================================================
-- RECIPE EXPLOSION (materialized, updated on recipe change)
-- ============================================================
CREATE TABLE recipe_explosion (
    explosion_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id),
    ingredient_id UUID NOT NULL REFERENCES ingredients(ingredient_id),
    quantity_per_unit NUMERIC(10,6) NOT NULL, -- ingredient qty per 1 menu item sold
    unit TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(menu_item_id, ingredient_id)
);
CREATE INDEX idx_re_org ON recipe_explosion(org_id);
CREATE INDEX idx_re_menu_item ON recipe_explosion(menu_item_id);

ALTER TABLE recipe_explosion ENABLE ROW LEVEL SECURITY;
ALTER TABLE recipe_explosion FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON recipe_explosion
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON recipe_explosion TO fireline_app;

-- ============================================================
-- CHECKS (orders/tickets)
-- ============================================================
CREATE TABLE checks (
    check_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    external_id TEXT, -- POS system's order ID
    order_number TEXT,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'voided')),
    channel TEXT NOT NULL DEFAULT 'dine_in' CHECK (channel IN ('dine_in', 'takeout', 'delivery', 'drive_thru')),
    subtotal INT NOT NULL DEFAULT 0, -- cents
    tax INT NOT NULL DEFAULT 0,
    total INT NOT NULL DEFAULT 0,
    tip INT NOT NULL DEFAULT 0,
    discount INT NOT NULL DEFAULT 0,
    opened_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    closed_at TIMESTAMPTZ,
    source TEXT NOT NULL DEFAULT 'manual',
    raw_payload_id TEXT, -- reference to raw_data_log
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_checks_org ON checks(org_id);
CREATE INDEX idx_checks_location ON checks(org_id, location_id);
CREATE INDEX idx_checks_closed ON checks(org_id, location_id, closed_at DESC) WHERE closed_at IS NOT NULL;
CREATE INDEX idx_checks_external ON checks(location_id, external_id) WHERE external_id IS NOT NULL;

ALTER TABLE checks ENABLE ROW LEVEL SECURITY;
ALTER TABLE checks FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON checks
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE ON checks TO fireline_app;

-- ============================================================
-- CHECK ITEMS (line items)
-- ============================================================
CREATE TABLE check_items (
    check_item_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    check_id UUID NOT NULL REFERENCES checks(check_id),
    menu_item_id UUID REFERENCES menu_items(menu_item_id),
    external_id TEXT,
    name TEXT NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    unit_price INT NOT NULL DEFAULT 0, -- cents
    voided_at TIMESTAMPTZ,
    void_reason TEXT,
    fired_at TIMESTAMPTZ, -- when sent to kitchen
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_check_items_org ON check_items(org_id);
CREATE INDEX idx_check_items_check ON check_items(check_id);

ALTER TABLE check_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE check_items FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON check_items
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE ON check_items TO fireline_app;

-- ============================================================
-- PAYMENTS
-- ============================================================
CREATE TABLE payments (
    payment_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    check_id UUID NOT NULL REFERENCES checks(check_id),
    amount INT NOT NULL DEFAULT 0, -- cents
    tip INT NOT NULL DEFAULT 0,
    method TEXT NOT NULL DEFAULT 'card' CHECK (method IN ('card', 'cash', 'gift_card', 'other')),
    status TEXT NOT NULL DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'refunded', 'voided')),
    external_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_org ON payments(org_id);
CREATE INDEX idx_payments_check ON payments(check_id);

ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON payments
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE ON payments TO fireline_app;

-- ============================================================
-- ITEM ID MAPPING (POS external ID <-> internal ID)
-- ============================================================
CREATE TABLE item_id_mappings (
    mapping_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID NOT NULL REFERENCES locations(location_id),
    adapter_type TEXT NOT NULL,
    external_id TEXT NOT NULL,
    menu_item_id UUID NOT NULL REFERENCES menu_items(menu_item_id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(location_id, adapter_type, external_id)
);
CREATE INDEX idx_item_mapping_org ON item_id_mappings(org_id);

ALTER TABLE item_id_mappings ENABLE ROW LEVEL SECURITY;
ALTER TABLE item_id_mappings FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON item_id_mappings
    USING (org_id = current_setting('app.current_org_id')::UUID);
GRANT SELECT, INSERT, UPDATE, DELETE ON item_id_mappings TO fireline_app;

-- ============================================================
-- RAW DATA LOG (immutable, no RLS — uses separate audit-style access)
-- ============================================================
CREATE TABLE raw_data_log (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id UUID NOT NULL,
    location_id UUID NOT NULL,
    adapter_type TEXT NOT NULL,
    data_type TEXT NOT NULL, -- 'order', 'menu', 'employee', 'webhook'
    external_id TEXT,
    payload JSONB NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_raw_log_org ON raw_data_log(org_id, received_at DESC);
CREATE INDEX idx_raw_log_location ON raw_data_log(location_id, data_type, received_at DESC);

-- No RLS on raw_data_log — immutable audit log
GRANT INSERT, SELECT ON raw_data_log TO fireline_app;
