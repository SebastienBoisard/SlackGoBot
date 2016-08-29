package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync/atomic"

	"golang.org/x/net/websocket"
)

type slackConnection struct {
	ws    *websocket.Conn
	botID string
}

type responseRtmStart struct {
	Ok    bool         `json:"ok"`
	URL   string       `json:"url"`
	Error string       `json:"error"`
	Self  responseSelf `json:"self"`
}

type responseSelf struct {
	ID string `json:"id"`
}

func startSlack(token string) (string, string, error) {

	// To begin a RTM session make an authenticated call to the rtm.start API method
	// Cf. https://api.slack.com/methods/rtm.start
	response, err := http.Get("https://slack.com/api/rtm.start?token=" + token)
	if err != nil {
		log.Printf("Error while getting a websocket from Slack err=%v", err)
		return "", "", err
	}

	if response.StatusCode != 200 {
		err = fmt.Errorf("API request failed with code %d", response.StatusCode)
		return "", "", err
	}

	body, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		log.Printf("Error while getting a websocket from Slack err=%v", err)
		return "", "", err
	}

	var responseObj responseRtmStart
	err = json.Unmarshal(body, &responseObj)
	if err != nil {
		return "", "", err
	}

	if responseObj.Ok == false {
		err = fmt.Errorf("Slack error: %s", responseObj.Error)
		return "", "", err
	}

	return responseObj.URL, responseObj.Self.ID, nil
}

// Starts a websocket-based Real Time API session and return the websocket
// and the ID of the (bot-)user whom the token belongs to.
func (sc *slackConnection) connect(token string) error {
	websocketURL, botID, err := startSlack(token)
	if err != nil {
		log.Printf("Error with startSlack")
		return err
	}

	sc.ws, err = websocket.Dial(websocketURL, "", "https://api.slack.com/")
	if err != nil {
		log.Printf("error with websocket.Dial")
		return err
	}

	sc.botID = botID

	return nil
}

// These are the messages read off and written into the websocket. Since this
// struct serves as both read and write, we include the "Id" field which is
// required only for writing.

// Message is...
type Message struct {
	ID      uint64 `json:"id"`
	Type    string `json:"type"`
	Channel string `json:"channel"`
	User    string `json:"user"`
	Text    string `json:"text"`
}

func (sc *slackConnection) receiveMessage() (Message, error) {
	var msg Message
	err := websocket.JSON.Receive(sc.ws, &msg)
	return msg, err
}

var counter uint64

// sendSlackMessage sends a message to Slack by sending JSON over the websocket connection.
func (sc *slackConnection) sendMessage(msg Message) error {
	msg.ID = atomic.AddUint64(&counter, 1)
	err := websocket.JSON.Send(sc.ws, msg)
	return err
}
