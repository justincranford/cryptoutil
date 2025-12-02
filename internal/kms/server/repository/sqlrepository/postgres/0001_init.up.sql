-- PostgreSQL initialization
-- This migration creates the basic database schema for cryptoutil

-- Create barrier_content_keys table
CREATE TABLE public.barrier_content_keys (
    uuid uuid NOT NULL PRIMARY KEY,
    encrypted text NOT NULL,
    kek_uuid uuid NOT NULL
);

-- Create barrier_intermediate_keys table
CREATE TABLE public.barrier_intermediate_keys (
    uuid uuid NOT NULL PRIMARY KEY,
    encrypted text NOT NULL,
    kek_uuid uuid NOT NULL
);

-- Create barrier_root_keys table
CREATE TABLE public.barrier_root_keys (
    uuid uuid NOT NULL PRIMARY KEY,
    encrypted text NOT NULL,
    kek_uuid uuid NOT NULL
);

-- Create elastic_keys table
CREATE TABLE public.elastic_keys (
    elastic_key_id uuid NOT NULL PRIMARY KEY,
    elastic_key_name character varying(63) NOT NULL,
    elastic_key_description character varying(255) NOT NULL,
    elastic_key_provider character varying(8) NOT NULL,
    elastic_key_algorithm character varying(26) NOT NULL,
    elastic_key_versioning_allowed boolean NOT NULL,
    elastic_key_import_allowed boolean NOT NULL,
    elastic_key_status character varying(34) NOT NULL,
    CONSTRAINT chk_elastic_keys_elastic_key_description CHECK (length(elastic_key_description) >= 1),
    CONSTRAINT chk_elastic_keys_elastic_key_import_allowed CHECK (elastic_key_import_allowed IN (true, false)),
    CONSTRAINT chk_elastic_keys_elastic_key_name CHECK (length(elastic_key_name) >= 1),
    CONSTRAINT chk_elastic_keys_elastic_key_provider CHECK (elastic_key_provider = 'Internal'),
    CONSTRAINT chk_elastic_keys_elastic_key_status CHECK (elastic_key_status IN ('creating', 'import_failed', 'pending_import', 'pending_generate', 'generate_failed', 'active', 'disabled', 'pending_delete_was_import_failed', 'pending_delete_was_pending_import', 'pending_delete_was_active', 'pending_delete_was_disabled', 'pending_delete_was_generate_failed', 'started_delete', 'finished_delete')),
    CONSTRAINT chk_elastic_keys_elastic_key_versioning_allowed CHECK (elastic_key_versioning_allowed IN (true, false)),
    CONSTRAINT uni_elastic_keys_elastic_key_name UNIQUE (elastic_key_name)
);

-- Create material_keys table
CREATE TABLE public.material_keys (
    elastic_key_id uuid NOT NULL,
    material_key_id uuid NOT NULL,
    material_key_clear_public bytea,
    material_key_encrypted_non_public bytea NOT NULL,
    material_key_generate_date timestamp with time zone,
    material_key_import_date timestamp with time zone,
    material_key_expiration_date timestamp with time zone,
    material_key_revocation_date timestamp with time zone,
    PRIMARY KEY (elastic_key_id, material_key_id),
    CONSTRAINT fk_material_keys_elastic_key_id FOREIGN KEY (elastic_key_id) REFERENCES public.elastic_keys(elastic_key_id)
);
