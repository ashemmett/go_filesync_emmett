// -----------------------------------------------------------------------------------
// file sync client application - intern project hyprfire - ash
// -----------------------------------------------------------------------------------

// please look at README.md

package client

// imports
import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// constants
const (
	SERVER_PORT = "49784"
	SERVER_TYPE = "tcp"
)

// variables
var (
	SERVER_HOST string
)

func Client() {

	// define done
	done := make(chan bool) // creates done channel to signal watcher to stop

	// check if a command-line argument is provided
	if len(os.Args) > 1 {
		directory := os.Args[1]
		fmt.Println("Path:", directory)
		SERVER_HOST = "172.30.199.202"
		fmt.Println("Server IP: 172.30.199.202")
		go watchDirectory(directory, done)
		killProgram(done)
	} else {
		// normal menu instead
		fmt.Println("------------------------------------------------")
		fmt.Println("File Sync Clientside")
		fmt.Println("------------------------------------------------")
		fmt.Println("Select an option:")
		fmt.Println("Press '1' to monitor a directory, '2' for default directory or '3' to exit.")
		fmt.Println("If monitoring a directory, press 'q' to quit or 'r' to restart!")
		fmt.Println("------------------------")
		var option string
		fmt.Scanln(&option)

		switch option {

		// specify directory
		case "1":
			fmt.Println("------------------------")
			directory := getUserDirectoryInput() // call directory input function
			fmt.Println("------------------------")
			fmt.Println("Path:", directory)
			fmt.Println("Enter '1' to enter a server IP or '2' for the default server IP:")
			var choice string
			fmt.Scanln(&choice)
			switch choice {
			// enter ip address
			case "1":
				fmt.Println("------------------------")
				getServerIP()
				fmt.Println("------------------------")
				fmt.Println("Server IP Address:", SERVER_HOST)
				fmt.Println("------------------------")
				go watchDirectory(directory, done)
				killProgram(done)
			// default ip address
			case "2":
				fmt.Println("------------------------")
				fmt.Println("Server IP: 172.30.199.202")
				SERVER_HOST = "172.30.199.202"
				fmt.Println("------------------------")
				go watchDirectory(directory, done)
				killProgram(done)
			// invalid input
			default:
				fmt.Println("------------------------")
				fmt.Println("Invalid selection, please try again.")
				Client() // recursive call
			}
		// exit
		case "3":
			fmt.Println("------------------------")
			fmt.Println("Exiting...")
			fmt.Println("------------------------")
			os.Exit(0)
		// default directory
		case "2":
			fmt.Println("------------------------")
			fmt.Println("Path: /Users/ashtonemmett/Desktop/")
			directory := "/Users/ashtonemmett/Desktop/" // set default directory (desktop)
			fmt.Println("------------------------")
			fmt.Println("Enter '1' to enter a server IP or '2' for the default server IP:")
			var choice string
			fmt.Scanln(&choice)
			switch choice {
			// enter ip address
			case "1":
				fmt.Println("------------------------")
				getServerIP()
				fmt.Println("------------------------")
				fmt.Println("Server IP Address:", SERVER_HOST)
				fmt.Println("------------------------")
				go watchDirectory(directory, done)
				killProgram(done)
			// default ip address
			case "2":
				fmt.Println("------------------------")
				fmt.Println("Server IP: 172.30.199.202")
				SERVER_HOST = "172.30.199.202"
				fmt.Println("------------------------")
				go watchDirectory(directory, done)
				killProgram(done)
			//invalid input
			default:
				fmt.Println("------------------------")
				fmt.Println("Invalid selection, please try again.")
				Client() // recursive call
			}
		// invalid input
		default:
			fmt.Println("------------------------")
			fmt.Println("Invalid selection, please try again.")
			Client() // recursive call
		}
	}
}

// input directory path function
func getUserDirectoryInput() string {
	var directory string // directory variable
	//input line
	fmt.Println("Enter directory path: ")
	fmt.Scan(&directory)
	return directory
}

// input server IP function
func getServerIP() string {
	//input line
	fmt.Println("Enter a Server IP: ")
	fmt.Scan(&SERVER_HOST)
	return SERVER_HOST
}

// watcher function (calls directory input)
func watchDirectory(directory string, done chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				timestamp := time.Now().Format(time.RFC3339)     // make timestamp format
				if !strings.HasSuffix(event.Name, ".DS_Store") { // ignore "".DS_Store" files
					log.Printf("[%s] event: %v", timestamp, event)
					if event.Op&fsnotify.Write == fsnotify.Write { // check for modified files (changes)
						log.Printf("[%s] modified file: %v", timestamp, event.Name) // logs the detected modified file
						sendFile(event.Name, SERVER_HOST, 49784)
					}
				}
				if event.Op&fsnotify.Create == fsnotify.Create { // check for new files (added)
					log.Printf("[%s] new file: %v", timestamp, event.Name) // logs the detected new file
					sendFile(event.Name, SERVER_HOST, 49784)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("[%s] error: %v", time.Now().Format(time.RFC3339), err) // logs errors in the watcher
			}
		}
	}()

	err = watcher.Add(directory) // adds inputted directory to the watcher
	if err != nil {
		log.Fatal(err)
	}

	<-done // waits indefintely for a signal in the done channel to stop the watcher
}

// function to quit and restart program
func killProgram(done chan bool) {
	input := ""
	fmt.Scanln(&input)
	if input == "q" {
		close(done) // signal to stop watching
		fmt.Println("Exiting, thank you.")
		os.Exit(0) // terminate the program
	}
	if input == "r" {
		close(done) // signal to stop watching
		fmt.Println("Restarting, thank you.")
		Client() // recursive call to restart program
	}
}

// function to send file from client to server
func sendFile(filePath string, host string, port int) {
	// dial TCP connection
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// open file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// get file name
	fileName := filepath.Base(filePath)
	// fmt.Println("Sending file name:", fileName)

	// specify and send file name size
	fileNameSize := len(fileName)
	fileNameSizeBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(fileNameSizeBytes, uint16(fileNameSize))
	_, err = conn.Write(fileNameSizeBytes)
	if err != nil {
		log.Fatal(err)
	}

	// send file name
	_, err = conn.Write([]byte(fileName))
	if err != nil {
		log.Fatal(err)
	}

	// move file pointer to the beginning of the file
	_, err = file.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	// specify file size
	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	fileSize := fileInfo.Size()

	// send file size
	fileSizeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(fileSizeBytes[:], uint64(fileSize))
	_, err = conn.Write(fileSizeBytes)
	if err != nil {
		log.Fatal(err)
	}

	// send file contents
	buffer := make([]byte, fileSize)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		log.Fatal(err)
	}
	_, err = conn.Write(buffer)
	if err != nil {
		log.Fatal(err)
	}
}
