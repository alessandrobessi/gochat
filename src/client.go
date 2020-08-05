package main

import "net"
import "fmt"
import "os"
import "bufio"
import "strings"


func receiveMessages(conn net.Conn) {
	for {
		msg, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Print(msg)
	}
}

func main() {
	fmt.Print("Insert a username: ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')

	conn, _ := net.Dial("tcp", "127.0.0.1:8000")
	defer conn.Close()

	fmt.Fprintf(conn, "!name " + name)

	go receiveMessages(conn)
	
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, text + "\n")

		if strings.HasPrefix(text, "!quit") {
			break
		}		
	}
}
