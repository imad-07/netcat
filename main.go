package main

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

func handleConnection(conn net.Conn) {
	text, err1 := os.ReadFile("t.txt")
	if err1 != nil {
		return
	}

	defer conn.Close()
	reader := bufio.NewReader(conn)
	name := ""
	var err error
	count := 0
	for len(name) == 0 && len(clients) <= 10{
		if len(clients) == 10{
			conn.Write([]byte("Max clients reached\n"))
			return
		}
		conn.Write([]byte("Welcome to TCP-Chat!\n"))
		conn.Write(text)
		conn.Write([]byte("[ENTER YOUR NAME]:"))
		name, err = reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if err != nil || len(name) == 0 {
			conn.Write([]byte("Error reading name\n"))
			fmt.Println("Error reading name")
			count++

		}
		if count == 3 {
			conn.Write([]byte("you ran out of tries\n"))
			return
		}

	}

	fmt.Println("Client connected with name: ", name)

	mu.Lock()
	clients[conn] = name // Add client to the map
	mu.Unlock()

	// Send welcome message
	conn.Write([]byte("Welcome, " + name + "!\n"))
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(name +" disconnected:")
			broadcastMessage(conn, name +" disconnected:")
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			break
		}

		message = strings.TrimSpace(message)

		fmt.Println(name + ": " + message)
		if len(message) != 0 {
			// Broadcast the message to all clients
			broadcastMessage(conn, "["+name+"]"+": "+message)
		}
	}
}

// broadcastMessage sends a message to all connected clients except the sender
func broadcastMessage(sender net.Conn, message string) {
	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		if client != sender {
			date := time.Now().Format("2006-01-02 15:04:05")
			client.Write([]byte("[" + date + "]" + " " + message + "\n"))
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8989")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8989...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}
