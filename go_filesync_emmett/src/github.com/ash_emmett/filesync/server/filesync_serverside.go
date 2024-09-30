// -----------------------------------------------------------------------------------
// file sync server application - intern project hyprfire - ash
// -----------------------------------------------------------------------------------

// please look at README.md

package server

// imports
import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
)

const (
	// SERVER_HOST = "172.30.199.56"
	SERVER_HOST = "0.0.0.0" // listen on any address
	SERVER_PORT = "49784"
	SERVER_TYPE = "tcp"
)

var directory = "/deadfs/datashare/template/iso/"

// yaml struct
type VMConfig struct {
	Create Create `yaml:"create"`
}

type Create struct {
	QM QM `yaml:"qm"`
}

type QM struct {
	ID      string `'yaml:id'`
	Name    string `yaml:"name"`
	Memory  string `yaml:"memory"`
	Cores   string `yaml:"cores"`
	Sockets string `yaml:"sockets"`
	SCSIHW  string `yaml:"scsihw"`
	SCSI0   string `yaml:"scsi0"`
	Net0    Net0   `yaml:"net0"`
}

type Net0 struct {
	Model  string `yaml:"model"`
	Bridge string `yaml:"bridge"`
}

var config VMConfig

func Server() {

	// menu
	server, err := net.Listen(SERVER_TYPE, SERVER_HOST+":"+SERVER_PORT)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	defer server.Close()
	fmt.Println("listening on " + SERVER_HOST + ":" + SERVER_PORT + "...")
	fmt.Println("------------------------------------------------")
	fmt.Println("File Sync Serverside")
	fmt.Println("------------------------------------------------")
	fmt.Println("This program will recieve files from the client application and \ndetect '.iso' files to spin up a VM using proxmox. Press CTRL C to exit.")
	fmt.Println("------------------------------------------------")
	fmt.Println("waiting for client...")
	for {
		connection, err := server.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// start main receiving function
		go receiveFile(connection)
		// start looking for ISO files
		go lookForISO(directory, make(chan bool))

	}
}

func receiveFile(connection net.Conn) {
	// read file name size
	fileNameSize := make([]byte, 2)
	_, err := io.ReadFull(connection, fileNameSize)
	if err != nil {
		log.Println("Error reading file name size:", err)
		return
	}

	// allocate buffer for file name
	fileName := make([]byte, int(binary.LittleEndian.Uint16(fileNameSize)))
	_, err = io.ReadFull(connection, fileName)
	if err != nil {
		log.Println("Error reading file name:", err)
		return
	}

	// full path for saving the file
	fullPath := "/deadfs/datashare/template/iso/" + string(fileName) // change this for your own directory

	// create the directory to store files
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}

	// create a new file with the given name
	file, err := os.Create(fullPath)
	if err != nil {
		log.Fatalf("Error creating file: %v", err)
	}
	defer file.Close()

	// read file size
	var fileSize int64
	err = binary.Read(connection, binary.LittleEndian, &fileSize)
	if err != nil {
		log.Println("Error reading file size:", err)
		return
	}

	// create a progress bar
	bar := progressbar.DefaultBytes(fileSize, "Receiving file...")

	// allocate buffer for file content
	content := make([]byte, fileSize)
	_, err = io.ReadFull(connection, content)
	if err != nil && err != io.EOF {
		log.Println("Error reading file content:", err)
		return
	}

	// write file content
	err = os.WriteFile(fullPath, content, 0644)
	if err != nil {
		log.Println("Error writing file content:", err)
		return
	}

	// update progress bar
	bar.Add(1)

	// successfully received file
	fmt.Println("File received and saved successfully.")
	fmt.Println("File Received: " + fullPath)
	// close connection
	bar.Finish()
	connection.Close()
}

// watcher function
func lookForISO(directory string, done chan bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op&fsnotify.Write == fsnotify.Write && event.Name != "" {
					// fmt.Printf("Modified file: %s\n", event.Name)
					if strings.HasSuffix(event.Name, ".iso") {
						//	removePath := strings.TrimPrefix(event.Name, "deadfs/datashare/template/iso")
						fmt.Println("\nDetected .iso file:", event.Name)

						// read into yaml file
						yamlFile, err := os.ReadFile("vmconfig.yaml")
						if err != nil {
							fmt.Printf("Error reading file: %v\n", err)
							return
						}
						err = yaml.Unmarshal(yamlFile, &config)
						if err != nil {
							fmt.Printf("Error parsing YAML: %v\n", err)
							return
						}

						// create vm based on specs in yaml file/detected .iso file
						cmd := exec.Command("qm", "create", config.Create.QM.ID, "--name", config.Create.QM.Name, "--memory", config.Create.QM.Memory, "--cores", config.Create.QM.Cores, "--sockets", config.Create.QM.Sockets, "--scsihw", config.Create.QM.SCSIHW, "--scsi0", config.Create.QM.SCSI0, "-net0", "virtio,bridge=vmbr0", "--ide2", ""+event.Name+",media=cdrom")
						output, err := cmd.CombinedOutput()
						if err != nil {
							log.Printf("Error creating VM: %v\nOutput: %s", err, output)
						} else {
							fmt.Println("VM creation command executed successfully.")
						}
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(directory)
	if err != nil {
		log.Fatal(err)
	}

	<-done // wait indefinitely for a signal to stop the watcher
}

/* original proxmox vm creation command:
- "qm create 5000 --name "AshTestBox" --memory 8192 --cores 8 --sockets 1 --scsihw virtio-scsi-pci --scsi0 DeadFS:50 -net0 virtio,bridge=vmbr0 --ide2 DeadFS:iso/firebug-2024.02.13-x86_64.iso,media=cdrom"
*/
