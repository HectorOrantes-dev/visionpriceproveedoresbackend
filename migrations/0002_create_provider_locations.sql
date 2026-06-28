-- Migración: crea la tabla provider_locations para la ubicación (dirección)
-- del almacén / sucursal de cada proveedor.
-- Ejecutar en el editor SQL de Supabase (o vía psql) ANTES de usar /geo.
--
-- - id: UUID, llave primaria, generado por la base de datos.
-- - provider_id: UUID, ÚNICO (un proveedor = una ubicación). El UNIQUE es
--   obligatorio porque el backend hace UPSERT con ON CONFLICT (provider_id);
--   sin él, PUT /api/v1/providers/location devuelve 500.
-- - address: texto de la dirección.
-- - updated_at: marca de tiempo de la última actualización.

CREATE TABLE IF NOT EXISTS provider_locations (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id UUID        NOT NULL UNIQUE REFERENCES providers (id) ON DELETE CASCADE,
    address     TEXT        NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Por si la tabla ya existía sin la restricción UNIQUE en provider_id
-- (necesaria para el ON CONFLICT del UPSERT).
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'provider_locations_provider_id_key'
    ) THEN
        ALTER TABLE provider_locations
            ADD CONSTRAINT provider_locations_provider_id_key UNIQUE (provider_id);
    END IF;
END$$;
