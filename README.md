# twitchpl - Get direct Twitch m3u8 playlist ▶️🎵
 > I know you were looking for a small lib to do that "dirty" job for you

# Installation 🔨
```console
go install github.com/wmw9/twitchpl/cmd/twitchpl@latest
```

# Features
- 🚀  Choose stream quality: best, worst or audio_only

# Usage: CLI 🔬
# 🔬 Basic usage 
```console
wmw@ubuntu:~$ twitchpl asmongold
{
        "channel": "asmongold",
        "quality": "best",
        "resolution": "1080p",
        "url": "https://video-weaver.fra05.hls.ttvnw.net/v1/playlist/Cow...aqw.m3u8"
}
wmw@ubuntu:~$
```

# Use in your project 🔬
```golang
package main

import (
	"fmt"
	"os"

	"github.com/wmw9/twitchpl"
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

# What I Learned 🧠
- m3u8 parsing
- GraphQL requests
- Golangs basics (HTTP requests, nested structs, error handling)

# ToDo
- [ ] Detect if channel doesn't exists
- [ ] Ignore restreams

# License 📑
(c) 2021 Ivan Smyshlyaev. [MIT License](https://tldrlegal.com/license/mit-license)
