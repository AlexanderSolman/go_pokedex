package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	pokecache "github.com/AlexanderSolman/go_pokedex/internal"
)

type cliCommand struct {
	name        string
	description string
}

type jsonLocationResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

type jsonLocationExplore struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	BaseExperience int    `json:"base_experience"`
	Height         int    `json:"height"`
	Name           string `json:"name"`
	Stats          []struct {
		BaseStat int `json:"base_stat"`
		Effort   int `json:"effort"`
		Stat     struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"type"`
	} `json:"types"`
	Weight int `json:"weight"`
	caught bool
}

type pokedex struct {
	m map[string]Pokemon
}

func commands() map[string]cliCommand {
	return map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
		},
		"map": {
			name:        "map",
			description: "Display 20 location areas",
		},
		"mapb": {
			name:        "mapb",
			description: "Display previous 20 location areas",
		},
		"explore": {
			name:        "explore",
			description: "explore <location-area> lists pokemon in the area",
		},
		"catch": {
			name:        "catch",
			description: "Try to catch a Pokémon and add it to the pokedex",
		},
		"inspect": {
			name:        "inspect",
			description: "Display information about caught pokemon",
		},
		"pokedex": {
			name:        "pokedex",
			description: "Display all caught pokemon",
		},
	}
}

func parseJsonLocation(res *http.Response, locationArea *jsonLocationResponse) []byte {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	errr := json.Unmarshal(body, &locationArea)
	if errr != nil {
		fmt.Println(errr, "\nCould not parse json")
		return nil
	}

	addLocations := []byte{}
	for _, i := range locationArea.Results {
		fmt.Println(i.Name)
		addLocations = append(addLocations, []byte(i.Name+"\n")...)
	}
	return addLocations
}

func parseJsonExplore(res *http.Response, locationExplore *jsonLocationExplore) []byte {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	errr := json.Unmarshal(body, &locationExplore)
	if errr != nil {
		fmt.Println(errr, "\nCould not parse json")
		return nil
	}

	addExplored := []byte{}
	for _, i := range locationExplore.PokemonEncounters {
		fmt.Println("-", i.Pokemon.Name)
		addExplored = append(addExplored, []byte("- "+i.Pokemon.Name+"\n")...)
	}
	return addExplored
}

func parsePokemon(res *http.Response, pokemon *Pokemon) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	errr := json.Unmarshal(body, &pokemon)
	if errr != nil {
		fmt.Println(errr, "\nCould not parse json")
		return
	}
}

func catchingThePokemon(exp int) bool {
	// Base exp is our limit, if we roll > 0.75 * base exp we catch the pokemon
	catchChance := rand.Intn(exp)
	if float64(catchChance) >= (float64(exp) * 0.75) {
		return true
	}
	return false
}

func printInspectPokemon(pokemon Pokemon) {
	fmt.Printf("Name: %s\nHeight: %v\nWeight: %v\nStats:\n", pokemon.Name, pokemon.Height, pokemon.Weight)
	for _, i := range pokemon.Stats {
		fmt.Printf("\t-%s: %v\n", i.Stat.Name, i.BaseStat)
	}
	fmt.Println("Types:")
	for _, i := range pokemon.Types {
		fmt.Printf("\t- %s\n", i.Type.Name)
	}

}

func main() {
	var locationArea jsonLocationResponse
	var locationExplore jsonLocationExplore
	var pokemon Pokemon

	cache := pokecache.NewCache(5 * time.Minute) // Cache created at start with 5min interval
	thePokedex := pokedex{m: make(map[string]Pokemon)}

	for {
		fmt.Print("pokedex > ")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}

		m_com := commands()
		splitString := strings.Split(scanner.Text(), " ")

		switch splitString[0] {
		case "help":
			fmt.Println("\nHow to use the Pokedex:\n\n")
			fmt.Println("help:     ", m_com["help"].description)
			fmt.Println("map:      ", m_com["map"].description)
			fmt.Println("mapb:     ", m_com["mapb"].description)
			fmt.Println("explore:  ", m_com["explore"].description)
			fmt.Println("inspect:  ", m_com["inspect"].description)
			fmt.Println("pokedex:  ", m_com["pokedex"].description)
			fmt.Println("exit:     ", m_com["exit"].description, "\n")
		case "map":
			// Initial call to the API given no call has been made.
			// Data gets printed and added to the cache
			if locationArea.Next == "" {
				res, err := http.Get("https://pokeapi.co/api/v2/location-area/")
				if err != nil {
					log.Fatal(err)
				}
				data := parseJsonLocation(res, &locationArea)
				cache.Add("https://pokeapi.co/api/v2/location-area/?offset=0&limit=20", data, locationArea.Next, locationArea.Previous)
			} else {
				// Checks if data is cached and prints else calls API for it and adds to cache
				if i, ok, n, p := cache.Get(locationArea.Next); ok {
					fmt.Println("From cache: \n")
					fmt.Println(string(i))
					locationArea.Next = n
					locationArea.Previous = p
				} else {
					url := locationArea.Next
					res, err := http.Get(locationArea.Next)
					if err != nil {
						log.Fatal(err)
					}
					data := parseJsonLocation(res, &locationArea)
					cache.Add(url, data, locationArea.Next, locationArea.Previous)
				}
			}
		case "mapb":
			if locationArea.Previous == "" {
				fmt.Println("There were no previous locations")
			} else {
				// Checks if data is cached and prints else calls API for it and adds to cache
				if i, ok, n, p := cache.Get(locationArea.Previous); ok {
					fmt.Println("From cache: \n")
					fmt.Println(string(i))
					locationArea.Next = n
					locationArea.Previous = p
				} else {
					res, err := http.Get(locationArea.Previous)
					if err != nil {
						log.Fatal(err)
					}
					data := parseJsonLocation(res, &locationArea)
					cache.Add(locationArea.Previous, data, locationArea.Next, locationArea.Previous)
				}
			}
		case "explore":
			if i, ok, _, _ := cache.Get("https://pokeapi.co/api/v2/location-area/" + splitString[1]); ok {
				fmt.Println("From cache: \n")
				fmt.Println(string(i))
			} else {
				res, err := http.Get("https://pokeapi.co/api/v2/location-area/" + splitString[1])
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("Exploring %s...\n", splitString[1])
				data := parseJsonExplore(res, &locationExplore)
				cache.Add("https://pokeapi.co/api/v2/location-area/"+splitString[1], data, "", "")
			}
		case "catch":
			if !thePokedex.m[splitString[1]].caught {
				if _, ok := thePokedex.m[splitString[1]]; !ok {
					res, err := http.Get("https://pokeapi.co/api/v2/pokemon/" + splitString[1])
					if err != nil {
						fmt.Println("No pokemon named", splitString[1])
						log.Fatal(err)
					}
					parsePokemon(res, &pokemon)
					thePokedex.m[splitString[1]] = pokemon
				}
				fmt.Printf("Throwing a Pokeball at %s...\n", splitString[1])
				time.Sleep(2 * time.Second)
				if catch := catchingThePokemon(thePokedex.m[splitString[1]].BaseExperience); catch {
					fmt.Println(splitString[1], "was caught!")
					pokemon.caught = true
					thePokedex.m[splitString[1]] = pokemon
					pokemon.caught = false
				} else {
					fmt.Println(splitString[1], "escaped!")
				}
			} else {
				fmt.Println(splitString[1], "has already been caught")
			}
		case "inspect":
			if i, ok := thePokedex.m[splitString[1]]; ok {
				if thePokedex.m[splitString[1]].caught == true {
					printInspectPokemon(i)
				} else {
					fmt.Println(splitString[1], "has not been caught yet!")
				}
			} else {
				fmt.Println(splitString[1], "has not been caught yet!")
			}
		case "pokedex":
			fmt.Println("The Pokedex:")
			for _, v := range thePokedex.m {
				if v.caught == true {
					fmt.Printf("\t- %s\n", v.Name)
				}
			}
		case "exit":
			return
		default:
			fmt.Println("Type <help> for information on usage")
		}
	}
}
