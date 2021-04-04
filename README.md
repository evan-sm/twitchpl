# get-twitch-m3u8

<img align="center" width="200" src="https://user-images.githubusercontent.com/4693125/113519851-8b9f3800-9597-11eb-90c4-ca41be0f848d.png" alt="gopher">

Gets the direct .m3u8 HLS twitch playlist URL

## Usage

Install CLI binary using Go

```bash
go install github.com/wMw9/get-twitch-m3u8/cmd/hls
hls sodapoppin
https://video-weaver.hel01.hls.ttvnw.net/v1/playlist/CocEr.....mggc.m3u8
```

Or use in your project:

```golang
package main

import ()
	"github.com/wmw9/get-twitch-m3u8"
)

func main() {
	hls, err := hls.Get("asmongold")
	if err != nil {
		panic(err)
	}
	println(hls)
}

```
