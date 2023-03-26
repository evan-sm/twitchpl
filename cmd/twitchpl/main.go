package main

import (
	"context"
	"os"

	"github.com/wmw64/twitchpl"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		println("Pass twitch channel as an argument. Example: twitchpl asmongold")
		os.Exit(3)
	}

	pl, err := twitchpl.Get(context.Background(), args[0], true)
	if err != nil {
		panic(err)
	}

	// println(pl.AsJSON()) // Best quality by default
	// {
	// 	"channel": "asmongold",
	// 	"quality": "best",
	// 	"resolution": "1920x1080",
	// 	"frame_rate": 60,
	// 	"url": "https://video-weaver.arn04.hls.ttvnw.net/v1/playlist/C..JIG.m3u8"
	// }

	println(pl.AsURL())
	// https://video-weaver.arn04.hls.ttvnw.net/v1/playlist/C..JIG.m3u8

	//	println(pl.Worst().AsJSON())
	//	println(pl.Best().AsURL())
}
