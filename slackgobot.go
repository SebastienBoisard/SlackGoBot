package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		log.Println("No configuration file loaded - using defaults")
		return
	}

	token := viper.GetString("connection.token")

	var sc slackConnection
	// start a websocket-based Real Time API session
	err = sc.connect(token)
	if err != nil {
		log.Fatalf("Can't connect to Slack [%s]", err)
	}
	fmt.Println("SlackGoBot is running...")
	fmt.Println("SlackGoBot id is =", sc.botID)

	for {

		// Read each incoming message
		msg, err := sc.receiveMessage()
		if err != nil {
			log.Fatal("Error while getting message", err)
		}

		if msg.Type != "message" {
			continue
		}

		// Test if the message was written by the bot
		if msg.User == sc.botID {
			continue
		}

		// The received message is a 'message' type.

		if strings.Contains(msg.Text, "help") {
			fmt.Println("Message received:", msg)
			// NOTE: the Message object is copied, this is intentional
			go func(msg Message) {
				msg.Text = helpUser(msg.Text)
				sc.sendMessage(msg)
			}(msg)
		}
	}
}

func helpUser(text string) string {
	return "Don't be scared, I'm here to help you!"
}
