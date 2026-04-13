-- +migrate Up
CREATE TABLE posts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    author_id   UUID NOT NULL REFERENCES users(id),
    image_url   TEXT NOT NULL,
    caption     TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX posts_group_id_idx ON posts(group_id);
CREATE INDEX posts_author_id_idx ON posts(author_id);
CREATE INDEX posts_created_at_idx ON posts(created_at DESC);

-- +migrate Down
DROP TABLE posts;
