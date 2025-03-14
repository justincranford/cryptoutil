CREATE TABLE kek (
                     id TEXT PRIMARY KEY,
                     name TEXT NOT NULL CHECK(LENGTH(name) BETWEEN 3 AND 50),
                     description TEXT,
                     algorithm TEXT NOT NULL CHECK(LENGTH(algorithm) BETWEEN 5 AND 20),
                     create_date TEXT NOT NULL CHECK(create_date GLOB '[0-9]*T[0-9]*Z'),
    status TEXT CHECK(status IN ('pending_import', 'active', 'disabled', 'expired')) NOT NULL,
    enable_versioning BOOLEAN NOT NULL,
    imported BOOLEAN NOT NULL,
    provider TEXT CHECK(provider IN ('AWS', 'GCP', 'Azure', 'Internal')) NOT NULL,
    exportable BOOLEAN NOT NULL,
    key_material TEXT
);

CREATE UNIQUE INDEX idx_kek_name_id ON kek(name);

CREATE TABLE kek_version (
                             version TEXT PRIMARY KEY CHECK(LENGTH(version) BETWEEN 5 AND 100),
                             kek_id TEXT NOT NULL,
                             create_date TEXT NOT NULL CHECK(create_date GLOB '[0-9]*T[0-9]*Z'),
    update_date TEXT CHECK(update_date GLOB '[0-9]*T[0-9]*Z'),
    delete_date TEXT CHECK(delete_date GLOB '[0-9]*T[0-9]*Z'),
    import_date TEXT CHECK(import_date GLOB '[0-9]*T[0-9]*Z'),
    FOREIGN KEY (kek_id) REFERENCES kek(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_kek_version_version_kek_id ON kek_version(kek_id, version);
