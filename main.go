package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	SiteList              map[string]int
	chatID                int64
	sslDaysToExipireAlert int64
	telegramBotToken      string
	configFile            string
	pprofListen           string
	HelpMsg               = "Это простой мониторинг доступности сайтов. Он обходит сайты в списке и ждет что он ответит 200, если возвращается не 200 или ошибки подключения, то бот пришлет уведомления в групповой чат\n" +
		"Список доступных комманд:\n" +
		"/site_list - покажет список сайтов в мониторинге и их статусы (про статусы ниже)\n" +
		"/site_add [url] - добавит url в список мониторинга\n" +
		"/site_del [url] - удалит url из списка мониторинга\n" +
		"/help - отобразить это сообщение\n" +
		"\n" +
		"У сайтов может быть несколько статусов:\n" +
		"0 - никогда не проверялся (ждем проверки)\n" +
		"1 - ошибка подключения \n" +
		"2 - истекает сертификат \n" +
		"200 - ОК-статус" +
		"все остальные http-коды считаются некорректными"
)

func init() {
	SiteList = make(map[string]int)
	flag.StringVar(&configFile, "config", "config.json", "config file")

}

func send_notif(bot *tgbotapi.BotAPI) {
	for site, status := range SiteList {
		if status != 200 {
			alarm := fmt.Sprintf("CRIT - %s ; status: %v", site, status)
			bot.Send(tgbotapi.NewMessage(chatID, alarm))
		}
	}

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

func load_list() {
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Printf("No such file - starting without config: %s", err)
		return
	}

	if err = json.Unmarshal(data, &SiteList); err != nil {
		log.Printf("Can't read file - starting without config: %s", err)
		return
	}

	log.Println(string(data))
}

func main() {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
}
