INSERT INTO users (id, name, username, password, email, is_active) VALUES
('019eb960-a27d-73c8-9703-b23a9f50dc83', 'Admin', 'admin', '$2a$10$D7UJmo0/bnXUyvsvRNKmc.cLeiLPNGQ8TfBnQHc2hkQV.oSFBh.qO', 'admin@example.com', true);

INSERT INTO access (name, alias) VALUES ('root', 'root');

INSERT INTO roles (name) VALUES ('superadmin');

INSERT INTO access_roles (role_id, access_id)  VALUES (1, 1);

INSERT INTO roles_users (role_id, user_id) VALUES (1, '019eb960-a27d-73c8-9703-b23a9f50dc83');

