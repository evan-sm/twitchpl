package main

import (
	"fmt"
	"github.com/wmw9/get-twitch-m3u8"
)

func main() {
	hls, err := hls.Get("asmongold")
	if err != nil {
		panic(err)
	}
	fmt.Println(hls)
}
