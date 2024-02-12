package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	pokecache "github.com/AlexanderSolman/go_pokedex/internal"
)

type cliCommand struct {
	name        string
	description string
}

type jsonResponse struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
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
	}
}

func parseJsonResponse(res *http.Response, locationArea *jsonResponse) {
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
		return
	}

	for _, i := range locationArea.Results {
		fmt.Println(i.Name)
	}
}

func main() {
	var locationArea jsonResponse
	pokecache.NewCache(5 * time.Minute) //test line //TO BE MOVED
	for {
		fmt.Println("pokedex >")

		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()

		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}

		m_com := commands()

		switch scanner.Text() {
		case "help":
			fmt.Println("\nHow to use the Pokedex:\n\n")
			fmt.Println("help: ", m_com["help"].description)
			fmt.Println("map:  ", m_com["map"].description)
			fmt.Println("mapb: ", m_com["mapb"].description)
			fmt.Println("exit: ", m_com["exit"].description, "\n")
		case "exit":
			return
		case "map":
			if locationArea.Next == "" {
				res, err := http.Get("https://pokeapi.co/api/v2/location-area/")
				if err != nil {
					log.Fatal(err)
				}
				parseJsonResponse(res, &locationArea)
			} else {
				res, err := http.Get(locationArea.Next)
				if err != nil {
					log.Fatal(err)
				}
				parseJsonResponse(res, &locationArea)
			}
		case "mapb":
			if locationArea.Previous == "" {
				fmt.Println("There were no previous locations")
			} else {
				res, err := http.Get(locationArea.Previous)
				if err != nil {
					log.Fatal(err)
				}
				parseJsonResponse(res, &locationArea)
			}
		default:
			fmt.Println("Type <help> for information on usage")
		}
		fmt.Println("Url for next: ", locationArea.Next)
		fmt.Println("Url for previous: ", locationArea.Previous)
	}
}
