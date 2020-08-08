# gochat

A simple multi user client/server chat application written in Go using less than 250 loc.
Useful to learn how to deal with concurrency in Go (goroutines, channels, and mutexes). 

### Usage
Launch the server:
`go run src/cmd/server/server.go`

Launch one or more clients:
`go run src/cmd/client/client.go`

### Commands
- `!quit` to leave the chat
- `!name [your-name]` to change your name
- `!dm [user] [message]` to send a private message to a user
