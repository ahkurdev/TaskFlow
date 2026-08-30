CREATE TABLE IF NOT EXISTS lists (
    id         BIGSERIAL PRIMARY KEY,
    board_id   BIGINT NOT NULL REFERENCES boards(id) ON DELETE CASCADE,
    name       TEXT NOT NULL,
    position   INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_lists_board_id ON lists (board_id);
