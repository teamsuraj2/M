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

// Serve JSON response
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func startAPIServer() {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	// ================= BIOMODE =================
	http.HandleFunc("/api/biomode", func(w http.ResponseWriter, r *http.Request) {
		chatIDRaw := r.URL.Query().Get("chat_id")
		if chatIDRaw == "" {
			http.Error(w, "chat_id required", http.StatusBadRequest)
			return
		}

		chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
		if err != nil {
			http.Error(w, "invalid chat_id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			enabled, err := database.GetBioMode(chatID)
			if err != nil {
				http.Error(w, "failed to check biomode", http.StatusInternalServerError)
				return
			}
			writeJSON(w, enabled)

		case http.MethodPost:
			var enabled bool
			if err := json.NewDecoder(r.Body).Decode(&enabled); err != nil {
				http.Error(w, "Invalid JSON: expected true/false", http.StatusBadRequest)
				return
			}

			if enabled {
				err = database.SetBioMode(chatID)
			} else {
				err = database.DelBioMode(chatID)
			}
			if err != nil {
				http.Error(w, "failed to update biomode", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]string{"status": "ok"})

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// ================= ECHO SETTINGS =================
	http.HandleFunc("/api/echo", func(w http.ResponseWriter, r *http.Request) {
		chatIDRaw := r.URL.Query().Get("chat_id")
		if chatIDRaw == "" {
			http.Error(w, "chat_id required", http.StatusBadRequest)
			return
		}

		chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
		if err != nil {
			http.Error(w, "invalid chat_id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			settings, err := database.GetEchoSettings(chatID)
			if err != nil {
				http.Error(w, "failed to get echo settings", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]interface{}{
				"long_mode":  strings.ToLower(settings.Mode),
				"long_limit": settings.Limit,
			})

		case http.MethodPost:
			var body struct {
				LongMode  string `json:"long_mode"`
				LongLimit int    `json:"long_limit"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if body.LongLimit < 200 || body.LongLimit > 4000 {
				http.Error(w, "Invalid long_limit (200 - 4000)", http.StatusBadRequest)
				return
			}
			settings := &database.EchoSettings{
				ChatID: chatID,
				Mode:   body.LongMode,
				Limit:  body.LongLimit,
			}
			if err := database.SetEchoSettings(settings); err != nil {
				http.Error(w, "failed to save echo settings", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]string{"status": "ok"})

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// ================= LINKFILTER =================
	http.HandleFunc("/api/linkfilter", func(w http.ResponseWriter, r *http.Request) {
		chatIDRaw := r.URL.Query().Get("chat_id")
		if chatIDRaw == "" {
			http.Error(w, "chat_id required", http.StatusBadRequest)
			return
		}
		chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
		if err != nil {
			http.Error(w, "invalid chat_id", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			enabled, err := database.IsLinkFilterEnabled(chatID)
			if err != nil {
				http.Error(w, "failed to get enabled state", http.StatusInternalServerError)
				return
			}
			hosts, err := database.GetAllowedHostnames(chatID)
			if err != nil {
				http.Error(w, "failed to get hostnames", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]interface{}{
				"enabled":         enabled,
				"allowed_domains": hosts,
			})

		case http.MethodPost:
			var body struct {
				Enabled bool `json:"enabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}
			if err := database.SetLinkFilterEnabled(chatID, body.Enabled); err != nil {
				http.Error(w, "failed to update filter state", http.StatusInternalServerError)
				return
			}
			writeJSON(w, map[string]string{"status": "ok"})

		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// === /allow (add domain)
	http.HandleFunc("/api/linkfilter/allow", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		chatIDRaw := r.URL.Query().Get("chat_id")
		if chatIDRaw == "" {
			http.Error(w, "chat_id required", http.StatusBadRequest)
			return
		}
		chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
		if err != nil {
			http.Error(w, "invalid chat_id", http.StatusBadRequest)
			return
		}
		var body struct {
			Domain string `json:"domain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Domain) == "" {
			http.Error(w, "invalid domain", http.StatusBadRequest)
			return
		}
		if err := database.AddAllowedHostname(chatID, strings.ToLower(strings.TrimSpace(body.Domain))); err != nil {
			http.Error(w, "failed to add domain", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "ok"})
	})

	// === /remove (remove domain)
	http.HandleFunc("/api/linkfilter/remove", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		chatIDRaw := r.URL.Query().Get("chat_id")
		if chatIDRaw == "" {
			http.Error(w, "chat_id required", http.StatusBadRequest)
			return
		}
		chatID, err := strconv.ParseInt(chatIDRaw, 10, 64)
		if err != nil {
			http.Error(w, "invalid chat_id", http.StatusBadRequest)
			return
		}
		var body struct {
			Domain string `json:"domain"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Domain) == "" {
			http.Error(w, "invalid domain", http.StatusBadRequest)
			return
		}
		if err := database.RemoveAllowedHostname(chatID, strings.ToLower(strings.TrimSpace(body.Domain))); err != nil {
			http.Error(w, "failed to remove domain", http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]string{"status": "ok"})
	})
}

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
