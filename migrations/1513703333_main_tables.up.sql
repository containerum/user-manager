CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TABLE users
(
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
    login TEXT,
    password_hash TEXT,
    salt TEXT,
    role INTEGER,
    is_active BOOLEAN,
    is_deleted BOOLEAN,
    is_in_blacklist BOOLEAN
);
CREATE TABLE accounts
(
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
    user_id UUID,
    github TEXT,
    facebook TEXT,
    google TEXT,
    CONSTRAINT accounts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE TABLE links
(
    link TEXT PRIMARY KEY NOT NULL,
    user_id UUID,
    type TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    expired_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN,
    sent_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT links_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE TABLE profiles
(
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY NOT NULL,
    user_id UUID,
    referral TEXT,
    access TEXT,
    data TEXT,
    created_at TIMESTAMP WITH TIME ZONE,
    blacklist_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT profiles_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id)
);
CREATE TABLE tokens
(
    token TEXT PRIMARY KEY NOT NULL,
    user_id UUID,
    created_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN,
    session_id UUID,
    CONSTRAINT tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id)
);
