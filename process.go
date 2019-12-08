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
)

var bufSize = flag.Int("buffer size", 256, "Size of the buffer to read file to ")

func processCLParameters() (string, string) {

	if len(os.Args) < 3 {
		log.Fatalf("Not enough parameters specified, required 2, got %d", len(os.Args)-1)
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

func getLastSentence(inputFile, outputDir string, channel chan string) {

	file, err := os.Open(inputFile)
	if err != nil {
		channel <- err.Error()
		return
	}

	var sentence strings.Builder
	var fl bool //Indicates whether there is a finished sentence in 'sentence'
	re, _ := regexp.Compile(`(?:\.{3}|!|\.|\?)\s+`)

	for {
		buffer := make([]byte, *bufSize)
		_, err = file.Read(buffer)

		if err == io.EOF {
			break
		}

		match := re.FindAllIndex(buffer, -1)

		if len(match) < 1 {
			nonempty, _ := regexp.Match(`[^$|\s+|\x00]`, buffer[:]) //Indicates whether the last part of buffer is empty

			if nonempty && !fl {
				sentence.Write(buffer)
			} else if nonempty && fl { //Will skip if we have an empty line
				sentence.Reset()
				sentence.Write(buffer)
				fl = false
			}

		} else if len(match) == 1 {
			nonempty, _ := regexp.Match(`[^$|\s+|\x00]`, buffer[match[len(match)-1][1]:])

			if !nonempty {
				sentence.Write(buffer[0:match[len(match)-1][1]])
				fl = true
			} else if len(buffer[match[len(match)-1][1]:]) > 0 {
				sentence.Reset()
				sentence.Write(buffer[match[len(match)-1][1]:*bufSize])
				fl = false
			}

		} else {
			nonempty, _ := regexp.Match(`[^$|\s+|\x00]`, buffer[match[len(match)-1][1]:])

			sentence.Reset()

			if nonempty {
				sentence.Write(buffer[match[len(match)-1][1]:*bufSize])
				fl = false
			} else {
				sentence.Write(buffer[match[len(match)-2][1]:match[len(match)-1][1]])
				fl = true
			}
		}

	}

	if err = file.Close(); err != nil {
		panic(err)
	}

	res := strings.Split(inputFile, "/")
	outputFileName := strings.Split(res[len(res)-1], ".")[0] + ".res"

	if err = os.MkdirAll(outputDir, 0777); err != nil { //Will create the directory only if it doesn't exist already
		channel <- err.Error()
		return
	}

	outfile, err := os.Create(outputDir + "/" + outputFileName)

	if err != nil {
		channel <- err.Error()
		return
	}

	if _, err = outfile.WriteString(sentence.String()); err != nil {
		channel <- err.Error()
		return
	}

	if err = outfile.Close(); err != nil {
		channel <- err.Error()
		return
	}

	channel <- ""
}

func main() {

	inputDir, outputDir := processCLParameters()

	inputFiles, err := getInputFiles(inputDir) //Get all files, including the ones in subdirectories
	if err != nil {
		log.Fatalf("An error occurred while processing the input directory: %s", err)
	}

	channel := make(chan string)

	for _, file := range inputFiles {
		go getLastSentence(file, outputDir, channel)
	}

	filesAmount := len(inputFiles)
	counter := 0

	for counter < filesAmount {
		if err := <-channel; err != "" {
			log.Fatalf("Error inside the go routine : %s", err)
		} else {
			counter++
		}
	}

	fmt.Println("Total number of processed files:", counter)

}
