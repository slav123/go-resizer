package storage

import (
	"io/ioutil"
	"os"
)

func SaveFile(buf []byte, filename string) {
	// open output file
	fo, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// write a chunk
	if _, err := fo.Write(buf); err != nil {
		panic(err)
	}
}

func ReadFile(filename string) ([]byte, error) {
	f, err := os.Open(filename)

	if err != nil {
		return make([]byte, 0), err
	} else {
		inBuf, _ := ioutil.ReadAll(f)
		return inBuf, nil
	}
}

func CheckCacheFs(hash,cacheDir string) bool {
	if _, err := os.Stat(cacheDir + hash); err == nil {
		return true
	} else {
		return false
	}
}