package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jastribl/photosync/config"
	"github.com/jastribl/photosync/photos"
	"golang.org/x/oauth2"
)

var (
	oauthConfig *oauth2.Config
	gCfg        *config.Config
)

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	queryParts, _ := url.ParseQuery(r.URL.RawQuery)

	// Use the authorization code that is pushed to the redirect
	// URL.
	code := queryParts["code"][0]

	// Exchange will do the handshake to retrieve the initial access token.
	tok, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		log.Fatal(err)
	}
	photos.SaveToken(gCfg, tok)

	// show succes page
	msg := "<p><strong>Success!</strong></p>"
	msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
	fmt.Fprintf(w, msg)

	gCfg.TokenDoneSignal <- true
}

func main() {
	// Setup logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Setup configs
	gCfg = config.NewConfig()
	oauthConfig = photos.GetAuthConfig(gCfg)

	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)

	var err error
	var client *photos.Client
	if !photos.HasToken(gCfg) {
		err := exec.Command("open", url).Start()
		if err != nil {
			log.Fatal(err)
		}

		http.HandleFunc("/oauth/callback", callbackHandler)
		go func() {
			log.Fatal(http.ListenAndServe("localhost:8080", nil))
		}()

		done := <-gCfg.TokenDoneSignal
		if !done {
			log.Fatal("Error in getting token")
		}
	}

	client, err = photos.NewClientForUser(gCfg)
	if err != nil {
		log.Fatal(err)
	}

	rootPicturesDir := gCfg.RootPicturesDir
	allFileNames, err := getAllFileNamesInDir(rootPicturesDir, gCfg.PicturePathSubstringsToIgnore)
	if err != nil {
		log.Fatal(err)
	}

	mediaItmes, err := client.GetAllMediaItems()
	if err != nil {
		log.Fatal(err)
	}

	freeBefore, _ := time.Parse("2006-01-02", gCfg.FreeBeforeDate)

	for _, mediaItem := range mediaItmes {
		filename := mediaItem.Filename
		_, found := allFileNames[filename]
		if !found {
			timeOfImage, err := time.Parse(
				"2006-01-02T15:04:05Z",
				mediaItem.MediaMetadata.CreationTime,
			)
			if err != nil {
				log.Fatal(err)
			}
			if timeOfImage.After(freeBefore) {
				log.Println(mediaItem.ProductULR)
			}
		}
	}
}

func getAllFileNamesInDir(rootDir string, folderSubstringsToIgnore []string) (
	map[string]bool,
	error,
) {
	toReturn := make(map[string]bool)

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
				if !strContainsAnySubstrings(file.Name(), folderSubstringsToIgnore) {
					queue = append(queue, nextItem+file.Name()+"/")
				}
			} else if _, found := gCfg.FileNamesToIgnoreMap[file.Name()]; found {
				continue
			} else if file.Name() == ".DS_Store" {
				log.Fatal("wat")
			} else if _, hasKey := toReturn[file.Name()]; !hasKey {
				toReturn[file.Name()] = true
			}
		}
	}

	return toReturn, nil
}

func strContainsAnySubstrings(s string, anySubstrings []string) bool {
	s = strings.ToLower(s)
	for _, test := range anySubstrings {
		if strings.Contains(s, test) {
			return true
		}
	}
	return false
}
