ALTER TABLE profiles
  ALTER COLUMN "data" DROP DEFAULT,
  ALTER COLUMN "data" SET DATA TYPE jsonb USING "data"::jsonb,
  ALTER COLUMN "data" SET DEFAULT '{}'::jsonb;