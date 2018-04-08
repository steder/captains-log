package main

import (
	"fmt"
	"net/http"
	"os"
	"io"
)


const PRINTER_HOST string = "192.168.0.135"


func main() {
	fmt.Printf("Connecting to %s\n", PRINTER_HOST )
	snapshot, err := http.Get(fmt.Sprintf("http://%s:8080/?action=snapshot", PRINTER_HOST))
	if err != nil {
		// handle it
		fmt.Println("Unable to load the snapshot: %s", err)
	}
	defer snapshot.Body.Close()

	destPath := fmt.Sprintf("snap.%d.png", 1)
	destFile, err := os.Create(destPath)
	if err != nil {
		fmt.Printf("Unable to create %s: %s", destPath, err)
	}
	defer destFile.Close()
	os.Chmod(destPath, 0755)

	io.Copy(destFile, snapshot.Body)
}
