package main

import (
	"github.com/ksyq12/vhost/internal/cli"
	_ "github.com/ksyq12/vhost/internal/driver" // Register drivers
)

func main() {
	cli.Execute()
}
