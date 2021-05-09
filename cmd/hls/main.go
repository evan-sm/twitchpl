package main

import (
	"flag"
	"fmt"
	hls "github.com/wmw9/get-twitch-m3u8"
	"os"
)

const VERSION = "0.2"

func main() {
	mpl := flag.Bool("m", false, "Show master playlist instead")
	flag.Parse()

	if flag.NArg() < 1 {
		os.Stderr.Write([]byte(fmt.Sprintf("get-twitch-hls %v - Gets the m3u8 HTTP Live Streaming (HLS) direct URL of a live stream on twitch.tv\n", VERSION)))
		os.Stderr.Write([]byte("Copyright (C) 2021 Ivan Smyshlaev.\n"))
		os.Stderr.Write([]byte("Usage: hls sodapoppin\n"))
		flag.PrintDefaults()
		os.Exit(2)
	} else if *mpl {
		hls, err := hls.GetMPL(flag.Arg(0))
		if err != nil {
			panic(err)
		}
		fmt.Println(hls)
	} else {
		hls, err := hls.Get(flag.Arg(0))
		if err != nil {
			panic(err)
		}
		fmt.Println(hls)
	}
}
