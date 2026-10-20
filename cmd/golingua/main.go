package main

import (
	"os"

	"github.com/clickbytes/golingua/internal/cli"
)

func main() {
	os.Exit(cli.Run(os.Args[1:], os.Stdout, os.Stderr))
}
