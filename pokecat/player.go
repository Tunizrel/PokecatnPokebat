package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"github.com/eiannone/keyboard"
	"sync"
	"time"
)

const GridSize = 20

type Pokemon struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Types        []string          `json:"types"`
	Stats        map[string]string `json:"stats"`
	Exp          string            `json:"exp"`
	WhenAttacked map[string]string `json:"when_attacked"`
	X            int
	Y            int
}

var lastNotification string
var grid [GridSize][GridSize]rune
var playerX, playerY int
var pokemons []Pokemon
var mu sync.Mutex

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\nExiting the game. Goodbye!")
		keyboard.Close()
		os.Exit(0)
	}()

	var playerName, password string
	fmt.Print("Enter your username: ")
	fmt.Scanln(&playerName)

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close()

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

	// Initialize player position
	playerX, playerY = GridSize/2, GridSize/2

	// Receive Pokémon data from server
	n, err = conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read Pokémon data: %v\n", err)
		return
	}

	fmt.Printf("Received data: %s\n", string(buffer[:n]))

	err = json.Unmarshal(buffer[:n], &pokemons)
	if err != nil {
		fmt.Printf("Failed to parse Pokémon data: %v\n", err)
		return
	}

	initGrid()

	err = keyboard.Open()
	if err != nil {
		fmt.Printf("Failed to initialize keyboard: %v\n", err)
		return
	}
	defer keyboard.Close()

	for {
		printGrid()
		_, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Printf("Error reading keyboard input: %v\n", err)
			break
		}

		if handleMovement(key) {
			break
		}

		checkCapture(playerName)
	}
}

func initGrid() {
	clearGrid()
	drawTitle()
	for _, p := range pokemons {
		if p.X < GridSize && p.Y < GridSize {
			grid[p.Y][p.X] = '❓'
		}
	}
	grid[playerY][playerX] = '💂'
}

func clearGrid() {
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			grid[i][j] = '.'
		}
	}
}

func printGrid() {
	clearScreen()
	drawTitle()
	for _, row := range grid {
		for _, cell := range row {
			fmt.Printf("%c ", cell)
		}
		fmt.Println()
	}

	if lastNotification != "" {
		fmt.Println("\n" + lastNotification)
		lastNotification = ""
	}
	if len(pokemons) == 0 {
		drawCongrats()
	}
}

func drawCongrats() {
	fmt.Println("░█████╗░░█████╗░███╗░░██╗░██████╗░██████╗░░█████╗░████████╗░██████╗")
	fmt.Println("██╔══██╗██╔══██╗████╗░██║██╔════╝░██╔══██╗██╔══██╗╚══██╔══╝██╔════╝")
	fmt.Println("██║░░╚═╝██║░░██║██╔██╗██║██║░░██╗░██████╔╝███████║░░░██║░░░╚█████╗░")
	fmt.Println("██║░░██╗██║░░██║██║╚████║██║░░╚██╗██╔══██╗██╔══██║░░░██║░░░░╚═══██╗")
	fmt.Println("╚█████╔╝╚█████╔╝██║░╚███║╚██████╔╝██║░░██║██║░░██║░░░██║░░░██████╔╝")
	fmt.Println("░╚════╝░░╚════╝░╚═╝░░╚══╝░╚═════╝░╚═╝░░╚═╝╚═╝░░╚═╝░░░╚═╝░░░╚═════╝░")
	fmt.Println("\n Congratulations! You've caught all the Pokémon!")
	fmt.Println(" Exiting the game. Goodbye!")
	fmt.Println("=========================================================================\n")
}

func drawTitle() {
	fmt.Println("                                  ,'\\")
	fmt.Println("    _.----.        ____         ,'  _\\   ___    ___     ____")
	fmt.Println("_,-'       `.     |    |  /`.   \\,-'    |   \\  /   |   |    \\  |`.")
	fmt.Println("\\      __    \\    '-.  | /   `.  ___    |    \\/    |   '-.   \\ |  |")
	fmt.Println(" \\.    \\ \\   |  __  |  |/    ,','_  `.  |          | __  |    \\|  |")
	fmt.Println("   \\    \\/   /,' _`.|      ,' / / / /   |          ,' _`.|     |  |")
	fmt.Println("    \\     ,-'/  / \\ \\    ,'   | \\/ / ,`.|         /  / \\ \\  |     |")
	fmt.Println("     \\    \\ |   \\_/  |   `-.  \\    `'  /|  |    ||   \\_/  | |\\    |")
	fmt.Println("      \\    \\ \\      /       `-.`.___,-' |  |\\  /| \\      /  | |   |")
	fmt.Println("       \\    \\ `.__,'|  |`-._    `|      |__| \\/ |  `.__,'|  | |   |")
	fmt.Println("        \\_.-'       |__|    `-._ |              '-.|     '-.| |   |")
	fmt.Println("                                `'                            '-._|")
}

func handleMovement(key keyboard.Key) bool {
	grid[playerY][playerX] = '.'

	switch key {
	case keyboard.KeyArrowUp:
		if playerY > 0 {
			playerY--
		}
	case keyboard.KeyArrowDown:
		if playerY < GridSize-1 {
			playerY++
		}
	case keyboard.KeyArrowLeft:
		if playerX > 0 {
			playerX--
		}
	case keyboard.KeyArrowRight:
		if playerX < GridSize-1 {
			playerX++
		}
	case keyboard.KeyEsc:
		return true
	}

	grid[playerY][playerX] = '💂'
	return false
}

func clearScreen() {
	cmd := exec.Command("clear")
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

var caughtPokemons []Pokemon

func savePlayerData(playerName string, pokemons []Pokemon) {
	mu.Lock()
	defer mu.Unlock()

	// Open the player data file
	file, err := os.OpenFile("./player_data.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Failed to open player data file: %v\n", err)
		return
	}
	defer file.Close()

	var allPlayers []map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&allPlayers)
	if err != nil && err.Error() != "EOF" {
		fmt.Printf("Failed to read player data: %v\n", err)
		return
	}

	// Prepare the Pokémon data to be added
	var cleanedPokemons []map[string]interface{}
	for _, p := range pokemons {
		cleanedPokemon := map[string]interface{}{
			"id":            p.ID,
			"name":          p.Name,
			"types":         p.Types,
			"stats":         p.Stats,
			"exp":           p.Exp,
			"when_attacked": p.WhenAttacked,
		}
		cleanedPokemons = append(cleanedPokemons, cleanedPokemon)
	}

	playerFound := false
	for i, player := range allPlayers {
		if player["player_name"] == playerName {
			// Check if player already has the Pokémon
			if existingPokemons, ok := player["pokemons"].([]interface{}); ok {
				existingIDs := make(map[string]bool)
				for _, ep := range existingPokemons {
					if epMap, ok := ep.(map[string]interface{}); ok {
						if id, ok := epMap["id"].(string); ok {
							existingIDs[id] = true
						}
					}
				}

				// Add only new Pokémon
				for _, newPokemon := range cleanedPokemons {
					if id, ok := newPokemon["id"].(string); ok {
						if !existingIDs[id] {
							existingPokemons = append(existingPokemons, newPokemon)
						}
					}
				}

				// Update the player's Pokémon list
				allPlayers[i]["pokemons"] = existingPokemons
			} else {
				// If "pokemons" is not correctly structured, replace it
				allPlayers[i]["pokemons"] = cleanedPokemons
			}
			playerFound = true
			break
		}
	}

	if !playerFound {
		// Add a new player if not found
		playerData := map[string]interface{}{
			"player_name": playerName,
			"pokemons":    cleanedPokemons,
		}
		allPlayers = append(allPlayers, playerData)
	}

	// Write updated data back to the file
	file.Seek(0, 0)
	file.Truncate(0) // Clear existing content

	data, err := json.MarshalIndent(allPlayers, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal player data: %v\n", err)
		return
	}

	_, err = file.Write(data)
	if err != nil {
		fmt.Printf("Failed to write player data to file: %v\n", err)
	}
}




func checkCapture(playerName string) {
	for i, p := range pokemons {
		if p.X == playerX && p.Y == playerY {
			lastNotification = fmt.Sprintf("You caught a Pokémon: %s (ID: %s)!", p.Name, p.ID)
			caughtPokemons = append(caughtPokemons, p)
			pokemons = append(pokemons[:i], pokemons[i+1:]...)
			grid[p.Y][p.X] = '👍'

			if len(pokemons) == 0 {
				savePlayerData(playerName, caughtPokemons)
				grid[playerY][playerX] = '🏆'
				printGrid()
				keyboard.Close()

				os.Exit(0)
			}
			return
		}
	}
}
