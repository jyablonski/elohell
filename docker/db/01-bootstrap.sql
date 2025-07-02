CREATE SCHEMA source;
SET search_path TO source;

-- this has to come after setting the schema search path ;-)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

DROP TABLE IF EXISTS matches;
CREATE TABLE IF NOT EXISTS matches (
  id UUID primary key default uuid_generate_v4(),
  red_team TEXT[],   -- array of player IDs
  blue_team TEXT[],
  region TEXT,
  average_skill INT,
  created_at TIMESTAMP DEFAULT NOW()
);