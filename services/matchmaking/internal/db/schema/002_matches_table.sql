-- +goose Up

CREATE TABLE IF NOT EXISTS source.matches (
  id UUID primary key default uuid_generate_v4(),
  region TEXT,
  average_elo INT,
  created_at TIMESTAMP DEFAULT NOW()
);


-- +goose Down

DROP TABLE source.matches;