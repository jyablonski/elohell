# Elohell

A practice application simulating a matchmaking service for an online 5v5 game.

Components:

- Queue Producer Service (Python): Simulates users joining the matchmaking queue.
- Matchmaking Service (Go): Handles matchmaking logic to create balanced teams.
- Redis: Used as the queue for matchmaking requests
- Postgres: Stores persistent user and match data
