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

// SlackBot contains all the data to run a bot on Slack
type SlackBot struct {
	ws *websocket.Conn
	id string
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

// NewBot starts a websocket-based Real Time API session and return the websocket
// and the ID of the (bot-)user whom the token belongs to.
func NewBot(token string) (*SlackBot, error) {

	websocketURL, botID, err := startSlack(token)
	if err != nil {
		log.Printf("Error with startSlack")
		return nil, err
	}

	ws, err := websocket.Dial(websocketURL, "", "https://api.slack.com/")
	if err != nil {
		log.Printf("error with websocket.Dial")
		return nil, err
	}

	return &SlackBot{ws: ws, id: botID}, nil
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

func (bot *SlackBot) receiveMessage() (Message, error) {
	var msg Message
	err := websocket.JSON.Receive(bot.ws, &msg)
	return msg, err
}

var counter uint64

// sendSlackMessage sends a message to Slack by sending JSON over the websocket connection.
func (bot *SlackBot) sendMessage(msg Message) error {
	msg.ID = atomic.AddUint64(&counter, 1)
	err := websocket.JSON.Send(bot.ws, msg)
	return err
}
