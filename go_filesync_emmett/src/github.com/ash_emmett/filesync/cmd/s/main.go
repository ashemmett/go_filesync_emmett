package main

import (
	"fmt"

	"github.com/ash_emmett/filesync/server" //must change based on file directory
)

// main function calls serverside program
func main() {

	fmt.Println("server starting up...")
	server.Server()
}
