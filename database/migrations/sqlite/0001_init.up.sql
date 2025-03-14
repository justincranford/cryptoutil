CREATE TABLE kek (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK(LENGTH(name) BETWEEN 3 AND 50),
    description TEXT CHECK(LENGTH(name) BETWEEN 1 AND 255),
    algorithm TEXT NOT NULL CHECK(LENGTH(algorithm) BETWEEN 5 AND 20),
    status TEXT CHECK(status IN ('active', 'disabled', 'pending_import', 'expired')) NOT NULL,
    provider TEXT CHECK(provider IN ('Internal', 'AWS', 'GCP', 'Azure')) NOT NULL,
    is_versioning_enabled BOOLEAN NOT NULL,
    is_imported BOOLEAN NOT NULL,
    is_exportable BOOLEAN NOT NULL,
    create_date TEXT NOT NULL CHECK(create_date GLOB '[0-9]*T[0-9]*Z'),
);

CREATE UNIQUE INDEX idx_kek_name_id ON kek(name);

CREATE TABLE kek_version (
    kek_id TEXT NOT NULL,
    version TEXT PRIMARY KEY CHECK(LENGTH(version) BETWEEN 5 AND 100),
    key_material TEXT
    update_date TEXT CHECK(update_date GLOB '[0-9]*T[0-9]*Z'),
    delete_date TEXT CHECK(delete_date GLOB '[0-9]*T[0-9]*Z'),
    import_date TEXT CHECK(import_date GLOB '[0-9]*T[0-9]*Z'),
    create_date TEXT NOT NULL CHECK(create_date GLOB '[0-9]*T[0-9]*Z'),
    FOREIGN KEY (kek_id) REFERENCES kek(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_kek_version_version_kek_id ON kek_version(kek_id, version);
