package main

import (
	"fmt"

	"github.com/ash_emmett/filesync/client" //must change based on file directory
)

// main function calls clientside program
func main() {

	fmt.Println("client starting up..")
	client.Client()
}
