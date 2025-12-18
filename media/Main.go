package main

import (
	"log"

	"syscall"
)

func main() {
	dll, err := syscall.LoadDLL("scan.dll")
	if err != nil {
		log.Fatal(err)
	}
	dll.FindProc("")

}
