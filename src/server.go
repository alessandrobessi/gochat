package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
)

type PublicMessage struct {
	sender string
	body   string
}

type PrivateMessage struct {
	sender    string
	body      string
	recipient string
}

type Client struct {
	id        string
	name      string
	conn      net.Conn
	isActive  bool
	isNameSet bool // indicates whether a name have been set by the user or the default (client id) is used
}

type ActiveClients struct {
	m   map[string]Client
	mux sync.Mutex // to avoid race conditions
}

func (c *ActiveClients) HasKey(key string) bool {
	c.mux.Lock()
	if _, ok := c.m[key]; ok {
		c.mux.Unlock()
		return true
	}
	c.mux.Unlock()
	return false
}

func (c *ActiveClients) Map() *map[string]Client {
	return &c.m
}

func (c *ActiveClients) Count() int {
	numActiveClients := 0
	for _, client := range c.m {
		if client.isActive == true {
			numActiveClients++
		}
	}
	return numActiveClients
}

func (c *ActiveClients) CleanUp() {
	c.mux.Lock()
	for _, client := range c.m {
		if client.isActive == false {
			delete(c.m, client.name)
		}
	}
	c.mux.Unlock()
}

func (c *ActiveClients) AddClient(key string, value Client) {
	c.mux.Lock()
	c.m[key] = value
	c.mux.Unlock()
}

func (c *ActiveClients) DeleteClient(key string) {
	c.mux.Lock()
	delete(c.m, key)
	c.mux.Unlock()
}

var activeClients ActiveClients // global struct containing information about active clients

// reads from the privateMessageChan and send private messages to users
func sendPrivateMessages(privateMessagesChan chan PrivateMessage) {
	for {
		select {
		case msg := <-privateMessagesChan:
			sent := false
			for name, client := range *activeClients.Map() {
				if name == msg.recipient && client.isActive == true {
					_, err := fmt.Fprintf(client.conn, "[DM] "+msg.sender+": "+msg.body)
					if err != nil {
						log.Print("Can't send message to " + msg.recipient + "\n")
					}
					log.Print("DM from " + msg.sender + " to " + msg.recipient + ": " + msg.body)
					sent = true
				}
			}

			// in case the server can't send the message to a user
			if sent == false {
				for name, client := range *activeClients.Map() {
					if name == msg.sender && client.isActive == true {
						_, err := fmt.Fprintf(client.conn, "Failed to send DM: "+msg.recipient+" is not in the chat\n")
						if err != nil {
							log.Print("Can't send message to " + client.name + "\n")
						}
						sent = true
					}
				}
			}
		}
	}
}

// reads from the publicMessageChan and send public messages to all the active users
func sendPublicMessages(publicMessagesChan chan PublicMessage) {
	for {
		select {
		case msg := <-publicMessagesChan:
			for name, client := range *activeClients.Map() {
				if name != msg.sender && client.isActive == true {
					_, err := fmt.Fprintf(client.conn, msg.sender+": "+msg.body)
					if err != nil {
						log.Print("Can't send message to " + client.name + "\n")
					}
				}
			}
			log.Print(msg.sender + ": " + msg.body)
		}
	}
}

// reads from the clientsChan and updates the ActiveClient struct
func updateActiveClients(clientsChan chan Client) {
	for {
		select {
		case c := <-clientsChan:
			activeClients.DeleteClient(c.name)
			activeClients.AddClient(c.name, c)
			activeClients.CleanUp()
		}
	}
}

// handles !quit
func handleQuit(client Client, publicMessagesChan chan PublicMessage, clientsChan chan Client) {
	publicMessagesChan <- PublicMessage{
		sender: "Server",
		body:   fmt.Sprintf("%s (%s) left the chat\n", client.name, client.id),
	}
	client.isActive = false
	clientsChan <- client
}

// handle !name
func handleName(msg string, client Client, publicMessagesChan chan PublicMessage, privateMessagesChan chan PrivateMessage, clientsChan chan Client) {
	msgSplit := strings.Split(msg, " ")
	name := msgSplit[1]
	name = name[:len(name)-1]
	if activeClients.HasKey(name) {

		// in case the user chooses a username already used by an active user the client id is used as name
		if client.isNameSet == false {
			client.isActive = false
			clientsChan <- client

			client.name = client.id
			client.isActive = true
			clientsChan <- client

			publicMessagesChan <- PublicMessage{
				sender: "Server",
				body:   fmt.Sprintf("Client %s is %s\n", client.id, client.id),
			}
		}

		privateMessagesChan <- PrivateMessage{
			sender:    "Server",
			body:      fmt.Sprintf("%s is already taken. Use `!name [your-name]` to change it\n", name),
			recipient: client.name,
		}

	} else {
		// in case the user chooses an available name
		client.isActive = false
		clientsChan <- client

		client.name = name
		client.isActive = true
		client.isNameSet = true
		clientsChan <- client

		publicMessagesChan <- PublicMessage{
			sender: "Server",
			body:   fmt.Sprintf("Client %s is %s\n", client.id, client.name),
		}
	}
}

// handles !dm
func handleDM(msg string, client Client, privateMessagesChan chan PrivateMessage, clientsChan chan Client) {
	msgSplit := strings.Split(msg, " ")
	recipient := msgSplit[1]

	privateMessagesChan <- PrivateMessage{
		sender:    client.name,
		body:      fmt.Sprint(strings.Join(msgSplit[2:], " ")),
		recipient: recipient,
	}
}

// handles a new connection once it has been established
func handleConnection(client Client, publicMessagesChan chan PublicMessage, privateMessagesChan chan PrivateMessage, clientsChan chan Client) {

	publicMessagesChan <- PublicMessage{
		sender: "Server",
		body:   fmt.Sprintf("%s (%s) joined the chat\n", client.name, client.id),
	}

	for {
		msg, _ := bufio.NewReader(client.conn).ReadString('\n')

		if msg == "" || strings.HasPrefix(msg, "!quit") {
			handleQuit(client, publicMessagesChan, clientsChan)
			return
		} else if strings.HasPrefix(msg, "!name") {
			handleName(msg, client, publicMessagesChan, privateMessagesChan, clientsChan)
		} else if strings.HasPrefix(msg, "!dm") {
			handleDM(msg, client, privateMessagesChan, clientsChan)
		} else {
			publicMessagesChan <- PublicMessage{
				sender: client.name,
				body:   fmt.Sprint(msg),
			}
		}
	}
}

func main() {

	activeClients = ActiveClients{m: make(map[string]Client)}

	publicMessagesChan := make(chan PublicMessage)
	privateMessagesChan := make(chan PrivateMessage)
	clientsChan := make(chan Client)

	go sendPublicMessages(publicMessagesChan)
	go sendPrivateMessages(privateMessagesChan)
	go updateActiveClients(clientsChan)

	log.Println("Start chat server...")
	server, _ := net.Listen("tcp", ":8000")

	for {
		conn, _ := server.Accept()

		client := Client{
			id:       conn.RemoteAddr().String(),
			name:     conn.RemoteAddr().String(),
			conn:     conn,
			isActive: true,
		}

		go handleConnection(client, publicMessagesChan, privateMessagesChan, clientsChan)
	}
}
