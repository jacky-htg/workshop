-- Enable extension yang diperlukan
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";  -- Untuk search name

-- Membuat tabel users
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    username    VARCHAR(100) NOT NULL,
    password    VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now()),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now()),
    deleted_at  TIMESTAMPTZ NULL
);

-- =====================================================
-- INDEX SET (Minimum Viable Index untuk awal project)
-- =====================================================

-- 1. Unique partial index untuk username (wajib, untuk login/auth)
CREATE UNIQUE INDEX idx_users_username_unique 
ON users(username) 
WHERE deleted_at IS NULL;

-- 2. Unique partial index untuk email (wajib, untuk komunikasi)
CREATE UNIQUE INDEX idx_users_email_unique 
ON users(email) 
WHERE deleted_at IS NULL;

-- 3. Index untuk pagination/sorting (sering diperlukan)
CREATE INDEX idx_users_created_at_active 
ON users(created_at DESC) 
WHERE deleted_at IS NULL;

-- 4. Partial index untuk filter is_active (kecil dan murah)
CREATE INDEX idx_users_is_active 
ON users(is_active) 
WHERE deleted_at IS NULL AND is_active = true;

-- 5. Trigram index untuk pencarian name (buat hanya jika fitur search diperlukan)
CREATE INDEX idx_users_name_trgm 
ON users USING gin(name gin_trgm_ops) 
WHERE deleted_at IS NULL;

-- =====================================================
-- TRIGGER untuk auto-update updated_at
-- =====================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = timezone('utc', now());
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- KOMENTAR TABEL & KOLOM
-- =====================================================

COMMENT ON TABLE users IS 'Table untuk menyimpan data user dengan soft delete';

COMMENT ON COLUMN users.id IS 'Primary key UUID v7, dibuat di Golang';
COMMENT ON COLUMN users.name IS 'Nama lengkap user';
COMMENT ON COLUMN users.username IS 'Username untuk login, harus unik (hanya untuk yang belum terhapus)';
COMMENT ON COLUMN users.password IS 'Password yang sudah di-hash (bcrypt/argon2)';
COMMENT ON COLUMN users.email IS 'Email user, harus unik (hanya untuk yang belum terhapus)';
COMMENT ON COLUMN users.is_active IS 'Status aktif user (true = aktif, false = non-aktif)';
COMMENT ON COLUMN users.created_at IS 'Waktu pembuatan record (UTC)';
COMMENT ON COLUMN users.updated_at IS 'Waktu terakhir update record (UTC), otomatis terupdate via trigger';
COMMENT ON COLUMN users.deleted_at IS 'Waktu soft delete (NULL = tidak terhapus, terisi = sudah dihapus)';

COMMENT ON INDEX idx_users_username_unique IS 'Menjamin username unik untuk data yang belum dihapus, juga mempercepat query login';
COMMENT ON INDEX idx_users_email_unique IS 'Menjamin email unik untuk data yang belum dihapus, juga mempercepat query by email';
COMMENT ON INDEX idx_users_created_at_active IS 'Mempercepat query dengan sorting created_at DESC untuk data aktif';
COMMENT ON INDEX idx_users_is_active IS 'Mempercepat query filter user aktif (partial index kecil)';
COMMENT ON INDEX idx_users_name_trgm IS 'Mempercepat pencarian name dengan partial match (LIKE) untuk data aktif';