ALTER TABLE links
  ADD CONSTRAINT type_user_id UNIQUE (type, user_id);