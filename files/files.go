package files

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
)

func GetAllFileNamesInDir(
	rootDir string,
	folderDenyRegexs, folderAllowRegexs []*regexp.Regexp,
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
				if !StrMatchesAnyAndNotAny(file.Name(), folderDenyRegexs, folderAllowRegexs) {
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

func GetAllFileNamesInDirAsMap(
	rootDir string,
	folderDenyRegexs, folderAllowRegexs []*regexp.Regexp,
	fileNamesToIgnoreMap map[string]bool,
) (map[string]int, error) {
	allDriveFileNamesArr, err := GetAllFileNamesInDir(
		rootDir,
		folderDenyRegexs,
		folderAllowRegexs,
		fileNamesToIgnoreMap,
	)
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

func StrMatchesAnyAndNotAny(s string, denyRegexs, allowRegexs []*regexp.Regexp) bool {
	for _, denyRegex := range denyRegexs {
		if denyRegex.MatchString(s) {
			denyRuling := true
			for _, allowRegex := range allowRegexs {
				if allowRegex.MatchString(s) {
					denyRuling = false
					break
				}
			}
			if denyRuling {
				return true
			}
		}
	}
	return false
}

func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
