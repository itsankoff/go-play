package main

import (
	"fmt"
	"github.com/itsankoff/go-play/pipelines/parallel"
	"github.com/itsankoff/go-play/pipelines/serial"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: pipelines <mode (serial|parallel)> <directory>")
		os.Exit(1)
	}

	mode := os.Args[1]
	dir := os.Args[2]
	fmt.Println("Mode: ", mode)
	if mode == "serial" {
		// 		files, err := serial.Hasher(dir)
		// 		if err != nil {
		// 			fmt.Println(err)
		// 		}
		//
		// 		for path, digest := range files {
		// 			fmt.Printf("%x, %s\n", digest, path)
		// 		}
		serial.Hasher(dir)
	}

	if mode == "parallel" {
		// 		files, err := parallel.Hasher(dir)
		// 		if err != nil {
		// 			fmt.Println(err)
		// 		}
		//
		// 		for path, digest := range files {
		// 			fmt.Printf("%x, %s\n", digest, path)
		// 		}
		parallel.Hasher(dir)
	}
}
