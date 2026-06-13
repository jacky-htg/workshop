-- =====================================================
-- TABEL access
-- =====================================================
CREATE TABLE IF NOT EXISTS access (
    id          SERIAL PRIMARY KEY,
    parent_id   INTEGER,
    name        VARCHAR(255) NOT NULL UNIQUE,
    alias       VARCHAR(255) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now())
);

-- Index untuk parent_id (hierarchical queries)
CREATE INDEX idx_access_parent_id ON access(parent_id);

-- =====================================================
-- TABEL roles
-- =====================================================
CREATE TABLE IF NOT EXISTS roles (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL UNIQUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now())
);

-- =====================================================
-- TABEL access_roles (many-to-many)
-- =====================================================
CREATE TABLE IF NOT EXISTS access_roles (
    id          SERIAL PRIMARY KEY,
    access_id   INTEGER NOT NULL,
    role_id     INTEGER NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now()),
    
    -- Constraint unique composite
    CONSTRAINT access_roles_unique UNIQUE (access_id, role_id),
    
    -- Foreign keys
    CONSTRAINT fk_access_roles_to_access 
        FOREIGN KEY (access_id) 
        REFERENCES access(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE,
    
    CONSTRAINT fk_access_roles_to_roles 
        FOREIGN KEY (role_id) 
        REFERENCES roles(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE
);

-- Index untuk performance
CREATE INDEX idx_access_roles_access_id ON access_roles(access_id);
CREATE INDEX idx_access_roles_role_id ON access_roles(role_id);

-- =====================================================
-- TABEL roles_users (many-to-many)
-- =====================================================
CREATE TABLE IF NOT EXISTS roles_users (
    id          BIGSERIAL PRIMARY KEY,
    role_id     INTEGER NOT NULL,
    user_id     UUID NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now()),
    
    -- Constraint unique composite
    CONSTRAINT roles_users_unique UNIQUE (role_id, user_id),
    
    -- Foreign keys
    CONSTRAINT fk_roles_users_to_roles 
        FOREIGN KEY (role_id) 
        REFERENCES roles(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE,
    
    CONSTRAINT fk_roles_users_to_users 
        FOREIGN KEY (user_id) 
        REFERENCES users(id) 
        ON DELETE CASCADE 
        ON UPDATE CASCADE
);

-- Index untuk performance
CREATE INDEX idx_roles_users_role_id ON roles_users(role_id);
CREATE INDEX idx_roles_users_user_id ON roles_users(user_id);

-- =====================================================
-- INDEX TAMBAHAN UNTUK PERFORMANCE (Opsional)
-- =====================================================

-- Index composite untuk query permission checking yang umum
CREATE INDEX idx_access_roles_composite_lookup 
    ON access_roles(role_id, access_id);

-- Index untuk roles_users jika sering join dengan users
CREATE INDEX idx_roles_users_user_role 
    ON roles_users(user_id, role_id);
