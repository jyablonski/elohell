package matchmaking

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype" // Import pgtype
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jyablonski/elohell/services/matchmaking/internal/db"

	"github.com/redis/go-redis/v9"
)

type User struct {
	UserID   string `json:"user_id"`
	Elo      int    `json:"elo"`
	Region   string `json:"region"`
	QueuedAt string `json:"queued_at"`
}

type Matchmaker struct {
	redisClient *redis.Client
	queueKey    string
	ctx         context.Context // This context is for the Matchmaker's lifetime operations
	db          *db.Queries
	pgxPool     *pgxpool.Pool // for transactions
}

func NewMatchmaker(redisAddr, dbURL string) (*Matchmaker, error) {
	ctx := context.Background()
	pgxPool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &Matchmaker{
		redisClient: rdb,
		queueKey:    "match_queue",
		ctx:         ctx, // Initialize with a background context
		db:          db.New(pgxPool),
		pgxPool:     pgxPool,
	}, nil
}

// PopUser pops one user from the queue
func (m *Matchmaker) PopUser(ctx context.Context) (*User, error) { // Added ctx parameter
	res, err := m.redisClient.RPop(ctx, m.queueKey).Result() // Use the passed ctx
	if err == redis.Nil {
		return nil, nil // empty queue
	}
	if err != nil {
		return nil, err
	}

	var user User
	if err := json.Unmarshal([]byte(res), &user); err != nil {
		return nil, err
	}
	return &user, nil
}

// BasicMatchmakingLoop shows a simple loop popping users and forming matches
// It now accepts a context to allow for graceful shutdown.
func (m *Matchmaker) BasicMatchmakingLoop(ctx context.Context) { // Added ctx parameter
	const teamSize = 5
	var waitingUsers []*User

	for {
		select {
		case <-ctx.Done():
			// Context was cancelled (e.g., by test timeout), exit the loop
			slog.Info("Matchmaking loop stopped due to context cancellation.")
			return
		default:
			// Continue with the loop
		}

		user, err := m.PopUser(ctx) // Pass the context to PopUser
		if err != nil {
			slog.Error("Error popping user", "error", err)
			time.Sleep(time.Second)
			continue
		}
		if user == nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		waitingUsers = append(waitingUsers, user)
		slog.Info("User queued", "user_id", user.UserID, "waiting_count", len(waitingUsers))

		if len(waitingUsers) >= 10 {
			redTeam := waitingUsers[:5]
			blueTeam := waitingUsers[5:10]
			waitingUsers = waitingUsers[10:]

			totalElo := 0
			for i := 0; i < 5; i++ {
				totalElo += redTeam[i].Elo + blueTeam[i].Elo
			}
			avgElo := int32(totalElo / 10)
			region := redTeam[0].Region

			tx, err := m.pgxPool.Begin(ctx) // Use the passed ctx for transaction
			if err != nil {
				slog.Error("Failed to start transaction", "error", err)
				continue
			}

			// Create a new Queries instance bound to the transaction
			q := db.New(tx)

			// Insert match
			matchParams := db.CreateMatchParams{
				Region: pgtype.Text{
					String: region,
					Valid:  true,
				},
				AverageElo: pgtype.Int4{
					Int32: avgElo,
					Valid: true,
				},
			}
			match, err := q.CreateMatch(ctx, matchParams) // Pass the context
			if err != nil {
				slog.Error("Failed to create match", "error", err)
				tx.Rollback(ctx) // Rollback on error
				continue
			}

			// Convert pgtype.UUID to uuid.UUID
			matchID, err := uuid.FromBytes(match.ID.Bytes[:])
			if err != nil {
				slog.Error("Failed to convert match ID to uuid.UUID", "error", err)
				tx.Rollback(ctx) // Rollback on error
				continue
			}

			insertPlayer := func(matchID uuid.UUID, userID string, team string, elo int32) error {
				params := db.InsertMatchPlayerParams{
					MatchID: pgtype.UUID{
						Bytes: matchID,
						Valid: true,
					},
					UserID: userID,
					Team: pgtype.Text{
						String: team,
						Valid:  true,
					},
					Elo: pgtype.Int4{
						Int32: elo,
						Valid: true,
					},
				}
				return q.InsertMatchPlayer(ctx, params) // Pass the context
			}

			// Insert all players
			allPlayersInserted := true
			for i := 0; i < 5; i++ {
				if err := insertPlayer(matchID, redTeam[i].UserID, "red", int32(redTeam[i].Elo)); err != nil {
					slog.Error("Failed to insert red player", "error", err)
					allPlayersInserted = false
					break // Exit loop on first error
				}
				if err := insertPlayer(matchID, blueTeam[i].UserID, "blue", int32(blueTeam[i].Elo)); err != nil {
					slog.Error("Failed to insert blue player", "error", err)
					allPlayersInserted = false
					break // Exit loop on first error
				}
			}

			if !allPlayersInserted {
				tx.Rollback(ctx) // Rollback if any player insertion failed
				continue
			}

			// Commit the transaction after all operations are successful
			if err := tx.Commit(ctx); err != nil {
				slog.Error("Failed to commit transaction", "error", err)
				continue
			}

			slog.Info("Match and players saved", "match_id", match.ID, "region", region, "average_elo", avgElo)
		}
	}
}
