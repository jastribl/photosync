package files

import (
	"log"
	"os"
	"strings"
)

func GetAllFilenamesInDir(rootDir string) []string {
	filenames := []string{}

	queue := []string{rootDir}
	for len(queue) > 0 {
		nextItem := queue[0]
		queue = queue[1:]

		files, err := os.ReadDir(nextItem)
		if err != nil {
			log.Fatalf("Error reading dir '%s': %s\n", nextItem, err.Error())
		}
		for _, file := range files {
			if file.IsDir() {
				queue = append(queue, nextItem+file.Name()+"/")
			} else if file.Name() == ".DS_Store" {
				continue
			} else {
				filenames = append(filenames, file.Name())
			}
		}
	}

	return filenames
}

func GetAllLowercaseFilenamesInDir(rootDir string) []string {
	lowercaseFilenames := []string{}

	for _, b := range GetAllFilenamesInDir(rootDir) {
		lowercaseFilenames = append(lowercaseFilenames, strings.ToLower(b))
	}

	return lowercaseFilenames
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
