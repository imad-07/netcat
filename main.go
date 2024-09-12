package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	clients = make(map[net.Conn]string) // Map of connected clients
	mu      sync.Mutex                  // Mutex to protect the clients map

)

func broadcasttyping() {
	date := time.Now().Format("2006-01-02 15:04:05")

	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		// if client != sender {
		client.Write([]byte("[" + date + "] " + "[" + clients[client] + "]:"))
		//}
	}
}

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
	for len(name) == 0 && len(clients) <= 10 || isArrowKey(name){
		if len(clients) == 10 {
			conn.Write([]byte("Max clients reached\n"))
			return
		}
		conn.Write([]byte("Welcome to TCP-Chat!\n"))
		conn.Write(text)
		conn.Write([]byte("[ENTER YOUR NAME]:"))
		name, err = reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if err != nil || len(name) == 0 || isArrowKey(name){
			conn.Write([]byte("Error reading name\n"))
			fmt.Println("Error reading name")
			count++

		}
		if count == 3 {
			conn.Write([]byte("you ran out of tries\n"))
			count = 0
			return
		}

	}
	for _, otherName := range clients {
		if name == otherName {
			conn.Write([]byte("User Alredy exist!"))
			conn.Close()
			return
		}
	}
	fmt.Println(" has joined our chat...", name)
	broadcastMessage(conn, "\n"+name+" has joined our chat...")

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
	broadcasttyping()
	for {
		date := time.Now().Format("2006-01-02 15:04:05")

		message, err := reader.ReadString('\n')
		if isArrowKey(message){
			conn.Write([]byte("u cannot write anarrow my freind :)\n"))
			message=""
		}else if err != nil {
			broadcastMessage(conn, "\n"+name+" has left our chat...")
			fmt.Println(name + " has left our chat...")
			broadcasttyping()
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
			broadcasttyping()
		} else {
			conn.Write([]byte("[" + date + "] " + "[" + name + "]:"))
		}

	}
}

// broadcastMessage sends a message to all connected clients except the sender
func broadcastMessage(sender net.Conn, message string) {
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

	file, err := os.OpenFile("l.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

func main() {
	os.Truncate("l.txt", 0)
	args := os.Args[1:]
	port := ":8989"
	if len(args) != 0 {
		num, err := strconv.Atoi(args[0])
		if err != nil || len(args[0]) != 4 {
			fmt.Println("[USAGE]: ./TCPChat $port")
			return
		} else {
			port = ":" + strconv.Itoa(num)
		}
	}
	listener, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Printf("Server is listening on port %s...\n", port)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}
func isArrowKey(s string) bool {
	arrowKeys := []string{
		"\x1b[A", // Up arrow
		"\x1b[B", // Down arrow
		"\x1b[C", // Right arrow
		"\x1b[D", // Left arrow
		"\x1b[H", // Left arrow
	}
	for _, key := range arrowKeys {
		if strings.Contains(s, key) {
			return true
		}
	}
	return false
}
