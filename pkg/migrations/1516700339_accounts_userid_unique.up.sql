ALTER TABLE accounts
  DROP CONSTRAINT unique_github,
  DROP CONSTRAINT unique_facebook,
  DROP CONSTRAINT unique_google,
  ADD CONSTRAINT unique_user_id UNIQUE (user_id);