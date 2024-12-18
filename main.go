package main

import (
	"fmt"
	"os"
	"os/exec"
	"log"
	"bufio"
	"encoding/json"
)

type Account struct{
	Username string `json:"Name"`
	Password string `json:"Password"`
}

func main() {
	for {
		fmt.Println("Welcome to the Game Hub!")
		fmt.Println("Please choose a game to play:")
		fmt.Println("1. Pokecat")
		fmt.Println("2. Pokebat")
		fmt.Println("3. Exit")
		fmt.Println("4. Create a new account")
		fmt.Print("Enter your choice: ")

		var choice int
		_, err := fmt.Scanln(&choice)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}

		switch choice {
		case 1:
			fmt.Println("Launching Pokecat...")
			cmd := exec.Command("go", "run", "pokecat/player.go")
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Run the player.go script
			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to launch Pokecat: %v\n", err)
			}
		case 2:
			fmt.Println("Launching Pokebat...")
			cmd := exec.Command("go", "run", "pokebat/clients.go")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			// Run the clients.go script
			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to launch Pokebat: %v\n", err)
			}
		case 3:
			fmt.Println("Exiting the Game Hub. Goodbye!")
			os.Exit(0)
		case 4:
			scanner := bufio.NewScanner(os.Stdin)
			fmt.Println("Username: ")
			scanner.Scan()
			username := scanner.Text()
			fmt.Println("Password: ")
			scanner.Scan()
			password := scanner.Text()
			account := Account{
				Username: username,
				Password: password,
			}
		
			// Open or create the accounts.json file
			file, err := os.Open("accounts.json")
			if err != nil {
				log.Fatalf("Failed to open accounts.json: %v", err)
			}
			defer file.Close()
		
			// Read existing accounts from the file
			var accounts []Account
			if err := json.NewDecoder(file).Decode(&accounts); err != nil && err.Error() != "EOF" {
				log.Fatalf("Failed to decode accounts.json: %v", err)
			}
			// Append the new account to the list
			accounts = append(accounts, account)
			
			// Write updated accounts back to the file
			encoder := json.NewEncoder(file)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(accounts); err != nil {
				log.Fatalf("Failed to write to accounts.json: %v", err)
			}
		
			fmt.Println("Account successfully saved!")
		
		
		default:
			fmt.Println("Invalid choice. Please choose a valid option.")
		}
	}
}
