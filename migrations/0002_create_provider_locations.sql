-- Migración: tabla provider_locations (ubicación / dirección del almacén del proveedor).
-- Ejecutar en el editor SQL de Supabase (o vía psql) ANTES de usar /geo.
--
-- El backend (UpsertLocation) usa estas columnas: id, provider_id, address, updated_at,
-- y hace UPSERT con ON CONFLICT (provider_id), por lo que provider_id debe ser ÚNICO.
--
-- Nota: en algunos entornos la tabla ya existía con un esquema previo que tenía
-- lat / lng / delivery_radius_km (NOT NULL) y SIN columna address. Eso provocaba
-- un 500 en PUT /api/v1/providers/location. Esta migración es idempotente y
-- reconcilia ambos casos: crea la tabla si falta, agrega address si falta y
-- relaja los NOT NULL de las columnas que el backend no escribe.

-- 1) Crear la tabla si no existe (instalación limpia).
CREATE TABLE IF NOT EXISTS provider_locations (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID        NOT NULL REFERENCES providers (id) ON DELETE CASCADE,
    address     TEXT,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2) Asegurar la columna address (si la tabla ya existía sin ella).
ALTER TABLE provider_locations ADD COLUMN IF NOT EXISTS address TEXT;

-- 3) provider_id debe ser ÚNICO para el ON CONFLICT del UPSERT.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conrelid = 'public.provider_locations'::regclass AND contype = 'u'
    ) THEN
        ALTER TABLE provider_locations
            ADD CONSTRAINT provider_locations_provider_id_key UNIQUE (provider_id);
    END IF;
END$$;

-- 4) Relajar NOT NULL en columnas que el backend NO escribe (esquema previo),
--    para que el INSERT (provider_id, address, updated_at) no falle.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'provider_locations' AND column_name = 'lat') THEN
        ALTER TABLE provider_locations ALTER COLUMN lat DROP NOT NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'provider_locations' AND column_name = 'lng') THEN
        ALTER TABLE provider_locations ALTER COLUMN lng DROP NOT NULL;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'provider_locations' AND column_name = 'delivery_radius_km') THEN
        ALTER TABLE provider_locations ALTER COLUMN delivery_radius_km DROP NOT NULL;
    END IF;
END$$;
