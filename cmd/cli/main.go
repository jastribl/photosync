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

	resp, err := client.Get("https://photoslibrary.googleapis.com/v1/mediaItems")
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Print(string(body))
}
