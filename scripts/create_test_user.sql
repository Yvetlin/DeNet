INSERT INTO users (id, username, email, balance) 
VALUES 
    ('550e8400-e29b-41d4-a716-446655440000', 'testuser1', 'test1@example.com', 0),
    ('550e8400-e29b-41d4-a716-446655440001', 'testuser2', 'test2@example.com', 0)
ON CONFLICT (id) DO NOTHING;

