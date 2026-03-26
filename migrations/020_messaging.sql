-- Messaging: chat channels and messages for staff communication

-- ============================================================
-- CHAT CHANNELS
-- ============================================================

CREATE TABLE chat_channels (
    channel_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(org_id),
    location_id UUID REFERENCES locations(location_id),
    name        TEXT NOT NULL,
    type        TEXT NOT NULL DEFAULT 'location' CHECK (type IN ('location', 'role', 'direct', 'broadcast')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_chat_channels_org ON chat_channels(org_id);
CREATE INDEX idx_chat_channels_location ON chat_channels(location_id);
CREATE INDEX idx_chat_channels_type ON chat_channels(type);

ALTER TABLE chat_channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE chat_channels FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON chat_channels USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE, DELETE ON chat_channels TO fireline_app;

-- ============================================================
-- CHAT MESSAGES
-- ============================================================

CREATE TABLE chat_messages (
    message_id  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    org_id      UUID NOT NULL REFERENCES organizations(org_id),
    channel_id  UUID NOT NULL REFERENCES chat_channels(channel_id),
    sender_id   UUID NOT NULL,
    sender_name TEXT NOT NULL,
    sender_role TEXT NOT NULL,
    body        TEXT NOT NULL,
    pinned      BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_chat_messages_channel ON chat_messages(channel_id);
CREATE INDEX idx_chat_messages_org ON chat_messages(org_id);
CREATE INDEX idx_chat_messages_created ON chat_messages(created_at);
CREATE INDEX idx_chat_messages_channel_created ON chat_messages(channel_id, created_at DESC);

ALTER TABLE chat_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE chat_messages FORCE ROW LEVEL SECURITY;
CREATE POLICY org_isolation ON chat_messages USING (org_id = current_setting('app.current_org_id')::UUID);

GRANT SELECT, INSERT, UPDATE, DELETE ON chat_messages TO fireline_app;

-- ============================================================
-- SEED DATA
-- for org 3f7ef589-f499-43e3-a1c5-aaacd9d543ec
-- ============================================================

DO $$
DECLARE
    v_org  UUID := '3f7ef589-f499-43e3-a1c5-aaacd9d543ec';
    v_loc_elgouna  UUID := 'a1111111-1111-1111-1111-111111111111';
    v_loc_newcairo UUID := 'b2222222-2222-2222-2222-222222222222';
    v_loc_zayed    UUID := 'c3333333-3333-3333-3333-333333333333';
    v_loc_maadi    UUID := 'd4444444-4444-4444-4444-444444444444';

    v_ch_elgouna  UUID;
    v_ch_newcairo UUID;
    v_ch_zayed    UUID;
    v_ch_maadi    UUID;
    v_ch_allhands UUID;
BEGIN

    -- --------------------------------------------------------
    -- CHANNELS (4 location + 1 broadcast)
    -- --------------------------------------------------------

    INSERT INTO chat_channels (org_id, location_id, name, type)
    VALUES (v_org, v_loc_elgouna, 'El Gouna Team', 'location')
    RETURNING channel_id INTO v_ch_elgouna;

    INSERT INTO chat_channels (org_id, location_id, name, type)
    VALUES (v_org, v_loc_newcairo, 'New Cairo Team', 'location')
    RETURNING channel_id INTO v_ch_newcairo;

    INSERT INTO chat_channels (org_id, location_id, name, type)
    VALUES (v_org, v_loc_zayed, 'Sheikh Zayed Team', 'location')
    RETURNING channel_id INTO v_ch_zayed;

    INSERT INTO chat_channels (org_id, location_id, name, type)
    VALUES (v_org, v_loc_maadi, 'Maadi Team', 'location')
    RETURNING channel_id INTO v_ch_maadi;

    INSERT INTO chat_channels (org_id, location_id, name, type)
    VALUES (v_org, NULL, 'All Hands', 'broadcast')
    RETURNING channel_id INTO v_ch_allhands;

    -- --------------------------------------------------------
    -- MESSAGES (20 realistic messages across channels)
    -- --------------------------------------------------------

    -- El Gouna Team (5 messages)
    INSERT INTO chat_messages (org_id, channel_id, sender_id, sender_name, sender_role, body, pinned, created_at) VALUES
    (v_org, v_ch_elgouna, gen_random_uuid(), 'Ahmed K.', 'shift_manager',
     'Shift handoff: Walk-in cooler temp is at 3.2C, all good. Two prep containers of hummus left, should last through dinner. POS #2 had a paper jam earlier but its fixed now.',
     false, now() - interval '4 hours'),
    (v_org, v_ch_elgouna, gen_random_uuid(), 'Fatma A.', 'staff',
     'Heads up - we are completely 86d on the lamb shank. Supplier delivery pushed to tomorrow morning.',
     true, now() - interval '3 hours 30 minutes'),
    (v_org, v_ch_elgouna, gen_random_uuid(), 'Omar S.', 'staff',
     'Can someone cover my station for 10 min? Need to restock the dessert display from the back.',
     false, now() - interval '2 hours'),
    (v_org, v_ch_elgouna, gen_random_uuid(), 'Ahmed K.', 'shift_manager',
     'Great job on the rush tonight team. We hit 142 covers with zero complaints. Drinks are on me after close.',
     false, now() - interval '45 minutes'),
    (v_org, v_ch_elgouna, gen_random_uuid(), 'Nour M.', 'staff',
     'Closing checklist done. All stations wiped, trash out, walk-in locked. See everyone tomorrow!',
     false, now() - interval '15 minutes');

    -- New Cairo Team (5 messages)
    INSERT INTO chat_messages (org_id, channel_id, sender_id, sender_name, sender_role, body, pinned, created_at) VALUES
    (v_org, v_ch_newcairo, gen_random_uuid(), 'Layla H.', 'gm',
     'Reminder: health inspection is Tuesday. I need all temperature logs printed and ready by Monday EOD. No exceptions.',
     true, now() - interval '6 hours'),
    (v_org, v_ch_newcairo, gen_random_uuid(), 'Hassan R.', 'staff',
     'We are running low on takeaway containers (medium size). Down to maybe 30 left. Can we get an emergency order?',
     false, now() - interval '5 hours'),
    (v_org, v_ch_newcairo, gen_random_uuid(), 'Layla H.', 'gm',
     'Emergency order placed for containers. Should arrive by 4pm today. Use the large ones for now and apologize to guests for the switch.',
     false, now() - interval '4 hours 45 minutes'),
    (v_org, v_ch_newcairo, gen_random_uuid(), 'Youssef T.', 'shift_manager',
     'Night shift handoff: Cash drawer balanced at EGP 12,450. Two tables still seated (est. 30 min). Dishwasher making a weird noise again - maintenance ticket submitted.',
     false, now() - interval '2 hours'),
    (v_org, v_ch_newcairo, gen_random_uuid(), 'Sara B.', 'staff',
     'Just finished deep cleaning the grill station. Before and after photos uploaded to the task.',
     false, now() - interval '1 hour');

    -- Sheikh Zayed Team (4 messages)
    INSERT INTO chat_messages (org_id, channel_id, sender_id, sender_name, sender_role, body, pinned, created_at) VALUES
    (v_org, v_ch_zayed, gen_random_uuid(), 'Karim D.', 'shift_manager',
     'We have a party of 25 booked for 8pm tonight. I need all hands on deck. If you can pick up an extra hour, let me know ASAP.',
     true, now() - interval '7 hours'),
    (v_org, v_ch_zayed, gen_random_uuid(), 'Mona F.', 'staff',
     'I can stay an extra hour! Already clocked the extension in the app.',
     false, now() - interval '6 hours 30 minutes'),
    (v_org, v_ch_zayed, gen_random_uuid(), 'Ali G.', 'staff',
     'Same here, count me in. Should we prep extra rice and grilled proteins now or wait until 6?',
     false, now() - interval '6 hours'),
    (v_org, v_ch_zayed, gen_random_uuid(), 'Karim D.', 'shift_manager',
     'Start prepping now. Double the usual rice batch and pull extra chicken from the freezer. Thanks team!',
     false, now() - interval '5 hours 45 minutes');

    -- Maadi Team (3 messages)
    INSERT INTO chat_messages (org_id, channel_id, sender_id, sender_name, sender_role, body, pinned, created_at) VALUES
    (v_org, v_ch_maadi, gen_random_uuid(), 'Dina W.', 'gm',
     'New branded aprons are in the office. Everyone pick one up before your next shift. They look great!',
     false, now() - interval '8 hours'),
    (v_org, v_ch_maadi, gen_random_uuid(), 'Tarek N.', 'staff',
     'The ice machine on the floor is acting up again. Only producing half capacity. Put in a maintenance request.',
     false, now() - interval '3 hours'),
    (v_org, v_ch_maadi, gen_random_uuid(), 'Dina W.', 'gm',
     'Thanks Tarek. Maintenance team confirmed they will be here tomorrow between 9-11am. Use the backup ice from the kitchen freezer tonight.',
     false, now() - interval '2 hours 30 minutes');

    -- All Hands Broadcast (3 messages)
    INSERT INTO chat_messages (org_id, channel_id, sender_id, sender_name, sender_role, body, pinned, created_at) VALUES
    (v_org, v_ch_allhands, gen_random_uuid(), 'Rania E.', 'ops_director',
     'Company-wide update: We hit our Q1 revenue target across all locations! Bonuses will be reflected in next months payroll. Thank you all for the incredible effort.',
     true, now() - interval '24 hours'),
    (v_org, v_ch_allhands, gen_random_uuid(), 'Rania E.', 'ops_director',
     'Ramadan schedule changes go into effect next week. Please check the updated schedules in the app. Iftar meals provided daily at all locations.',
     false, now() - interval '12 hours'),
    (v_org, v_ch_allhands, gen_random_uuid(), 'Rania E.', 'ops_director',
     'Reminder: The new summer menu training sessions are mandatory. Check your location channel for specific dates and times.',
     false, now() - interval '4 hours');

END $$;
