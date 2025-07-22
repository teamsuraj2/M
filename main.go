package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

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

	startAPIServer()
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("🌐 Web UI: http://localhost:%s\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("API server error: %v", err)
		}
	}()

	client.SendMessage(config.LoggerId, "Started...")
	log.Println("✅ Bot Started")
	client.Idle()
}
