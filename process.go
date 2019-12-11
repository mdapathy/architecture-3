package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var bufSize = flag.Int("buffer size", 64, "Size of the buffer to read file to ")
var w sync.WaitGroup

func processCLParameters() (string, string) {

	if len(os.Args) < 3 {
		log.Fatalf("Not enough parameters specified, required 2, got %d", len(os.Args)-1)
	} else if dir, err := os.Stat(os.Args[2]); !os.IsNotExist(err) && !dir.IsDir() {
		log.Fatalf("Second param is supposed to be a directory " + os.Args[2])
	} else if len(os.Args) > 3 {
		fmt.Println("Ignoring all parameters except " + os.Args[1] + " and " + os.Args[2])
	}

	return os.Args[1], os.Args[2]

}

func getInputFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

func getLastSentence(inputFile, outputDir string) {

	defer w.Done()

	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Error in the go routine : %s", err)

	}

	var sentence string

	re, _ := regexp.Compile(`(?:\.{3}|!|\.|\?)\s+`)

	stat, _ := file.Stat()

	if int(stat.Size()) > 2000 {
		pos := int(stat.Size() - 2000)
		_, _ = file.Seek(int64(pos), 0)

	}

	for {
		buffer := make([]byte, *bufSize)
		_, err = file.Read(buffer)

		if err == io.EOF {
			break
		}

		sentence += string(buffer)
		match := re.FindAllIndex([]byte(sentence), -1)

		if len(match) >= 1 {
			nonempty, _ := regexp.Match(`[^\s+|\x00]`, []byte(sentence[match[len(match)-1][1]:]))

			if !nonempty && len(match) > 1 {
				sentence = sentence[match[len(match)-2][1]:match[len(match)-1][1]]
			} else if !nonempty {
				sentence = sentence[0:match[len(match)-1][1]]
			} else {
				sentence = sentence[match[len(match)-1][1]:]
			}
		}

	}

	if err = file.Close(); err != nil {
		panic(err)
	}

	res := strings.Split(inputFile, "/")
	outputFileName := strings.Split(res[len(res)-1], ".")[0] + ".res"

	if err = os.MkdirAll(outputDir, 0777); err != nil { //Will create the directory only if it doesn't exist already
		log.Fatalf("Error in the go routine : %s", err)
	}

	outfile, err := os.Create(outputDir + "/" + outputFileName)

	if err != nil {
		log.Fatalf("Error in the go routine : %s", err)
	}

	if _, err = outfile.WriteString(sentence); err != nil {
		log.Fatalf("Error in the go routine : %s", err)
	}

	if err = outfile.Close(); err != nil {
		log.Fatalf("Error in the go routine : %s", err)
	}

}

func main() {

	inputDir, outputDir := processCLParameters()

	inputFiles, err := getInputFiles(inputDir) //Get all files, including the ones in subdirectories
	if err != nil {
		log.Fatalf("An error occurred while processing the input directory: %s", err)
	}

	for _, file := range inputFiles {
		w.Add(1)
		go getLastSentence(file, outputDir)
	}

	w.Wait()

	fmt.Println("Total number of processed files:", len(inputFiles))

}
