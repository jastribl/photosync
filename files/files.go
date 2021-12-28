package files

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

func GetAllFileNamesInDir(
	rootDir string,
	folderIgnoreRegexs []string,
	fileNamesToIgnoreMap map[string]bool,
) ([]string, error) {
	toReturn := []string{}

	queue := []string{rootDir}
	for len(queue) > 0 {
		nextItem := queue[0]
		queue = queue[1:]

		files, err := ioutil.ReadDir(nextItem)
		if err != nil {
			return nil, err
		}
		for _, file := range files {
			if file.IsDir() {
				if !StrContainsAnySubstrings(file.Name(), folderIgnoreRegexs) {
					queue = append(queue, nextItem+file.Name()+"/")
				} else {
					log.Println("skipping dir: " + file.Name())
				}
			} else if _, found := fileNamesToIgnoreMap[file.Name()]; found {
				continue
			} else if file.Name() == ".DS_Store" {
				log.Println("wat... found .DS_Store")
			} else {
				toReturn = append(toReturn, file.Name())
			}
		}
	}

	return toReturn, nil
}

func GetAllFileNamesInDirAsMap(rootDir string, folderSubstringsToIgnore []string, fileNamesToIgnoreMap map[string]bool) (
	map[string]int,
	error,
) {
	allDriveFileNamesArr, err := GetAllFileNamesInDir(rootDir, folderSubstringsToIgnore, fileNamesToIgnoreMap)
	if err != nil {
		return nil, err
	}
	toReturn := map[string]int{}
	for _, fileName := range allDriveFileNamesArr {
		if _, hasKey := toReturn[fileName]; hasKey {
			log.Printf("Duplicate key?: %s\n", fileName)
			toReturn[fileName] = toReturn[fileName] + 1
		}
		toReturn[fileName] = 1
	}
	return toReturn, nil
}

func StrContainsAnySubstrings(s string, regexs []string) bool {
	for _, regex := range regexs {
		if match, err := regexp.MatchString(regex, s); err == nil && match {
			return true
		}
	}
	return false
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
