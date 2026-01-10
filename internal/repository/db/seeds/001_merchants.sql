-- Seed 2 merchants
INSERT INTO merchants (id, mechant_name) VALUES
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', 'Acme Corporation'),
('b2c3d4e5-f6a7-8901-bcde-f12345678901', 'Global Retail Inc')
ON CONFLICT (id) DO NOTHING;
