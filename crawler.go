package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly"
	"github.com/PuerkitoBio/goquery"
)

type Pokemon struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Types        []string          `json:"types"`
	Stats        map[string]string `json:"stats"`
	EXP          string            `json:"exp"`
	WhenAttacked map[string]string `json:"when_attacked"`
}

func main() {
	// Create context for chromedp
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Extend the timeout for our operations
	ctx, cancel = context.WithTimeout(ctx, 999999*time.Second)
	defer cancel()

	var pokemons []Pokemon

	// Step 1: Scrape Pokémon data from Pokedex.org (first 3 Pokémon)
	for i := 1; i <= 100; i++ {
		var pokemon Pokemon
		err := chromedp.Run(ctx,
			chromedp.Navigate(fmt.Sprintf("https://pokedex.org/#/pokemon/%d", i)),
			chromedp.Sleep(2*time.Second),
			chromedp.Evaluate(`document.querySelector(".detail-header .detail-national-id").innerText.replace("#", "")`, &pokemon.ID),
			chromedp.Evaluate(`document.querySelector(".detail-panel-header").innerText`, &pokemon.Name),
			chromedp.Evaluate(`Array.from(document.querySelectorAll('.detail-types span.monster-type')).map(elem => elem.innerText)`, &pokemon.Types),
			chromedp.Evaluate(`Object.fromEntries(Array.from(document.querySelectorAll('.detail-stats-row')).map(row => {
				const label = row.querySelector('span:first-child').innerText;
				const value = row.querySelector('.stat-bar-fg').innerText;
				return [label, value];
			}))`, &pokemon.Stats),
		)
		if err != nil {
			fmt.Println("Failed to extract data for ID %d: %v", i, err)
			continue
		}
		pokemons = append(pokemons, pokemon)
		fmt.Printf("Crawled data for Pokemon ID %d\n", i)
	}

	// Step 2: Scrape EXP data from Bulbapedia
	c := colly.NewCollector(
		colly.AllowedDomains("bulbapedia.bulbagarden.net"),
	)

	// Create a map to hold EXP data by Pokémon ID
	expMap := make(map[string]string)

	// On each row of the Pokémon table (excluding the header row)
	c.OnHTML("table.sortable tbody tr", func(e *colly.HTMLElement) {
		// Extract Pokémon ID and EXP from the first and fourth columns
		id := strings.Trim(e.ChildText("td:nth-child(1)"), "\n ")
		exp := strings.Trim(e.ChildText("td:nth-child(4)"), "\n ")

		// Remove leading zeros from ID
		id = strings.TrimLeft(id, "0")

		// If both ID and EXP are valid, add to the expMap
		if id != "" && exp != "" {
			expMap[id] = exp
		}
	})

	// Handle errors during scraping
	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Error:", err)
	})

	// Visit the Bulbapedia page with the EXP data
	c.Visit("https://bulbapedia.bulbagarden.net/wiki/List_of_Pok%C3%A9mon_by_effort_value_yield")

	// Step 3: Merge the EXP data with existing Pokémon data
	for i := range pokemons {
		if exp, found := expMap[pokemons[i].ID]; found {
			pokemons[i].EXP = exp
		}
	}

	// Step 4: Scrape "When Attacked" data from Pokedex.org using chromedp
	basePokedexURL := "https://pokedex.org/#/pokemon/"
	for i := range pokemons {
		fmt.Printf("Fetching When Attacked data for %s...\n", pokemons[i].Name)
		url := fmt.Sprintf("%s%s", basePokedexURL, pokemons[i].ID)

		var whenAttackedHTML string
		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.Sleep(5*time.Second),
			chromedp.OuterHTML("div.when-attacked", &whenAttackedHTML), // Capture the HTML for "when attacked"
		)
		if err != nil {
			log.Printf("Failed to fetch When Attacked data for %s: %v", pokemons[i].Name, err)
			continue
		}

		// Parse the HTML with goquery
		whenAttacked := map[string]string{}
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(whenAttackedHTML))
		if err != nil {
			log.Printf("Failed to parse When Attacked HTML for %s: %v", pokemons[i].Name, err)
			continue
		}

		doc.Find("div.when-attacked-row").Each(func(j int, row *goquery.Selection) {
			row.Find("span.monster-type").Each(func(k int, t *goquery.Selection) {
				multiplier := strings.TrimSpace(t.Next().Text())
				typeName := strings.ToLower(strings.TrimSpace(t.Text()))
				if typeName != "" && multiplier != "" {
					whenAttacked[typeName] = multiplier
				}
			})
		})

		pokemons[i].WhenAttacked = whenAttacked
		fmt.Printf("Fetched When Attacked data for %s\n", pokemons[i].Name)
	}

	// Step 5: Save the merged Pokémon data to a JSON file
	file, err := os.Create("pokedex.json")
	fmt.Println("Saving Pokedex data to pokedex.json...")
	if err != nil {
		log.Fatal("Cannot create file", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(pokemons); err != nil {
		log.Fatal("Cannot encode to JSON", err)
	}

	fmt.Println("Pokedex data successfully saved to pokedex.json!")
}
