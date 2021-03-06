package photos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	urlApi "net/url"
	"os"
	"os/exec"

	"github.com/jastribl/photosync/config"
	"golang.org/x/oauth2"
)

// Client holds all things for Photos requests
type Client struct {
	httpClient *http.Client
}

// HasToken returns if the user has a token
func HasToken(cfg *config.Config) bool {
	_, err := tokenFromFile(cfg)
	return err == nil
}

// Retrieves a token from a local file.
func tokenFromFile(cfg *config.Config) (*oauth2.Token, error) {
	f, err := os.Open(cfg.TokenFileLocation)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// SaveToken saves a token given a config
func SaveToken(cfg *config.Config, token *oauth2.Token) error {
	f, err := os.OpenFile(cfg.TokenFileLocation, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}

// GetAuthConfig returns a new auth config
func GetAuthConfig(cfg *config.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Scopes:       cfg.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  cfg.AuthURL,
			TokenURL: cfg.TokenURL,
		},
		RedirectURL: cfg.RedirectURL,
	}
}

// NewClientForUser gets a new client for a user using the user token
func NewClientForUser(cfg *config.Config) (*Client, error) {
	oauthConfig := GetAuthConfig(cfg)

	url := oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)

	if !HasToken(cfg) {
		err := exec.Command("open", url).Start()
		if err != nil {
			log.Fatal(err)
		}

		http.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
			queryParts, _ := urlApi.ParseQuery(r.URL.RawQuery)

			// Use the authorization code that is pushed to the redirect
			// URL.
			code := queryParts["code"][0]

			// Exchange will do the handshake to retrieve the initial access token.
			tok, err := oauthConfig.Exchange(context.Background(), code)
			if err != nil {
				log.Fatal(err)
			}
			SaveToken(cfg, tok)

			// show succes page
			msg := "<p><strong>Success!</strong></p>"
			msg = msg + "<p>You are authenticated and can now return to the CLI.</p>"
			fmt.Fprint(w, msg)

			cfg.TokenDoneSignal <- true
		})
		go func() {
			log.Fatal(http.ListenAndServe("localhost:8080", nil))
		}()

		done := <-cfg.TokenDoneSignal
		if !done {
			log.Fatal("Error in getting token")
		}
	}

	tok, err := tokenFromFile(cfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: GetAuthConfig(cfg).Client(context.Background(), tok),
	}, nil
}

func (m *Client) postJson(url string, requestBody interface{}, responseObj interface{}) error {
	jsonStr, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}
	resp, err := m.httpClient.Post(
		url,
		"application/json",
		bytes.NewBuffer(jsonStr),
	)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&responseObj)
}
