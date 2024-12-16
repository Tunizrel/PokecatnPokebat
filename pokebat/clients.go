package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

type Account struct{
	Username string `json:"Name"`
	Password string `json:"Password"`
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

	drawTitle()
	if !Login() {
		fmt.Println("Authentication failed. Exiting...")
		return
	}

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

func Login() bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter your username: ")
	username, _ := reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Print("Enter your password: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	// Send credentials to the server
	fmt.Fprintf(conn, username+","+password+"\n")

	// Receive server response
	response := make([]byte, 1024)
	length, err := conn.Read(response)
	if err != nil {
		fmt.Println("Error reading from server:", err)
		return false
	}

	fmt.Print(string(response[:length]))
	return strings.Contains(string(response[:length]), "Login successful")
	

}

func
 

func drawTitle() {
	fmt.Println("                                                        ")
	fmt.Println(" ######  ####### #    # ####### ######     #    ####### ")
	fmt.Println(" #     # #     # #   #  #       #     #   # #      #    ")
	fmt.Println(" #     # #     # #  #   #       #     #  #   #     #    ")
	fmt.Println(" ######  #     # ###    #####   ######  #     #    #    ")
	fmt.Println(" #       #     # #  #   #       #     # #######    #    ")
	fmt.Println(" #       #     # #   #  #       #     # #     #    #    ")
	fmt.Println(" #       ####### #    # ####### ######  #     #    #    ")
	fmt.Println("                                                        ")
}
