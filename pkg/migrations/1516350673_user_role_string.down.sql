UPDATE users set "role" = '0' WHERE "role" = 'user';
UPDATE users set "role" = '1' WHERE "role" = 'admin';
UPDATE users set "role" = '-1' WHERE "role" NOT IN ('user', 'admin');
ALTER TABLE users
  ALTER COLUMN "role" SET DATA TYPE INTEGER USING trim("role")::INTEGER;
