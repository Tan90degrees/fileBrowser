package main

import (
	"fileBrowser/server"
	"os"
)

func main() {
	server.RunServer()
	os.Exit(0)
}
