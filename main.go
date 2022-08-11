package main

import (
	"os"

	"github.com/carlmjohnson/exitcode"
	"github.com/carlmjohnson/loggo/server"
)

func main() {
	exitcode.Exit(server.CLI(os.Args[1:]))
}
