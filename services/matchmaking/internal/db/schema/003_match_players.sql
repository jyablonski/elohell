-- +goose Up

CREATE TABLE IF NOT EXISTS source.match_players (
  match_id UUID REFERENCES source.matches(id) ON DELETE CASCADE,
  user_id TEXT,
  team TEXT CHECK (team IN ('red', 'blue')),
  elo INT,
  PRIMARY KEY (match_id, user_id)
);


-- +goose Down

DROP TABLE source.match_players;