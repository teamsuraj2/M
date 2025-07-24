package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/database"
	"main/modules"
)

func main() {
	defer database.Disconnect()

	if err := database.MigrateUsers(); err != nil {
		log.Panicf("MigrateUsers Error: %v", err)
	}
	if err := database.MigrateChats(); err != nil {
		log.Panicf("MigrateChats Error: %v", err)
	}

	client, err := telegram.NewClient(telegram.ClientConfig{
		AppID:     config.ApiId,
		AppHash:   config.ApiHash,
		LogLevel:  telegram.LogError,
		ParseMode: "HTML",
	})
	if err != nil {
		log.Panic(err)
	}

	if err := client.LoginBot(config.Token); err != nil {
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			fmt.Println("❌ Bot token has been revoked or expired.")
			os.Exit(1)
		}
		log.Panic(err)
	}

	modules.BotInfo, err = client.GetMe()
	if err != nil {
		client.SendMessage(config.LoggerId, "Failed to getMe: "+err.Error())
	}

	modules.LoadMods(client)

	startAPIServer(client)
if !pingApi() {
		config.PrintAndExit("Unable to connect to API. Make sure your backend is running and accessible.")
	}

	if modules.BotInfo.Username != "" && !strings.Contains(config.WebAppUrl, modules.BotInfo.Username) {
		config.PrintAndExit("WEB_APP_URL is filled incorrectly. It should contain your bot's username.\nExample: http://t.me/ViyomBot/settings")
	}

	client.SendMessage(config.LoggerId, "Started...")
	log.Println("✅ Bot Started")
	client.Idle()
}

func pingApi() bool {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(config.WebAppUrl + "/ping")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var res struct {
		Success bool `json:"success"`
	}
	return json.NewDecoder(resp.Body).Decode(&res) == nil && res.Success
}
