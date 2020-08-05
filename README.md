<div align="center">

# gochat

A simple multi user client/server chat application written in Go.
Useful to learn how to deal with concurrency in Go (goroutines, channels, and mutexes). 
</div>

### Usage
Launch the server:
`go run server.go`

Launch one or more clients:
`go run client.go`

### Commands
- `!quit` to leave the chat
- `!name [your-name]` to change your name
- `!dm [user] [message]` to send a private message to a user
