package main

import (
	"fmt"

	twpl "github.com/wmw9/get-twitch-m3u8"
)

func main() {
	pl, err := twpl.Get("asmongold")
	if err != nil {
		panic(err)
	}
	fmt.Println(pl)
}
