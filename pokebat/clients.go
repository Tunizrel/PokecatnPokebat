package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type Player struct{
	Username string `json:"Name"`
	Password string `json:"Password"`
	player_name string `json:"player_name"`
}
func main() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	go readMessages(conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		fmt.Fprintf(conn, strings.TrimSpace(text)+"\n")
	}
}

func readMessages(conn net.Conn) {
	for {
		message := make([]byte, 1024)
		length, err := conn.Read(message)
		if err != nil {
			fmt.Println("Failed to read message from server:", err)
			return
		}
		fmt.Print(string(message[:length]))
	}
}
