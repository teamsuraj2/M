package main

import (
	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/database"
	"main/modules"
)

func main() {
	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:     config.ApiId,
		AppHash:   config.ApiHash,
		LogLevel:  telegram.LogInfo,
		ParseMode: "HTML",
	})
	if err != nil {
		log.Panic(err)
	}
	client.LoginBot(config.Token)
	defer database.Disconnect()

	modules.LoadMods(client)
	client.SendMessage(config.LoggerId, "Started...")
	client.Idle()
}
