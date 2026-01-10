package photos

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"
)

func (m *Client) NewPickerSession() (*PickerSession, error) {
	request := PickerSessionRequest{}
	resp := &PickerSession{}

	err := m.postJson("https://photospicker.googleapis.com/v1/sessions", &request, &resp)
	if err != nil {
		return nil, err
	}

	resp.PollingURI = fmt.Sprintf("https://photospicker.googleapis.com/v1/sessions/%s", resp.ID)
	return resp, err
}

func (m *Client) GetSessionInfo(sessionID string) (*PickerSession, error) {
	resp, err := m.httpClient.Get(fmt.Sprintf(
		"https://photospicker.googleapis.com/v1/sessions/%s",
		sessionID,
	))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	d := &PickerSession{}
	err = json.NewDecoder(resp.Body).Decode(d)

	return d, err
}

func (m *Client) WaitForSession(session *PickerSession) (*PickerSession, error) {
	err := exec.Command("open", session.PickerURI).Start()
	if err != nil {
		return nil, err
	}

	sessionInfo := &PickerSession{}
	for true {
		log.Print("Checking session state...")
		sessionInfo, err = m.GetSessionInfo(session.ID)
		if err != nil {
			return nil, err
		}
		if sessionInfo.MediaItemsSet {
			break
		} else {
			log.Print("Media items not set yet, sleeping 5s...")
			time.Sleep(5 * time.Second)
		}
	}

	log.Print("Session waiting complete!")

	return sessionInfo, nil
}
