package main

// imports
import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// constants and constant variables
const maxThreads = 2 << 6
const copies = 1000

var delim = "-"
var source = "models_template"
var target = "models"

// main function
func main() {

	// copy over the originals
	copyFiles("", "")

	// create a wait group to manage threads
	var wg sync.WaitGroup

	// create a channel to limit concurrent threads
	ch := make(chan struct{}, maxThreads)

	// create the copies
	for i := 0; i < copies; i++ {

		// add to the wait group
		wg.Add(1)

		go func(id int) {
			// remove from the wait group
			defer wg.Done()

			// add to the channel
			ch <- struct{}{}

			log.Println("task id:", id)

			// copy the files
			err := copyFiles(expandId(id), delim)
			if err != nil {
				log.Fatal(err)
			}

			// remove from the channel
			<-ch

		}(i)
	}
	// wait for all threads to finish
	wg.Wait()
}

// helper functions
func copyFiles(id, delim string) (err error) {
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// if it's a file...
		if info.Mode().IsRegular() {

			// get the full path of the file
			src := filepath.Join(path)

			// extract for the destination
			base := filepath.Dir(path)
			ext := filepath.Ext(path)
			name := strings.Replace(filepath.Base(path), ext, "", 1) // without the extension

			// exit if the extension is not .sql
			if ext != ".sql" {
				return nil
			}

			// replace the source directory with the target directory
			base = strings.Replace(base, source, target, 1)

			// construct the name
			name = fmt.Sprintf("%s%s%v%s", name, delim, id, ext)

			// create the destination
			dst := filepath.Join(base, name)

			// copy the file
			err = copyFile(src, dst)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return
}

func expandId(id int) (expanded string) {
	// use number of digits in copies to determine the number of zeros to prepend
	expanded = fmt.Sprintf("%0*d", len(fmt.Sprintf("%d", copies-1)), id)
	return
}

func copyFile(src, dst string) (err error) {
	// open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// create the destination directory
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return err
	}

	// create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// copy the contents
	_, err = io.Copy(dstFile, srcFile)

	return
}
