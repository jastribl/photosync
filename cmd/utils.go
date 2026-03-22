package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
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
