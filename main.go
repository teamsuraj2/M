package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/amarnathcjd/gogram/telegram"

	"main/config"
	"main/database"
	"main/modules"
)

type BioModeSettings struct {
	Enabled bool `json:"enabled"`
}

type EchoSettings struct {
	EchoText  string `json:"echo_text"`
	LongMode  string `json:"long_mode"`
	LongLimit int    `json:"long_limit"`
}

type LinkFilterSettings struct {
	Enabled        bool     `json:"enabled"`
	AllowedDomains []string `json:"allowed_domains"`
}

var (
	bioModeConfig    = BioModeSettings{Enabled: false}
	echoConfig       = EchoSettings{EchoText: "", LongMode: "automatic", LongLimit: 800}
	linkFilterConfig = LinkFilterSettings{
		Enabled:        false,
		AllowedDomains: []string{"example.com"},
	}
)

func startAPIServer() {
    http.Handle("/", http.FileServer(http.Dir("./static")))

    http.HandleFunc("/api/biomode", func(w http.ResponseWriter, r *http.Request) {
        chatID := r.URL.Query().Get("chat_id")
        if chatID == "" {
            http.Error(w, "chat_id required", http.StatusBadRequest)
            return
        }

        switch r.Method {
        case http.MethodGet:
            cfg := database.GetBioMode(chatID)
            writeJSON(w, cfg)
        case http.MethodPost:
            var newCfg database.BioModeSettings
            if err := json.NewDecoder(r.Body).Decode(&newCfg); err != nil {
                http.Error(w, "Invalid JSON", http.StatusBadRequest)
                return
            }
            database.SaveBioMode(chatID, newCfg)
            writeJSON(w, map[string]string{"status": "ok"})
        }
    })

    http.HandleFunc("/api/echo", func(w http.ResponseWriter, r *http.Request) {
        chatID := r.URL.Query().Get("chat_id")
        if chatID == "" {
            http.Error(w, "chat_id required", http.StatusBadRequest)
            return
        }

        switch r.Method {
        case http.MethodGet:
            cfg := database.GetEcho(chatID)
            writeJSON(w, cfg)
        case http.MethodPost:
            var newCfg database.EchoSettings
            json.NewDecoder(r.Body).Decode(&newCfg)
            database.SaveEcho(chatID, newCfg)
            writeJSON(w, map[string]string{"status": "ok"})
        }
    })

    http.HandleFunc("/api/linkfilter", func(w http.ResponseWriter, r *http.Request) {
        chatID := r.URL.Query().Get("chat_id")
        if chatID == "" {
            http.Error(w, "chat_id required", http.StatusBadRequest)
            return
        }

        switch r.Method {
        case http.MethodGet:
            cfg := database.GetLinkConfig(chatID)
            writeJSON(w, cfg)
        case http.MethodPost:
            var newCfg database.LinkFilterSettings
            json.NewDecoder(r.Body).Decode(&newCfg)
            database.SaveLinkFilter(chatID, newCfg)
            writeJSON(w, map[string]string{"status": "ok"})
        }
    })
}

func writeJSON(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(data)
}


func main() {
	defer database.Disconnect()

	dbErr := database.MigrateUsers()
	if dbErr != nil {
		log.Panic(fmt.Sprintf("MigrateUsers Error: %v", dbErr))
	}

	dbErr = database.MigrateChats()
	if dbErr != nil {
		log.Panic(fmt.Sprintf("MigrateChats Error: %v", dbErr))
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
	err = client.LoginBot(config.Token)
	if err != nil {
		if strings.Contains(err.Error(), "ACCESS_TOKEN_EXPIRED") {
			fmt.Println("❌ Bot token has been revoked or expired.")
			os.Exit(1)
		}
		log.Panic(err)
	}

	modules.BotInfo, err = client.GetMe()
	if err != nil {
		client.SendMessage(config.LoggerId, "Failed to getme: "+err.Error())
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
	log.Println("Started...")
	client.Idle()
}
