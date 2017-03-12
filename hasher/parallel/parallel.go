package parallel

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type result struct {
	path   string
	digest [md5.Size]byte
	err    error
}

func Hasher(root string) (map[string][md5.Size]byte, error) {
	digests := make(map[string][md5.Size]byte)
	done := make(chan struct{})
	results := make(chan result)

	defer func() {
		close(done)
	}()

	pathsCh, _ := FileReader(done, root)
	Digester(done, pathsCh, results, 4)
	for {
		select {
		case result, ok := <-results:
			if !ok {
				fmt.Println("All files processed")
				return digests, nil
			}

			if result.err != nil {
				fmt.Println("Error occurred while reading file content", result.err)
				return digests, result.err
			}

			digests[result.path] = result.digest
		}
	}
}

func FileReader(done chan struct{}, root string) (chan string, chan error) {
	pathsCh := make(chan string)
	errCh := make(chan error, 1)
	go func() {
		defer close(pathsCh)
		defer close(errCh)
		errCh <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			select {
			case pathsCh <- path:
			case <-done:
				return errors.New("File reader canceled")
			}

			return nil
		})
	}()

	return pathsCh, errCh
}

func Digester(done chan struct{}, paths chan string, output chan result, maxRoutines int) {
	digest := func() {
		for path := range paths {
			content, err := ioutil.ReadFile(path)
			select {
			case output <- result{path, md5.Sum(content), err}:
			case <-done:
				fmt.Println("Digester stopped")
				return
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(maxRoutines)
	for i := 0; i < maxRoutines; i++ {
		go func() {
			digest()
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()
}
