package utils

import "log"

// TODO: Remove this
func FatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
