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
	pl, err := twitchpl.Get("asmongold")
	if err != nil {
		panic(err)
	}

	fmt.Println(pl.AsJSON()) // Best quality by default
//	fmt.Println(pl.Worst().AsJSON())
//	fmt.Println(pl.Best().AsURL())
//	fmt.Println(pl.AsText())

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

