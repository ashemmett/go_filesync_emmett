#!/bin/bash

# ***************************************
# cross compile and transfer script - ash
# ***************************************
# simple tool for compiling Golang projects from MACOS to Linux and then sends the built Go project to the dev box.

# set the linux operating system and architecture
export GOOS=linux
export GOARCH=amd64

# build the Golang application for the linux OS/architecture
go build -o server

# transfer to dev box 
scp server DeadRange:~/

# print confirmation
echo "cross-compilation completed and file sent to dev box successfully."