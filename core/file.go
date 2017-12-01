package core

import "os"

func createFileWithDefaultContentsIfNotExists(filename string, defaultContents string) error {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// no-op, filename will be created
	} else if err != nil {
		return err
	} else {
		// filename exists, never overwrite
		return nil
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(defaultContents))
	return err
}
