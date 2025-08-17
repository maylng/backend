ALTER TABLE accounts
ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;

-- Optionally, set current specific account(s) to admin (run manually)
-- UPDATE accounts SET is_admin = TRUE WHERE id = 'fa657c1c-01ec-4541-9fa3-ef4283b7ddaf';
