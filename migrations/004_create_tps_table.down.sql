-- Drop TPS (Third Party Software) table
DROP TRIGGER IF EXISTS update_tps_updated_at ON tps;
DROP TABLE IF EXISTS tps;
