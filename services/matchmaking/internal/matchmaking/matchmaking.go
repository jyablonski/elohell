package matchmaking

import (
	"context"
	"encoding/json"
	"log"
	"time"

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
	ctx         context.Context
}

func NewMatchmaker(redisAddr string) *Matchmaker {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	return &Matchmaker{
		redisClient: rdb,
		queueKey:    "match_queue",
		ctx:         context.Background(),
	}
}

// PopUser pops one user from the queue
func (m *Matchmaker) PopUser() (*User, error) {
	res, err := m.redisClient.RPop(m.ctx, m.queueKey).Result()
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
func (m *Matchmaker) BasicMatchmakingLoop() {
	const teamSize = 5
	var waitingUsers []*User

	for {
		user, err := m.PopUser()
		if err != nil {
			log.Printf("Error popping user: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if user == nil {
			// No users in queue, sleep and retry
			time.Sleep(500 * time.Millisecond)
			continue
		}

		waitingUsers = append(waitingUsers, user)
		log.Printf("User %s queued, waiting users: %d", user.UserID, len(waitingUsers))

		if len(waitingUsers) >= teamSize {
			match := waitingUsers[:teamSize]
			waitingUsers = waitingUsers[teamSize:]
			log.Printf("Match formed: %v", match)
			// TODO: Save match to DB, notify players, etc.
		}
	}
}
