-- +migrate Up
CREATE TABLE contacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    phone_hash  TEXT NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (owner_id, phone_hash)
);

CREATE INDEX contacts_owner_id_idx ON contacts(owner_id);
CREATE INDEX contacts_phone_hash_idx ON contacts(phone_hash);

-- +migrate Down
DROP TABLE contacts;
