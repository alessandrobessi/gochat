package types

import "net"

// PublicMessage is struct representing a public message
type PublicMessage struct {
	Sender string
	Body   string
}

// PrivateMessage is struct representing a private message
type PrivateMessage struct {
	Sender    string
	Body      string
	Recipient string
}

// Client is struct representing a client connected to the server
type Client struct {
	ID        string
	Name      string
	Conn      net.Conn
	IsActive  bool
	IsNameSet bool // indicates whether a name have been set by the user or the default (client id) is used
}
