# go-bedrockping [![Go Report Card](https://goreportcard.com/badge/github.com/ZeroErrors/go-bedrockping)](https://goreportcard.com/report/github.com/ZeroErrors/go-bedrockping) [![GoDoc](https://godoc.org/github.com/ZeroErrors/go-bedrockping?status.svg)](https://godoc.org/github.com/ZeroErrors/go-bedrockping) [![Build Status](https://travis-ci.org/ZeroErrors/go-bedrockping.svg?branch=master)](https://travis-ci.org/ZeroErrors/go-bedrockping)
A simple Go library to ping Minecraft Bedrock/MCPE servers.

## Usage
### Installation
Install using ```go get github.com/ZeroErrors/go-bedrockping```

### Example Usage
```golang
package main

import (
	"fmt"
	"github.com/ZeroErrors/go-bedrockping"
	"log"
	"time"
)

func main() {
	resp, err := bedrockping.Query("myip:19132", 5 * time.Second, 150 * time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d/%d players are online.", resp.PlayerCount, resp.MaxPlayers)
}
```

(The default port, 19132, is also available as a const, ```bedrockping.DefaultPort```.)

### Response
The response structure is described in [```bedrockping.Response```](https://github.com/ZeroErrors/go-bedrockping/blob/master/bedrockping.go#L22)
