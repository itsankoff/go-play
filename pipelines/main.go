package main

import (
	"fmt"
	"github.com/itsankoff/go-play/pipelines/serial"
)

func main() {
	files, err := serial.Hasher(".")
	if err != nil {
		fmt.Println(err)
	}

	for path, digest := range files {
		fmt.Printf("%x, %s\n", digest, path)
	}
}
