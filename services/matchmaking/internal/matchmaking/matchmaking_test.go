package matchmaking

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// Helper function to get environment variables or default values
func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

// setupTestDB establishes a connection to the test PostgreSQL database
func setupTestDB(t *testing.T) *pgxpool.Pool {
	dbConn := getEnv("DB_CONN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable&search_path=source")
	pool, err := pgxpool.New(context.Background(), dbConn)
	require.NoError(t, err, "Failed to connect to PostgreSQL for tests")

	// Ping the database to ensure connection is live
	err = pool.Ping(context.Background())
	require.NoError(t, err, "Failed to ping PostgreSQL database")

	return pool
}

// setupTestRedis establishes a connection to the test Redis instance
func setupTestRedis(t *testing.T) *redis.Client {
	redisConn := getEnv("REDIS_CONN", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{
		Addr: redisConn,
	})

	// Ping Redis to ensure connection is live
	_, err := rdb.Ping(context.Background()).Result()
	require.NoError(t, err, "Failed to connect to Redis for tests")

	return rdb
}

// cleanupDB truncates tables and cleans up Redis for a fresh test run
func cleanupDB(t *testing.T, pool *pgxpool.Pool, rdb *redis.Client) {
	ctx := context.Background()

	// Clear Redis queue
	_, err := rdb.Del(ctx, "match_queue").Result()
	require.NoError(t, err, "Failed to clear Redis queue")

	// Truncate tables in PostgreSQL
	_, err = pool.Exec(ctx, "TRUNCATE TABLE source.match_players RESTART IDENTITY CASCADE;")
	require.NoError(t, err, "Failed to truncate match_players table")
	_, err = pool.Exec(ctx, "TRUNCATE TABLE source.matches RESTART IDENTITY CASCADE;")
	require.NoError(t, err, "Failed to truncate matches table")
}

func TestNewMatchmaker(t *testing.T) {
	// Use dummy values for testing the constructor
	redisAddr := getEnv("REDIS_CONN", "localhost:6379")
	dbURL := getEnv("DB_CONN", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable&search_path=source")

	mm, err := NewMatchmaker(redisAddr, dbURL)
	require.NoError(t, err)
	require.NotNil(t, mm)
	require.NotNil(t, mm.redisClient)
	require.NotNil(t, mm.db)
	require.NotNil(t, mm.pgxPool)

	// Close connections opened by NewMatchmaker
	defer mm.redisClient.Close()
	defer mm.pgxPool.Close()
}

func TestPopUser(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()
	pool := setupTestDB(t) // Need pool for NewMatchmaker, even if not used in PopUser directly
	defer pool.Close()

	mm, err := NewMatchmaker(rdb.Options().Addr, pool.Config().ConnString())
	require.NoError(t, err)
	defer mm.redisClient.Close()
	defer mm.pgxPool.Close()

	// Ensure queue is empty before test
	cleanupDB(t, pool, rdb)

	// Test popping from an empty queue
	user, err := mm.PopUser(mm.ctx) // Pass mm.ctx here
	require.NoError(t, err)
	require.Nil(t, user)

	// Push a user to the queue
	testUser := User{
		UserID:   "user123",
		Elo:      1500,
		Region:   "NA",
		QueuedAt: time.Now().Format(time.RFC3339),
	}
	userJSON, _ := json.Marshal(testUser)
	_, err = rdb.LPush(mm.ctx, mm.queueKey, string(userJSON)).Result()
	require.NoError(t, err)

	// Test popping a user
	poppedUser, err := mm.PopUser(mm.ctx) // Pass mm.ctx here
	require.NoError(t, err)
	require.NotNil(t, poppedUser)
	require.Equal(t, testUser.UserID, poppedUser.UserID)
	require.Equal(t, testUser.Elo, poppedUser.Elo)
	require.Equal(t, testUser.Region, poppedUser.Region)
}

func TestBasicMatchmakingLoop_Integration(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()
	pool := setupTestDB(t)
	defer pool.Close()

	mm, err := NewMatchmaker(rdb.Options().Addr, pool.Config().ConnString())
	require.NoError(t, err)
	defer mm.redisClient.Close()
	defer mm.pgxPool.Close()

	// Ensure a clean state before running the test
	cleanupDB(t, pool, rdb)
	defer cleanupDB(t, pool, rdb) // Clean up after the test as well

	// Push 10 users to the Redis queue
	usersToQueue := make([]User, 10)
	for i := 0; i < 10; i++ {
		usersToQueue[i] = User{
			UserID:   fmt.Sprintf("user%d", i+1),
			Elo:      1000 + i*10,
			Region:   "NA",
			QueuedAt: time.Now().Format(time.RFC3339),
		}
		userJSON, _ := json.Marshal(usersToQueue[i])
		_, err := rdb.LPush(mm.ctx, mm.queueKey, string(userJSON)).Result()
		require.NoError(t, err, "Failed to push user to Redis")
	}

	// Create a context with a timeout for the matchmaking loop and the test itself.
	// This context will be cancelled either by timeout or when the DB condition is met.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // Ensure context is cancelled when the test exits

	// Use a channel to signal when the goroutine running BasicMatchmakingLoop is done.
	done := make(chan struct{})

	// Run the matchmaking loop in a goroutine
	go func() {
		defer close(done)            // Signal that the goroutine is done when it returns
		mm.BasicMatchmakingLoop(ctx) // Pass the cancellable context to the loop
	}()

	// Polling loop to check for DB insertion
	var matchCount int
	var playerCount int
	pollInterval := 100 * time.Millisecond

	for {
		select {
		case <-ctx.Done(): // This case handles the overall test context timeout
			// The test context itself timed out.
			// This means the condition (matchCount == 1 && playerCount == 10) was NOT met in time.
			// Wait for the goroutine to confirm it's done before failing the test.
			select {
			case <-done:
				// Goroutine exited gracefully after context cancellation.
			case <-time.After(1 * time.Second): // Give it a moment to clean up
				t.Log("Warning: Matchmaking loop goroutine did not exit promptly after test context cancellation.")
			}
			t.Fatal("Test timed out waiting for match and players to be inserted into DB.")
		default:
			// Perform database queries
			row := pool.QueryRow(ctx, "SELECT COUNT(*) FROM source.matches;")
			err := row.Scan(&matchCount)
			// Do not use require.NoError here, as ctx might be cancelled by the time we check.
			// The t.Fatal in ctx.Done() case will handle the overall timeout.
			if err != nil && err != context.Canceled {
				t.Fatalf("Failed to query match count: %v", err)
			}

			row = pool.QueryRow(ctx, "SELECT COUNT(*) FROM source.match_players;")
			err = row.Scan(&playerCount)
			if err != nil && err != context.Canceled {
				t.Fatalf("Failed to query player count: %v", err)
			}

			if matchCount == 1 && playerCount == 10 {
				// Data found. Break out of the polling loop.
				// The defer cancel() will handle context cancellation when the test function returns.
				// Wait for the goroutine to finish its cleanup before proceeding.
				<-done
				goto EndPolling
			}
			time.Sleep(pollInterval)
		}
	}

EndPolling: // Label to jump to after successful polling
	// Assertions after data is confirmed to be in the DB
	// Use context.Background() for assertions to avoid "context canceled" errors,
	// as the main test context might have been cancelled already by the time we reach here.
	// This ensures assertions can always run against the DB if data was found.
	require.Equal(t, 1, matchCount, "Expected 1 match to be created")
	require.Equal(t, 10, playerCount, "Expected 10 players to be inserted")

	// Verify the match details
	var matchID string
	var region string
	var avgElo int32
	err = pool.QueryRow(context.Background(), "SELECT id, region, average_elo FROM source.matches LIMIT 1;").Scan(&matchID, &region, &avgElo)
	require.NoError(t, err, "Failed to retrieve match details")
	require.Equal(t, "NA", region, "Match region mismatch")

	// Calculate expected average ELO from the users pushed
	expectedTotalElo := 0
	for _, u := range usersToQueue {
		expectedTotalElo += u.Elo
	}
	expectedAvgElo := int32(expectedTotalElo / 10)
	require.Equal(t, expectedAvgElo, avgElo, "Match average ELO mismatch")

	// Verify player details (e.g., teams and ELOs)
	rows, err := pool.Query(context.Background(), "SELECT user_id, team, elo FROM source.match_players ORDER BY user_id;")
	require.NoError(t, err, "Failed to retrieve match players")
	defer rows.Close()

	insertedPlayers := make(map[string]struct {
		Team string
		Elo  int32
	})
	for rows.Next() {
		var userID, team string
		var elo int32
		err := rows.Scan(&userID, &team, &elo)
		require.NoError(t, err, "Failed to scan player row")
		insertedPlayers[userID] = struct {
			Team string
			Elo  int32
		}{Team: team, Elo: elo}
	}
	require.NoError(t, rows.Err(), "Error iterating player rows")

	require.Len(t, insertedPlayers, 10, "Expected 10 players in the database")

	// Basic check for team assignment and ELO
	for _, u := range usersToQueue {
		player, ok := insertedPlayers[u.UserID]
		require.True(t, ok, "Player %s not found in database", u.UserID)
		require.Contains(t, []string{"red", "blue"}, player.Team, "Player %s has invalid team %s", u.UserID, player.Team)
		require.Equal(t, int32(u.Elo), player.Elo, "Player %s ELO mismatch", u.UserID)
	}
}
