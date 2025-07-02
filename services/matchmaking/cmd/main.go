package main

import (
	"log"

	"github.com/jyablonski/elohell/services/matchmaking/internal/matchmaking"
)

func main() {
	mm := matchmaking.NewMatchmaker("redis:6379")
	log.Println("Starting matchmaking loop...")
	mm.BasicMatchmakingLoop()
}
