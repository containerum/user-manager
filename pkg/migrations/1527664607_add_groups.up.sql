CREATE TABLE IF NOT EXISTS groups
(
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
  label TEXT NOT NULL,
  owner_user_id UUID NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL,
  CONSTRAINT group_owner_user_id FOREIGN KEY (owner_user_id) REFERENCES users (id)
);
CREATE TABLE IF NOT EXISTS groups_members
(
  id UUID DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
  group_id UUID  NOT NULL,
  user_id UUID NOT NULL,
  default_access TEXT NOT NULL,
  added_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW() NOT NULL,
  CONSTRAINT group_member_group_id FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
  CONSTRAINT group_member_user_id FOREIGN KEY (user_id) REFERENCES users (id),
  CONSTRAINT unique_user_id_group UNIQUE (group_id, user_id)
);