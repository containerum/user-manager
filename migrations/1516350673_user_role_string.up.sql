ALTER TABLE users
  ALTER COLUMN "role" SET DATA TYPE TEXT USING "role"::TEXT;
UPDATE users set "role" = 'user' WHERE "role" = '0';
UPDATE users set "role" = 'admin' WHERE "role" = '1';
