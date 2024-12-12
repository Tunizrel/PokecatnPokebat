package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal" // Import the signal package to handle interrupts
	"github.com/eiannone/keyboard"
	"sync" // Import the sync package for Mutex
	"time"
)

const GridSize = 20

// Pokemon represents a Pokémon structure
type Pokemon struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Types        []string          `json:"types"`
	Stats        map[string]string `json:"stats"`  // Change to map[string]string to store stats as strings
	Exp          string            `json:"exp"`
	WhenAttacked map[string]string `json:"when_attacked"` // Keep as map[string]string
	X            int
	Y            int
}

var lastNotification string
var grid [GridSize][GridSize]rune
var playerX, playerY int
var pokemons []Pokemon
var mu sync.Mutex // Declare the global mutex

func main() {
	// Capture interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\nExiting the game. Goodbye!")
		keyboard.Close()
		os.Exit(0)
	}()

	// Ask for player name
	var playerName string
	fmt.Print("Enter your name: ")
	fmt.Scanln(&playerName)
	
	fmt.Printf("Welcome %s To Pokecat!!!\n", playerName)
	drawTitle()
	time.Sleep(2 * time.Second)
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Printf("Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		conn.Close()
		keyboard.Close()
	}()

	// Initialize player position
	playerX, playerY = GridSize/2, GridSize/2

	// Receive Pokémon data from server
	buffer := make([]byte, 2048) // Increase buffer size
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("Failed to read Pokémon data: %v\n", err)
		return
	}

	// Log the raw data to debug
	fmt.Printf("Received data: %s\n", string(buffer[:n]))

	err = json.Unmarshal(buffer[:n], &pokemons)
	if err != nil {
		fmt.Printf("Failed to parse Pokémon data: %v\n", err)
		return
	}

	// Initialize grid with player and Pokémon
	initGrid()

	// Enable keyboard input
	err = keyboard.Open()
	if err != nil {
		fmt.Printf("Failed to initialize keyboard: %v\n", err)
		return
	}
	defer keyboard.Close()

	// Game loop
	for {
		printGrid()
		_, key, err := keyboard.GetKey()
		if err != nil {
			fmt.Printf("Error reading keyboard input: %v\n", err)
			break
		}

		// Handle player movement
		if handleMovement(key) {
			break
		}

		// Check for Pokémon capture
		checkCapture(playerName)
	}
}

// Initialize the grid with Pokémon and player
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

// Clear the grid
func clearGrid() {
	for i := 0; i < GridSize; i++ {
		for j := 0; j < GridSize; j++ {
			grid[i][j] = '.'
		}
	}
}

// Print the grid to the terminal
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
		fmt.Println("\n" + lastNotification) // Display the last notification
		lastNotification = ""               // Clear the notification after displaying it
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
// Handle player movement
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

// Clear the terminal screen
func clearScreen() {
	cmd := exec.Command("clear")
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
var caughtPokemons []Pokemon

// Save player data to a JSON file
func savePlayerData(playerName string, pokemons []Pokemon) {
	// Lock the mutex to ensure only one goroutine can access this section
	mu.Lock()
	defer mu.Unlock()

	// Open the file for reading and writing
	file, err := os.OpenFile("../player_data.json", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("Failed to open player data file: %v\n", err)
		return
	}
	defer file.Close()

	// Read the existing data from the file
	var allPlayers []map[string]interface{}
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&allPlayers)
	if err != nil && err.Error() != "EOF" {
		fmt.Printf("Failed to read player data: %v\n", err)
		return
	}

	// Remove X and Y from each Pokémon before saving
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

	// Create new player data with cleaned Pokémon list
	playerData := map[string]interface{}{
		"player_name": playerName,
		"pokemons":    cleanedPokemons,
	}

	// Append the new player data to the slice
	allPlayers = append(allPlayers, playerData)

	// Move the file pointer back to the beginning to overwrite
	file.Seek(0, 0)

	// Marshal the updated data to JSON
	data, err := json.MarshalIndent(allPlayers, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal player data: %v\n", err)
		return
	}

	// Write the updated data to the file
	_, err = file.Write(data)
	if err != nil {
		fmt.Printf("Failed to write player data to file: %v\n", err)
	}
}

// Check if player captured a Pokémon
func checkCapture(playerName string) {
	for i, p := range pokemons {
		if p.X == playerX && p.Y == playerY {
			lastNotification = fmt.Sprintf("You caught a Pokémon: %s (ID: %s, Types: %s, Stats: %s, Exp: %s, When Attacked: %s)!", p.Name, p.ID, p.Types, p.Stats, p.Exp, p.WhenAttacked)
			caughtPokemons = append(caughtPokemons, p) // Add caught Pokémon to the temporary array
			pokemons = append(pokemons[:i], pokemons[i+1:]...) // Remove caught Pokémon
			grid[p.Y][p.X] = '👍' // Keep player position on grid

			// Check if all Pokémon are caught
			if len(pokemons) == 0 {
				savePlayerData(playerName, caughtPokemons)
				grid[playerY][playerX] = '🏆'
				printGrid()
				keyboard.Close()

				os.Exit(0)
			}
			return // Exit after handling the captured Pokémon
		}
	}
}
