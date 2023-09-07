package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

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

	if telegramBotToken == "" {
		log.Print("-telegrambottoken is required")
		os.Exit(1)
	}

	if chatID == 0 {
		log.Print("-chatid is required")
		os.Exit(1)
	}

	load_list()

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

func monitor(bot *tgbotapi.BotAPI) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	var httpClient = &http.Client{
		Timeout:   time.Second * 10,
		Transport: tr,
	}

	for {
		save_list()

		for site, _ := range SiteList {
			response, err := httpClient.Get(site)
			if err != nil {
				SiteList[site] = 1
				log.Printf("Status of %s: %s: %s", site, "1 - Connection error", err)
			} else {
				log.Printf("Status of %s: %s", site, response.Status)
				SiteList[site] = response.StatusCode

				siteUrl, err := url.Parse(site)
				if err != nil {
					panic(err)
				}

				if siteUrl.Scheme == "https" {
					conn, err := tls.Dial("tcp", siteUrl.Host+":443", tr.TLSClientConfig)
					if err != nil {
						log.Printf("Error in SSL dial to %s: %s", siteUrl.Host, err)
					}

					certs := conn.ConnectionState().PeerCertificates

					for _, cert := range certs {
						difference := time.Since(cert.NotAfter)
						dayToExspire := int64(difference.Hours() / 24)
						if dayToExspire > -(sslDaysToExipireAlert) {
							log.Printf("Status of %s: %s", site, "2 - certificate is expiring")
							SiteList[site] = 2
						}
					}

					conn.Close()
				}
			}
		}

		send_notif(bot)
		time.Sleep(time.Minute * 5)
	}
}

func main() {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

}
