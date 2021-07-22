# twitchpl - Get direct Twitch m3u8 playlist ▶️🎵
 > I know you were looking for a small lib to do that "dirty" job for you

# Installation 🔨
```go install github.com/wmw9/twitchpl/cmd/twitchpl@latest``` <br>

# Features
- ✅  Avoid Ads
- 🚀  Choose stream quality: best, worst or audio_only

# Usage: CLI 🔬
```twitchpl asmongold``` <br>

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

	url, err := pl.Best()
	if err != nil {
		panic(err)
	}

	fmt.Println("Best m3u8:", url)

    pl, err = twitchpl.Get(flag.Arg(0)).Quality("best")
    if err != nil {
        panic(err)
    }
	fmt.Println("Best m3u8:", url)

	mpl, err = twitchpl.GetMPL(args[0])
	if err != nil {
		panic(err)
	}

	fmt.Println("Master playlist:", mpl)
}

```

# What I Learned 🧠
- m3u8 parsing
- GraphQL requests
- Golangs basics (OOP, HTTP requests, nested structs, error handling)

# ToDo
- [ ] Detect if channel doesn't exists
- [ ] Ignore restreams

# License 📑
(c) 2021 Ivan Smyshlyaev. [MIT License](https://tldrlegal.com/license/mit-license)
