CREATE SCHEMA source;
SET search_path TO source;

-- this has to come after setting the schema search path ;-)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP TABLE IF EXISTS matches;
CREATE TABLE IF NOT EXISTS matches (
  id UUID primary key default uuid_generate_v4(),
  region TEXT,
  average_elo INT,
  created_at TIMESTAMP DEFAULT NOW()
);

DROP TABLE IF EXISTS match_players;
CREATE TABLE IF NOT EXISTS match_players (
  match_id UUID REFERENCES source.matches(id) ON DELETE CASCADE,
  user_id TEXT,
  team TEXT CHECK (team IN ('red', 'blue')),
  elo INT,
  PRIMARY KEY (match_id, user_id)
);
