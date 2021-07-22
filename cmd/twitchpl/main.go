package main

import (
	"fmt"
	"os"

	"github.com/wmw9/twitchpl"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Pass twitch channel as an argument\nExample: twitchpl asmongold")
		os.Exit(3)
	}

	pl, err := twitchpl.Get(args[0])
	if err != nil {
		panic(err)
	}

	url, err := pl.Best()
	if err != nil {
		panic(err)
	}

	fmt.Println(url)
}
