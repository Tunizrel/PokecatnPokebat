package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	//"math/rand"
	"net"
	"strconv"
	"strings"
)

type Pokemon struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Types       []string          `json:"types"`
	Stats       Stats             `json:"stats"`
	Exp         int               `json:"exp,string"`
	WhenAttacked map[string]string `json:"when_attacked"`
}

type Stats struct {
	HP      int `json:"HP,string"`
	Attack  int `json:"Attack,string"`
	Defense int `json:"Defense,string"`
	Speed   int `json:"Speed,string"`
	SpAtk   int `json:"Sp Atk,string"`
	SpDef   int `json:"Sp Def,string"`
}

type Player struct {
	Name    string
	Pokemons []*Pokemon
	Active  *Pokemon
	Conn    net.Conn
}

func main() {
	// Load Pokémon data
	file, err := ioutil.ReadFile("../pokedex.json")
	if err != nil {
		log.Fatalf("Failed to load pokedex.json: %v", err)
	}

	var pokemons []Pokemon
	err = json.Unmarshal(file, &pokemons)
	if err != nil {
		log.Fatalf("Failed to parse pokedex.json: %v", err)
	}

	// Start server
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer listener.Close()

	fmt.Println("Server started. Waiting for players...")

	players := make([]*Player, 0, 2)
	for len(players) < 2 {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		player := &Player{
			Conn: conn,
		}
		players = append(players, player)
		fmt.Printf("Player %d has joined.\n", len(players))
	}

	// Assign names and let players choose Pokémons
	for i, player := range players {
		player.Conn.Write([]byte(fmt.Sprintf("Enter your name, Player %d: ", i+1)))
		name := make([]byte, 1024)
		n, err := player.Conn.Read(name)
		if err != nil {
			log.Printf("Failed to read player name: %v", err)
			continue
		}
		player.Name = strings.TrimSpace(string(name[:n]))

		for {
			player.Conn.Write([]byte("Choose 3 Pokémon by entering their IDs (separated by space): "))
			choice := make([]byte, 1024)
			n, err = player.Conn.Read(choice)
			if err != nil {
				log.Printf("Failed to read Pokémon choice: %v", err)
				continue
			}
			choices := strings.Fields(string(choice[:n]))

			if len(choices) != 3 {
				player.Conn.Write([]byte("Invalid Pokémon selection. Please select exactly 3 Pokémon.\n"))
				continue
			}

			player.Pokemons = nil
			for _, choice := range choices {
				found := false
				for _, pokemon := range pokemons {
					if pokemon.ID == choice {
						player.Pokemons = append(player.Pokemons, &pokemon)
						found = true
						break
					}
				}
				if !found {
					player.Conn.Write([]byte(fmt.Sprintf("Pokémon with ID %s not found. Please try again.\n", choice)))
					player.Pokemons = nil
					break
				}
			}

			if len(player.Pokemons) == 3 {
				player.Active = player.Pokemons[0]
				break
			}
		}
	}

	// Simplify turn order logic based on speed
	var firstPlayer, secondPlayer *Player
	if players[0].Active.Stats.Speed > players[1].Active.Stats.Speed {
		firstPlayer = players[0]
		secondPlayer = players[1]
	} else {
		firstPlayer = players[1]
		secondPlayer = players[0]
	}

	// Start battle loop
	firstPlayer.Conn.Write([]byte(fmt.Sprintf("%s, prepare for battle!\n", firstPlayer.Name)))
	secondPlayer.Conn.Write([]byte(fmt.Sprintf("%s, prepare for battle!\n", secondPlayer.Name)))
	for {
		for _, player := range []*Player{firstPlayer, secondPlayer} {
			player.Conn.Write([]byte(fmt.Sprintf("Active Pokémon: %s\n", player.Active.Name)))
			player.Conn.Write([]byte("Choose action:\n1. Attack\n2. Switch Pokémon\nEnter your choice: "))

			choice := make([]byte, 1024)
			n, err := player.Conn.Read(choice)
			if err != nil {
				log.Printf("Failed to read player choice: %v", err)
				continue
			}

			switch strings.TrimSpace(string(choice[:n])) {
			case "1":
				damage := calculateDamage(player.Active, secondPlayer.Active)
				secondPlayer.Active.Stats.HP -= damage
				player.Conn.Write([]byte(fmt.Sprintf("You dealt %d damage!\n", damage)))
				secondPlayer.Conn.Write([]byte(fmt.Sprintf("You received %d damage!\n", damage)))

				if secondPlayer.Active.Stats.HP <= 0 {
					secondPlayer.Conn.Write([]byte("Your Pokémon fainted!\n"))
					if allPokemonFainted(secondPlayer) {
						player.Conn.Write([]byte("You win!\n"))
						secondPlayer.Conn.Write([]byte("You lose!\n"))
						return
					}
					switchPokemon(secondPlayer)
				}
			case "2":
				switchPokemon(player)
			default:
				player.Conn.Write([]byte("Invalid choice. Try again.\n"))
			}

			// Switch turns
			firstPlayer, secondPlayer = secondPlayer, firstPlayer
		}
	}
}

func calculateDamage(attacker, defender *Pokemon) int {
	baseDamage := attacker.Stats.Attack - defender.Stats.Defense
	if baseDamage < 0 {
		baseDamage = 0
	}
	return baseDamage
}

func switchPokemon(player *Player) {
	player.Conn.Write([]byte("Choose a Pokémon to switch to:\n"))
	for i, pokemon := range player.Pokemons {
		if pokemon != player.Active && pokemon.Stats.HP > 0 {
			player.Conn.Write([]byte(fmt.Sprintf("%d. %s\n", i+1, pokemon.Name)))
		}
	}

	choice := make([]byte, 1024)
	n, err := player.Conn.Read(choice)
	if err != nil {
		log.Printf("Failed to read Pokémon switch choice: %v", err)
		return
	}

	selectedIndex, err := strconv.Atoi(strings.TrimSpace(string(choice[:n])))
	if err != nil || selectedIndex < 1 || selectedIndex > len(player.Pokemons) || player.Pokemons[selectedIndex-1] == player.Active {
		player.Conn.Write([]byte("Invalid choice. Try again.\n"))
		switchPokemon(player)
		return
	}

	player.Active = player.Pokemons[selectedIndex-1]
	player.Conn.Write([]byte(fmt.Sprintf("Switched to %s\n", player.Active.Name)))
}

func allPokemonFainted(player *Player) bool {
	for _, pokemon := range player.Pokemons {
		if pokemon.Stats.HP > 0 {
			return false
		}
	}
	return true
}
