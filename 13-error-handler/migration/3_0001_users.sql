INSERT INTO users (id, name, username, password, email, is_active) VALUES
(uuid_generate_v4(), 'John Doe', 'johndoe', 'secret', 'john.doe@example.com', true),
(uuid_generate_v4(), 'Jane Smith', 'janesmith', 'secret', 'jane.smith@example.com', false);