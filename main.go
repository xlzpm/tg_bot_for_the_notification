package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	configFile string
	SiteList   map[string]int
)

func init() {
	SiteList = make(map[string]int)
	flag.StringVar(&configFile, "config", "config.json", "config file")

}

func save_list() {
	data, err := json.Marshal(SiteList)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(configFile, data, 0644)
	if err != nil {
		panic(err)
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
}
