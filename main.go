package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type cliCommand struct {
	name        string
	description string
}

func commands(command string) map[string]cliCommand {
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
			description: "",
		},
		"mapb": {
			name:        "mapb",
			description: "",
		},
	}
}

func main() {
	for {
		fmt.Println("pokedex >")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
		}
		m_com := commands(scanner.Text())

		switch scanner.Text() {
		case "help":
			fmt.Println("\nHow to use the Pokedex:\n\nhelp: ", m_com[scanner.Text()].description)
			fmt.Println("exit: ", m_com["exit"].description, "\n")
		case "exit":
			return
		case "map":
			res, err := http.Get("https://pokeapi.co/api/v2/location/")
			if err != nil {
				log.Fatal(err)
			}
			body, err := io.ReadAll(res.Body)
			res.Body.Close()
			if res.StatusCode > 299 {
				log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
			}
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%s", body)
		case "mapb":

		default:
			fmt.Println("Type <help> for information on usage")
		}

	}
}
