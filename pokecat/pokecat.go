package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Configuration constants
const (
	GridSize         = 20 // Grid size of the world
	PokemonsPerPlayer = 3 // Number of Pokémon assigned per player
)

// Pokemon represents the structure of a Pokémon
type Pokemon struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Types       []string          `json:"types"`
	Stats       map[string]string `json:"stats"`
	Exp         string            `json:"exp"`
	WhenAttacked map[string]string `json:"when_attacked"`
	X           int               // X coordinate on the grid
	Y           int               // Y coordinate on the grid
}

// Player represents a player in the game
type Player struct {
	Conn net.Conn
}

var (
	pokemons []Pokemon
	mutex    sync.Mutex // Mutex for safe access to shared data
)

func main() {
	// Load Pokémon data from pokedex.json file
	if err := loadPokemonData("../pokedex.json"); err != nil {
		log.Fatalf("Failed to load Pokémon data: %v", err)
	}

	// Start the server
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	// Handle graceful shutdown on Ctrl+C
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-shutdown
		fmt.Println("\nShutting down server...")
		listener.Close()
		os.Exit(0)
	}()

	fmt.Println("Server started. Waiting for players...")

	// Accept incoming connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Connection error: %v", err)
			continue
		}

		log.Printf("Player connected from %s", conn.RemoteAddr())
		// Handle each player connection in a separate goroutine
		go handlePlayer(conn)
	}
}

// loadPokemonData loads Pokémon data from a JSON file
func loadPokemonData(filename string) error {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load Pokémon data file: %v", err)
	}
	if err := json.Unmarshal(file, &pokemons); err != nil {
		return fmt.Errorf("failed to parse Pokémon data: %v", err)
	}
	log.Printf("Loaded %d Pokémon from %s", len(pokemons), filename)
	return nil
}

// chooseRandomPokemons selects unique random Pokémon and ensures positions are within the grid
func chooseRandomPokemons() []Pokemon {
	selectedPokemons := make([]Pokemon, 0, PokemonsPerPlayer)
	uniqueIndexes := make(map[int]struct{}) // Track selected Pokémon to avoid duplicates

	for len(selectedPokemons) < PokemonsPerPlayer {
		index := rand.Intn(len(pokemons))
		if _, exists := uniqueIndexes[index]; exists {
			continue
		}
		uniqueIndexes[index] = struct{}{}

		pokemon := pokemons[index]
		pokemon.X = rand.Intn(GridSize) // Ensure within grid bounds
		pokemon.Y = rand.Intn(GridSize)
		selectedPokemons = append(selectedPokemons, pokemon)
	}

	return selectedPokemons
}

// handlePlayer handles each player's connection
func handlePlayer(conn net.Conn) {
	defer func() {
		log.Printf("Player disconnected: %s", conn.RemoteAddr())
		conn.Close()
	}()
	if !authenticatePlayer(conn) {
		log.Printf("Authentication failed for %s", conn.RemoteAddr())
		conn.Close()
		return
	}
	mutex.Lock()
	selectedPokemons := chooseRandomPokemons()
	mutex.Unlock()

	// Convert Pokémon data to JSON
	data, err := json.Marshal(selectedPokemons)
	if err != nil {
		log.Printf("Failed to serialize Pokémon data for %s: %v", conn.RemoteAddr(), err)
		return
	}

	// Send JSON data to client
	if _, err := conn.Write(data); err != nil {
		log.Printf("Failed to send Pokémon data to %s: %v", conn.RemoteAddr(), err)
		return
	}

	log.Printf("Sent %d Pokémon to %s", len(selectedPokemons), conn.RemoteAddr())
	for _, pokemon := range selectedPokemons {
		log.Printf("Sent Pokémon: ID=%s, Name=%s",
			pokemon.ID, pokemon.Name)
	}
}
// authenticatePlayer authenticates the player using the provided credentials
func authenticatePlayer(conn net.Conn) bool {
	buffer := make([]byte, 2048)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Printf("Failed to read authentication data: %v", err)
		return false
	}

	var authData map[string]string
	if err := json.Unmarshal(buffer[:n], &authData); err != nil {
		log.Printf("Failed to parse authentication data: %v", err)
		return false
	}

	accounts, err := loadAccountsData("../accounts.json")
	if err != nil {
		log.Printf("Failed to load accounts data: %v", err)
		return false
	}

	for _, account := range accounts {
		if account["Name"] == authData["name"] && account["Password"] == authData["password"] {
			response := map[string]string{"status": "success"}
			responseBytes, _ := json.Marshal(response)
			conn.Write(responseBytes)
			return true
		}
	}

	log.Println("Authentication failed. Exiting.")
	response := map[string]string{"status": "failure"}
	responseBytes, _ := json.Marshal(response)
	conn.Write(responseBytes)
	return false
}

// loadAccountsData loads account data from a JSON file
func loadAccountsData(filename string) ([]map[string]string, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to load accounts data file: %v", err)
	}
	var accounts []map[string]string
	if err := json.Unmarshal(file, &accounts); err != nil {
		return nil, fmt.Errorf("failed to parse accounts data: %v", err)
	}
	log.Printf("Loaded %d accounts from %s", len(accounts), filename)
	return accounts, nil
}
