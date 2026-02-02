package main

import (
	"github.com/ksyq12/vhost/internal/cli"
	_ "github.com/ksyq12/vhost/internal/driver" // Register drivers
)

// version is set by goreleaser via ldflags
var version = "dev"

func main() {
	cli.SetVersion(version)
	cli.Execute()
}
