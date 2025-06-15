package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			fmt.Println("❌ Bot token has been revoked or expired.")
			os.Exit(1)
		}
		log.Panic(err)
	}

	modules.LoadMods(client)
	client.SendMessage(config.LoggerId, "Started...")
	log.Println("Started...")
	client.Idle()
}
