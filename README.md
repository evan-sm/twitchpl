# twitchpl - Get direct Twitch m3u8 HLS playlist ▶️🎵
 > Small lib used to extract twitch.tv livestreams HLS playlist for future needs

# 🔨 Installation 
```console
go install github.com/wmw64/twitchpl/cmd/twitchpl@latest
```

# Features
- 🚀  Choose stream quality: best, worst or audio_only

# 🧑‍💻 Usage 
# CLI
```console
wmw@zsh:~$ twitchpl honeymad
{
        "channel": "honeymad",
        "quality": "best",
        "resolution": "1920x1080",
        "frame_rate": 60,
        "url": "https://video-weaver.fra05.hls.ttvnw.net/v1/playlist/CzmA.m3u8"
}
wmw@zsh:~$
```

# In your project 
```golang
package main

import (
	"fmt"
	"os"

	"github.com/wmw64/twitchpl"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		println("Pass twitch channel as an argument. Example: twitchpl asmongold")
		os.Exit(3)
	}

	// how to get token
	// https://streamlink.github.io/cli/plugins/twitch.html
	token := "turbotoken"

	pl, err := twitchpl.Get(context.Background(), args[0], token)
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

```

# 🧠 What I Learned 
- m3u8 parsing
- GraphQL requests
- Golangs basics (HTTP requests, nested structs, error handling)

# 📝 ToDo
- [ ] Detect if channel doesn't exists
- [ ] Ignore restreams

# 👤 Author & License
©️ 2021 Ivan Smyshlyaev. [MIT License](https://tldrlegal.com/license/mit-license)
👉 [My instagram page](https://instagram.com/wmw)

