-- name: CreateMatch :one
INSERT INTO source.matches (region, average_elo)
VALUES ($1, $2)
RETURNING id, created_at;

-- name: InsertMatchPlayer :exec
INSERT INTO source.match_players (match_id, user_id, team, elo)
VALUES ($1, $2, $3, $4);