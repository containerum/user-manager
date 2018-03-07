ALTER TABLE accounts
  ADD CONSTRAINT unique_github UNIQUE (user_id, github),
  ADD CONSTRAINT unique_facebook UNIQUE (user_id, facebook),
  ADD CONSTRAINT unique_google UNIQUE (user_id, google),
  DROP CONSTRAINT unique_user_id;