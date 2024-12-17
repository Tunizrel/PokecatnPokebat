package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

<<<<<<< Updated upstream
=======

>>>>>>> Stashed changes
func main() {
	conn, err := net.Dial("tcp", "localhost:8081")
	if err != nil {
		fmt.Println("Failed to connect to server:", err)
		return
	}
	defer conn.Close()

<<<<<<< Updated upstream
=======
	drawTitle()


	var playerName, password string
	fmt.Print("Enter your username: ")
	fmt.Scanln(&playerName)

	// Send player name to server for authentication
	fmt.Print("Enter your password: ")
	fmt.Scanln(&password)
	authData := map[string]string{"name": playerName, "password": password}
	authBytes, _ := json.Marshal(authData)
	conn.Write(authBytes)

	// Receive authentication response
	buffer := make([]byte, 2048)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read authentication response: %v\n", err)
		return
	}

	var authResponse map[string]interface{}
	err = json.Unmarshal(buffer[:n], &authResponse)
	if err != nil {
		fmt.Printf("Failed to parse authentication response: %v\n", err)
		return
	}

	if authResponse["status"] == "success" {
		fmt.Printf("Welcome %s To Pokecat!!!\n", playerName)
		drawTitle()
		time.Sleep(2 * time.Second)
	} else {
		fmt.Println("Authentication failed. Exiting.")
		return
	}


>>>>>>> Stashed changes
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
<<<<<<< Updated upstream
=======


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
>>>>>>> Stashed changes
