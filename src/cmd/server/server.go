package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/alessandrobessi/gochat/src/pkg/activeclients"
	"github.com/alessandrobessi/gochat/src/pkg/types"
)

var activeClients activeclients.ActiveClients

// reads from the privateMessageChan and send private messages to users
func sendPrivateMessages(privateMessagesChan chan types.PrivateMessage) {
	for {
		select {
		case msg := <-privateMessagesChan:
			sent := false
			for name, client := range *activeClients.GetMap() {
				if name == msg.Recipient && client.IsActive == true {
					_, err := fmt.Fprintf(client.Conn, "[DM] "+msg.Sender+": "+msg.Body)
					if err != nil {
						log.Print("Can't send message to " + msg.Recipient + "\n")
					}
					log.Print("DM from " + msg.Sender + " to " + msg.Recipient + ": " + msg.Body)
					sent = true
				}
			}

			// in case the server can't send the message to a user
			if sent == false {
				for name, client := range *activeClients.GetMap() {
					if name == msg.Sender && client.IsActive == true {
						_, err := fmt.Fprintf(client.Conn, "Failed to send DM: "+msg.Recipient+" is not in the chat\n")
						if err != nil {
							log.Print("Can't send message to " + client.Name + "\n")
						}
						sent = true
					}
				}
			}
		}
	}
}

// reads from the publicMessageChan and send public messages to all the active users
func sendPublicMessages(publicMessagesChan chan types.PublicMessage) {
	for {
		select {
		case msg := <-publicMessagesChan:
			for name, client := range *activeClients.GetMap() {
				if name != msg.Sender && client.IsActive == true {
					_, err := fmt.Fprintf(client.Conn, msg.Sender+": "+msg.Body)
					if err != nil {
						log.Print("Can't send message to " + client.Name + "\n")
					}
				}
			}
			log.Print(msg.Sender + ": " + msg.Body)
		}
	}
}

// reads from the clientsChan and updates the ActiveClient struct
func updateActiveClients(clientsChan chan types.Client) {
	for {
		select {
		case c := <-clientsChan:
			activeClients.DeleteClient(c.Name)
			activeClients.AddClient(c.Name, c)
			activeClients.CleanUp()
		}
	}
}

// handles !quit
func handleQuit(client types.Client, publicMessagesChan chan types.PublicMessage, clientsChan chan types.Client) {
	publicMessagesChan <- types.PublicMessage{
		Sender: "Server",
		Body:   fmt.Sprintf("%s (%s) left the chat\n", client.Name, client.ID),
	}
	client.IsActive = false
	clientsChan <- client
}

// handle !name
func handleName(msg string, client types.Client, publicMessagesChan chan types.PublicMessage, privateMessagesChan chan types.PrivateMessage, clientsChan chan types.Client) {
	msgSplit := strings.Split(msg, " ")
	name := msgSplit[1]
	name = name[:len(name)-1]
	if activeClients.HasKey(name) {

		// in case the user chooses a username already used by an active user the client id is used as name
		if client.IsNameSet == false {
			client.IsActive = false
			clientsChan <- client

			client.Name = client.ID
			client.IsActive = true
			clientsChan <- client

			publicMessagesChan <- types.PublicMessage{
				Sender: "Server",
				Body:   fmt.Sprintf("Client %s is %s\n", client.ID, client.ID),
			}
		}

		privateMessagesChan <- types.PrivateMessage{
			Sender:    "Server",
			Body:      fmt.Sprintf("%s is already taken. Use `!name [your-name]` to change it\n", name),
			Recipient: client.Name,
		}

	} else {
		// in case the user chooses an available name
		client.IsActive = false
		clientsChan <- client

		client.Name = name
		client.IsActive = true
		client.IsNameSet = true
		clientsChan <- client

		publicMessagesChan <- types.PublicMessage{
			Sender: "Server",
			Body:   fmt.Sprintf("Client %s is %s\n", client.ID, client.Name),
		}
	}
}

// handles !dm
func handleDM(msg string, client types.Client, privateMessagesChan chan types.PrivateMessage, clientsChan chan types.Client) {
	msgSplit := strings.Split(msg, " ")
	recipient := msgSplit[1]

	privateMessagesChan <- types.PrivateMessage{
		Sender:    client.Name,
		Body:      fmt.Sprint(strings.Join(msgSplit[2:], " ")),
		Recipient: recipient,
	}
}

// handles a new connection once it has been established
func handleConnection(client types.Client, publicMessagesChan chan types.PublicMessage, privateMessagesChan chan types.PrivateMessage, clientsChan chan types.Client) {

	publicMessagesChan <- types.PublicMessage{
		Sender: "Server",
		Body:   fmt.Sprintf("%s (%s) joined the chat\n", client.Name, client.ID),
	}

	for {
		msg, _ := bufio.NewReader(client.Conn).ReadString('\n')

		if msg == "" || strings.HasPrefix(msg, "!quit") {
			handleQuit(client, publicMessagesChan, clientsChan)
			return
		} else if strings.HasPrefix(msg, "!name") {
			handleName(msg, client, publicMessagesChan, privateMessagesChan, clientsChan)
		} else if strings.HasPrefix(msg, "!dm") {
			handleDM(msg, client, privateMessagesChan, clientsChan)
		} else {
			publicMessagesChan <- types.PublicMessage{
				Sender: client.Name,
				Body:   fmt.Sprint(msg),
			}
		}
	}
}

func main() {

	activeClients = activeclients.ActiveClients{Map: make(map[string]types.Client)}

	publicMessagesChan := make(chan types.PublicMessage)
	privateMessagesChan := make(chan types.PrivateMessage)
	clientsChan := make(chan types.Client)

	go sendPublicMessages(publicMessagesChan)
	go sendPrivateMessages(privateMessagesChan)
	go updateActiveClients(clientsChan)

	log.Println("Start chat server...")
	server, _ := net.Listen("tcp", ":8000")

	for {
		conn, _ := server.Accept()

		client := types.Client{
			ID:       conn.RemoteAddr().String(),
			Name:     conn.RemoteAddr().String(),
			Conn:     conn,
			IsActive: true,
		}

		go handleConnection(client, publicMessagesChan, privateMessagesChan, clientsChan)
	}
}
