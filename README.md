# go-bedrock-ping [![Go Report Card](https://goreportcard.com/badge/github.com/ZeroErrors/go-bedrock-ping)](https://goreportcard.com/report/github.com/ZeroErrors/go-bedrock-ping) [![GoDoc](https://godoc.org/github.com/ZeroErrors/go-bedrock-ping?status.svg)](https://godoc.org/github.com/ZeroErrors/go-bedrock-ping)
A simple Go library to ping Minecraft Bedrock/MCPE servers.

## Usage
### Installation
Install using ```go get github.com/ZeroErrors/go-bedrock-ping```

### Example Usage
```golang
package main

import (
	"fmt"
	"github.com/ZeroErrors/go-bedrock-ping"
	"log"
	"time"
)

func main() {
	resp, err := bedrockping.Query("myip:19132", time.Second * 5)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d/%d players are online.", resp.PlayerCount, resp.MaxPlayers)
}
```

(The default port, 19132, is also available as a const, ```bedrockping.DefaultPort```.)

### Response
The response structure is described in [```bedrockping.Response```](https://github.com/ZeroErrors/go-bedrock-ping/blob/master/bedrock-ping.go#L22)
