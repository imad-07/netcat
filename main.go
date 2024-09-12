package main

import (
	"fmt"
	"net"
	"os"
	"strconv"

	"funcs/funcs"
)

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

		go funcs.HandleConnection(conn)
	}
}
