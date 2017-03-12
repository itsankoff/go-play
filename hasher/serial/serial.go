package serial

import (
	"crypto/md5"
	"io/ioutil"
	"os"
	"path/filepath"
)

func Hasher(root string) (map[string][md5.Size]byte, error) {
	files := make(map[string][md5.Size]byte)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.Mode().IsRegular() {
			return nil
		}

		content, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		files[path] = md5.Sum(content)
		return nil
	})

	if err != nil {
		return map[string][md5.Size]byte{}, err
	}

	return files, nil
}
