package main

import (
	"fmt"
	"log"

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
	fmt.Printf("token=%s\n", token)
}
