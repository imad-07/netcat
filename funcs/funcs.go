package funcs

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	clients = make(map[net.Conn]string) // Map of connected clients
	mu      sync.Mutex                  // Mutex to protect the clients map

)

func Messagechecker(str string) (validity bool, s string) {
	validity = true
	if str != "" && str[len(str)-1] == '\n' {
		str = str[:len(str)-1]
	}
	message := ""
	if !Bytechecker(str) {
		validity = false
		message = "u cannot use non-ASCII characters :)\n"
	}
	return validity, message
}

func Bytechecker(str string) bool {
	xx := []byte(str)
	for i := 0; i < len(xx); i++ {
		if xx[i] < 32 || xx[i] > 127 {
			return false
		}
	}
	return true
}

func Welcome(conn net.Conn, text []byte) {
	conn.Write([]byte("Welcome to TCP-Chat!\n"))
}

func Namevalidity(name string) (validity bool, s string) {
	if len(name) == 0 {
		return false, "u cannot be named an empty string:)\n"
	}
	validity, message := Messagechecker(name)
	if !validity {
		return validity, message
	}
	for _, otherName := range clients {
		if name == otherName {
			return false, "name already in use:)\n"
		}
	}
	return true, ""
}

func BroadcastMessage(sender net.Conn, message string) {
	date := time.Now().Format("2006-01-02 15:04:05")

	mu.Lock()
	defer mu.Unlock()
	for client := range clients {
		if client != sender && message != "\n" {
			if strings.HasSuffix(message, "has joined our chat...") || strings.HasSuffix(message, "has left our chat...") {
				client.Write([]byte("\n" + message[1:] + "\n"))
			} else {
				client.Write([]byte("\n[" + date + "]" + " " + message[1:] + "\n"))
			}
		}
	}

	file, err := os.OpenFile("l.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Write the new content to the file
	if !(strings.HasSuffix(message, "has joined our chat...") || strings.HasSuffix(message, "has left our chat...")) {
		_, err = file.WriteString("[" + date + "] " + message + "\n")
	}
	message = ""
	if err != nil {
		fmt.Println("Error writing to file:", err)
	}
}

func Broadcasttyping() {
	date := time.Now().Format("2006-01-02 15:04:05")

	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		// if client != sender {
		client.Write([]byte("[" + date + "] " + "[" + clients[client] + "]:"))
		//}
	}
}

func HandleConnection(conn net.Conn) {
	text, err1 := os.ReadFile("t.txt")
	if err1 != nil {
		return
	}

	defer conn.Close()
	reader := bufio.NewReader(conn)
	name := ""
	var err error
	validity := false
	Welcome(conn, text)
	conn.Write(text)
	conn.Write([]byte("[ENTER YOUR NAME]:"))
	for !validity {
		name, err = reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if err != nil {
			conn.Write([]byte("error reading name"))
			return
		}
		iror := ""
		validity, iror = Namevalidity(name)
		if iror != "" {
			conn.Write([]byte(iror))
			conn.Write([]byte("[ENTER YOUR NAME]:"))
		}
	}
	fmt.Println(name, " has joined our chat...")
	BroadcastMessage(conn, "\n"+name+" has joined our chat...")

	mu.Lock()
	clients[conn] = name // Add client to the map
	mu.Unlock()
	// Send welcome message
	conn.Write([]byte("Welcome, " + name + "!\n"))
	str, err := os.ReadFile("l.txt")
	if err != nil {
		return
	}
	conn.Write([]byte(str))
	Broadcasttyping()
	for {
		date := time.Now().Format("2006-01-02 15:04:05")
		message, err := reader.ReadString('\n')
		validity, iror := Messagechecker(message)
		if !validity{
			conn.Write([]byte(iror))
			message = ""
		}
		if err != nil {
			BroadcastMessage(conn, "\n"+name+" has left our chat...")
			fmt.Println(name + " has left our chat...")
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			Broadcasttyping()
			break
		}
		message = strings.TrimSpace(message)
		if len(message) != 0 {
			// Broadcast the message to all clients
			BroadcastMessage(conn, "["+name+"]"+": "+message)
			Broadcasttyping()
		} else {
			conn.Write([]byte("[" + date + "] " + "[" + name + "]:"))
		}
	}
}
