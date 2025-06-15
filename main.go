package main

import (
	"log"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/database"
	"main/modules"
)

func main() {
	defer database.Disconnect()
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:     config.ApiId,
		AppHash:   config.ApiHash,
		LogLevel:  telegram.LogInfo,
		ParseMode: "HTML",
	})
	if err != nil {
		log.Panic(err)
	}
	err = client.LoginBot(config.Token)
	if err != nil {
		log.Panic(err)
	}

	modules.LoadMods(client)
	client.SendMessage(config.LoggerId, "Started...")
	log.Println("Started...")
	client.Idle()
}
