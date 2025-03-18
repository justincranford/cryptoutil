CREATE TABLE kek_pool (
    kek_pool_id                    TEXT         NOT NULL CHECK(length(kek_pool_id)          = 36), -- UUIDv7
    kek_pool_name                  VARCHAR(63)  NOT NULL CHECK(length(kek_pool_name)        >= 1),
    kek_pool_description           VARCHAR(255) NOT NULL CHECK(length(kek_pool_description) >= 1),
    kek_pool_algorithm             VARCHAR(15)  NOT NULL CHECK(kek_pool_algorithm             IN ('AES-256', 'AES-192', 'AES-128')),
    kek_pool_status                VARCHAR(14)  NOT NULL CHECK(kek_pool_status                IN ('active', 'disabled', 'pending_generate', 'pending_import')),
    kek_pool_provider              VARCHAR(8)   NOT NULL CHECK(kek_pool_provider              IN ('Internal')),
    kek_pool_is_versioning_allowed BOOLEAN      NOT NULL CHECK(kek_pool_is_versioning_allowed IN (0, 1)),
    kek_pool_is_import_allowed     BOOLEAN      NOT NULL CHECK(kek_pool_is_import_allowed     IN (0, 1)),
    kek_pool_is_export_allowed     BOOLEAN      NOT NULL CHECK(kek_pool_is_export_allowed     IN (0, 1)),
    CHECK(length(kek_pool_name)        <= 63),
    CHECK(length(kek_pool_description) <= 255)
);

CREATE UNIQUE INDEX idx_kek_pool_name_kek_pool_id ON kek_pool(kek_pool_name);

CREATE TABLE kek (
    kek_pool_id         TEXT         NOT NULL CHECK(length(kek_pool_id) = 36), -- UUIDv7
    kek_id              INTEGER      NOT NULL CHECK(kek_id >= 0),
    kek_material        BLOB         NOT NULL CHECK(length(kek_material) >= 1),
    kek_generate_date   CHAR(20)         NULL, -- ISO 8601
    kek_import_date     CHAR(20)         NULL, -- ISO 8601
    kek_expiration_date CHAR(20)         NULL, -- ISO 8601
    kek_revocation_date CHAR(20)         NULL, -- ISO 8601
    PRIMARY KEY (kek_pool_id, kek_id), -- Composite primary key
    FOREIGN KEY (kek_pool_id) REFERENCES kek_pool(kek_pool_id) ON DELETE CASCADE,
    CHECK(length(kek_material) <= 512),
    -- CHECK(kek_generate_date   CHECK(length(kek_generate_date)     == 20), -- ISO 8601
    -- CHECK(kek_import_date     CHECK(length(kek_import_date)       == 20), -- ISO 8601
    -- CHECK(kek_expiration_date CHECK(length(kek_expiration_date)   == 20), -- ISO 8601
    -- CHECK(kek_revocation_date CHECK(length(kek_revocation_date)   == 20), -- ISO 8601
    -- CHECK(kek_generate_date   LIKE '____-__-__T__:__:__Z'), -- ISO 8601
    -- CHECK(kek_import_date     LIKE '____-__-__T__:__:__Z'), -- ISO 8601
    -- CHECK(kek_expiration_date LIKE '____-__-__T__:__:__Z'), -- ISO 8601
    -- CHECK(kek_revocation_date LIKE '____-__-__T__:__:__Z'), -- ISO 8601
    -- CHECK(kek_generate_date   GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z'), -- ISO 8601
    -- CHECK(kek_import_date     GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z'), -- ISO 8601
    -- CHECK(kek_expiration_date GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z'), -- ISO 8601
    -- CHECK(kek_revocation_date GLOB '[0-9][0-9][0-9][0-9]-[0-9][0-9]-[0-9][0-9]T[0-9][0-9]:[0-9][0-9]:[0-9][0-9]Z'), -- ISO 8601
    CHECK (
        (kek_generate_date IS NOT NULL AND kek_import_date IS     NULL) OR
        (kek_generate_date IS     NULL AND kek_import_date IS NOT NULL)
    ), -- Exactly one of the two must be NOT NULL
    CHECK(kek_expiration_date IS NULL OR kek_expiration_date > COALESCE(kek_generate_date, kek_import_date)),
    CHECK(kek_revocation_date IS NULL OR kek_revocation_date > COALESCE(kek_generate_date, kek_import_date))
);
