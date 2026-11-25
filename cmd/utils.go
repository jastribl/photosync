package utils

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jastribl/photosync/photos"
)

func YesNo(q string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {

		fmt.Printf("%s[Y]es|[N]o?", q)
		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

// GetPickerSessionID returns the ID of a Picker session to use. It reuses
// the previously saved session when one exists and the user wants to, and
// otherwise opens a new session, waits for the user to pick photos in the
// browser, and saves its ID to savedSessionCacheFile for next time.
func GetPickerSessionID(client *photos.Client, savedSessionCacheFile string) (string, error) {
	_, err := os.Stat(savedSessionCacheFile)
	cacheDoesNotExist := errors.Is(err, os.ErrNotExist)
	if cacheDoesNotExist || !YesNo("Would you like to re-use your previous picker session?") {
		pickerSession, err := client.NewPickerSession()
		if err != nil {
			return "", err
		}

		pickerSession, err = client.WaitForSession(pickerSession)
		if err != nil {
			return "", err
		}

		pickerSessionID := pickerSession.ID
		f, err := os.Create(savedSessionCacheFile)
		if err != nil {
			return "", err
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		if _, err := w.WriteString(pickerSessionID); err != nil {
			return "", err
		}
		if err := w.Flush(); err != nil {
			return "", err
		}
		return pickerSessionID, nil
	}

	// here we know we failed the top condition because we wanted to use the cache
	out, err := os.ReadFile(savedSessionCacheFile)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
