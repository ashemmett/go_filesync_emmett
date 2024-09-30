---------------------------------------
File Sync | README.md | Hyprfire | Ash
---------------------------------------

## Description
The file sync clientside application sits on one's computer, where it can be run to monitor a specified directory on that computer. It will also ask for a server IP address so files that are updated/modified or created can be sent to the serverside application. The serverside application downloads all of the files recieved from the client side application to a specifed folder (in this case /deadfs/datashare/template/iso/) where a second watcher function monitors for recieved ".iso" files and then attempts to spin up a proxmox vm based on that disc image and the vm settings specified in the vmconfig.yaml file.

## Install
Ensure the proper GO install has been completed "go build" etc, full instructions can be found https://go.dev/doc/install . Any changes to the code have to be saved and then "go build" must be run.
I have used a number of GO extended libraries/packages. Use the command "go get" to import these libraries.

```go
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
	"strings"
	"github.com/fsnotify/fsnotify"
	"github.com/schollz/progressbar/v3"
	"gopkg.in/yaml.v3"
```

## Setup
There is 2 main files the filesync_clientside and filesync_serverside application. Inside the "s" and "c" file, there is main.go function that call/run filesync_clientside and filesync_serverside. 

```yaml
---
create:
  qm:
    id: 5000
    name: VMTestCreation
    memory: 8192
    cores: 8
    sockets: 1
    scsihw: virtio-scsi-pci
    scsi0: DeadFS:50
    net0: 
      model: virtio
      bridge: vmbr0
```

## Extra Files
There is a script called "com.sh", this script peforms the cross-compilation of the serverside GO code (macOS) to (Linux), it also sends the code to the devbox. 
There is also a config file designed to sit in the same directory as the directorymonitor_serverside called vmconfig.yaml - follow the set structure and assign values to the necessary headings to specify the size/specs of proxmox VM. 

## Usage
basic startup: "go run filesync_clientside.go" ~ "./c" 
               "go run filesync_serverside.go" ~ "./s"

Any changes or updates made to the program, the file needs to be saved and then inside the same directory the command "go build" has to be run or in the case of the serverside applcation - please see ./com.sh to cross compile and go build the file for linux and to send to dev box.

Both applications (client + server) have to be running for the program to work. Inside of both programs there is default settings that are hard coded. i.e default server IP, default directory etc. These can be changed at will, I have listed all of the lines that can be changed below.

## Client Application
"./c" when the clientside application is run it will take you to menu prompts. Follow menu prompts to specify a director and server address or specify the directory as a commandline argument "./c /Users/ashtonemmett/Desktop/" (choosing this option will use the hard coded server IP address which is configurable).

```go
filesync_clientside core functions:
Line 39: func Client() // main
Line 148: func getUserDirectoryInput() string
Line 157: func getServerIP() string
Line 165: func watchDirectory(directory string, done chan bool)
Line 210: func killProgram(done chan bool)
Line 226: func sendFile(filePath string, host string, port int)

filesync_clientside (things to config/change)
Line 30:    SERVER_PORT = "49784" // set default serverport
Line 48:    SERVER_HOST = "172.30.199.202" // set default IP
Line 89:	SERVER_HOST = "172.30.199.202" // set default IP
Line 109:   directory := "/Users/ashtonemmett/Desktop/" // set default 
Line 128:   SERVER_HOST = "172.30.199.202" // set default IP
Line 185: 	sendFile(event.Name, SERVER_HOST, 49784) // change serverport
Line 190: 	sendFile(event.Name, SERVER_HOST, 49784) // change serverport
```

## Server Application
"./s" when the serverside application is run, it will run indefinitely listening in the background waiting to connect with the clientside application. Files will get synced across automatically once the clientside is running. If the serverside application detects an ".iso" file - the program will attempt to spin up a VM using proxmox, the desired ".iso" file and the settings in the vmconfig.yaml file.

```go
filesync_serverside core functions:
Line 67: func Server() // main
Line 99: func receiveFile(connection net.Conn)
Line 170: func lookForISO(directory string, done chan bool)

filesync_serverside (things to config/change)
Line 34:    SERVER_HOST = "0.0.0.0" // set default server IP
Line 35:    SERVER_PORT = "49784" // set default server port 
Line 39:	var directory = "/deadfs/datashare/template/iso/" //specify directory to copy files
Line 117:   fullPath := "/deadfs/datashare/template/iso/" + string(fileName) //specify directory in recieve file function
Line 201:   //calls yaml config file - config settings for VM in vmconfig.yaml
```

## The End - Ash
